package frapi

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/local_platform"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// initAndCreateTestProperty is used by various unit tests as a starting point
func initAndCreateTestProperty(testCtx context.Context, t *testing.T) (*PropertyResolver, context.Context, *Resolver, *UserResolver, *frdate.Date) {
	testUserEmail = defaultEmail
	utilities.SetTestSystemUser("noreply@testing.com")
	utilities.SetAllowCreateProperty()
	utilities.TrialDuration, _ = time.ParseDuration("0h")

	frdate.TestTimeOffsetDays = nil
	PersistedEmailStore = localplatform.NewPersistedEmailStore()
	PersistedVersionedEvents = localplatform.NewPersistedVersionedEvents()
	PersistedPropertyList = localplatform.NewPersistedPropertyList()
	EmailSender = localplatform.NewEmailSender()

	resolver := &Resolver{}

	propertyResolver, err := resolver.CreateProperty(testCtx, &struct{ Input *models.NewPropertyInput }{Input: defaultPropertyInput})
	if err != nil {
		t.Fatal(err)
	}

	me, err := propertyResolver.Me()
	if err != nil {
		t.Fatal(err)
	}

	dateBuilder, err := frdate.NewDateBuilder("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	today := dateBuilder.Today()

	return propertyResolver, testCtx, resolver, me, today
}

func updateMembership(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, restrictionID string,
	userID string,
	purchase bool,
	comment string) *PropertyResolver {

	input := &models.UpdateMembershipStatusInput{}
	input.AdminUpdate = false
	input.Comment = &comment
	input.ForVersion = property.EventVersion()
	input.Purchase = purchase
	input.RestrictionId = restrictionID
	input.UpdateForUserId = userID

	property, err := resolver.UpdateMembershipStatus(ctx, &struct {
		PropertyID string
		Input      *models.UpdateMembershipStatusInput
	}{
		PropertyID: property.PropertyID(),
		Input:      input,
	})

	if err != nil {
		t.Fatal(err)
	}

	return property
}

func cancelReservation(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, reservationID string, adminRequest bool, forVersion int32) (*PropertyResolver, *ReservationResolver) {

	property, err := resolver.CancelReservation(ctx, &struct {
		PropertyID    string
		ForVersion    int32
		ReservationID string
		AdminRequest  *bool
	}{
		PropertyID:    property.PropertyID(),
		ForVersion:    forVersion,
		ReservationID: reservationID,
		AdminRequest:  &adminRequest,
	})

	if err != nil {
		t.Fatal(err)
	}

	reservations, err := property.Reservations(&reservationsArgs{ReservationID: &reservationID})

	if err != nil {
		t.Fatal(err)
	}

	if len(reservations) != 1 {
		t.Fatalf("reservation not found")
	}

	return property, reservations[0]
}

func createReservation(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, userID string, startDate string, endDate string) (*PropertyResolver, []*ReservationResolver) {
	// Create reservation
	newReservationInput := &models.NewReservationInput{
		// Fields received from the client
		ForVersion:        property.EventVersion(),
		ReservedForUserId: userID,
		StartDate:         startDate,
		EndDate:           endDate,
		Member:            true,
	}

	property, err := resolver.CreateReservation(ctx,
		&struct {
			PropertyID string
			Input      *models.NewReservationInput
		}{
			PropertyID: property.PropertyID(),
			Input:      newReservationInput,
		})

	if err != nil {
		t.Fatal(err)
	}

	reservations, err := property.Reservations(&reservationsArgs{Order: DESCENDING})

	if err != nil {
		t.Fatal(err)
	}

	return property, reservations
}

// checkLedger checks the ledger amounts and the last record for further testing
func checkLedger(ctx context.Context, t *testing.T, property *PropertyResolver, userID string, numRecords int, event LedgerEvent, balance int32, amount int32) (*LedgerRecordResolver, []*LedgerRecordResolver) {
	userRecords, _ := property.Ledgers(&ledgersArgs{UserID: &userID})

	records := userRecords[0].Records()

	// records, _ := property.Ledgers(&struct {
	// 	UserId  string
	// 	Last    *int32
	// 	Reverse *bool
	// }{UserId: userId})

	for _, item := range records {
		t.Logf("record: %+v", *item)
	}

	if len(records) != numRecords {
		t.Fatal(fmt.Errorf("TestLedgerResolver: wrong number or records %+v, expected %+v", len(records), numRecords))
	}

	record := records[numRecords-1]

	if record.Event() != event {
		t.Fatal(fmt.Errorf("TestLedgerResolver: wrong event %+v, expected %+v", record.Event(), event))
	}
	if record.amountInternal().NoDecimal() != strconv.Itoa(int(amount)) {
		t.Fatal(fmt.Errorf("TestLedgerResolver: wrong amount %+v, expected %+v", record.amountInternal().NoDecimal(), amount))
	}
	if record.balanceInternal().NoDecimal() != strconv.Itoa(int(balance)) {
		t.Fatal(fmt.Errorf("TestLedgerResolver: wrong balance %+v, expected %+v", record.balanceInternal(), balance))
	}

	return record, records
}

func createUser(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, email string, nickname string) (*PropertyResolver, *UserResolver) {
	newUserInput := &models.NewUserInput{
		Email:      email,
		Nickname:   nickname,
		ForVersion: property.EventVersion(),
	}
	property, err := resolver.CreateUser(ctx, &struct {
		PropertyID string
		Input      *models.NewUserInput
	}{
		PropertyID: property.PropertyID(),
		Input:      newUserInput,
	})

	if err != nil {
		t.Fatal(err)
	}

	users := property.Users(&usersArgs{Email: &email})
	//userResolver, err := property.GetUserByEmailHelper(ctx, email)

	if len(users) != 1 {
		t.Fatalf("expected a user")
	}

	return property, users[0]
}

func createPayment(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, paymentAmount int32, increase bool, forVersion int32) *PropertyResolver {
	//paymentAmount := int32(200)
	me, _ := property.Me()
	payment := &models.UpdateBalanceInput{
		UpdateForUserId: me.UserID(),
		Amount:          paymentAmount,
		Description:     "test payment",
		Increase:        increase,
		ForVersion:      forVersion,
	}

	property, err := resolver.UpdateBalance(ctx, &struct {
		PropertyID string
		Input      *models.UpdateBalanceInput
	}{
		PropertyID: property.PropertyID(),
		Input:      payment,
	})

	if err != nil {
		t.Fatal(err)
	}

	return property
}

func createMembershipRestriction(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, today *frdate.Date) (*PropertyResolver, *models.MembershipRestriction, []*RestrictionRecordResolver) {
	//nextYear := today.AddDays(365)
	first, last := today.YearInOut()
	t.Logf("create membership: in %+v out %+v", first.ToString(), last.ToString())
	membershipRestriction := &models.MembershipRestriction{}
	membershipRestriction.Amount = 30000
	membershipRestriction.PrePayStartDate = first.AddDays(-30).ToString()
	membershipRestriction.InDate = first.ToString()
	membershipRestriction.OutDate = last.ToString()
	membershipRestriction.GracePeriodOutDate = last.AddDays(300).ToString()

	newRestrictionInput := &models.NewRestrictionInput{}
	newRestrictionInput.ForVersion = property.EventVersion()
	newRestrictionInput.Membership = membershipRestriction
	newRestrictionInput.Description = strconv.Itoa(today.Year())

	property, err := resolver.CreateRestriction(ctx, &struct {
		PropertyID string
		Input      *models.NewRestrictionInput
	}{
		PropertyID: property.PropertyID(),
		Input:      newRestrictionInput,
	})

	if err != nil {
		t.Fatal(err)
	}

	restrictions, err := property.Restrictions(&restrictionsArgs{})

	if err != nil {
		t.Fatal(err)
	}

	return property, membershipRestriction, restrictions
}

func createBlackoutRestriction(ctx context.Context, t *testing.T, resolver *Resolver, property *PropertyResolver, today *frdate.Date) (*PropertyResolver, *models.BlackoutRestriction, []*RestrictionRecordResolver) {
	//nextYear := today.AddDays(365)
	first, last := today.YearInOut()
	blackoutRestriction := &models.BlackoutRestriction{}
	blackoutRestriction.StartDate = first.ToString()
	blackoutRestriction.EndDate = last.ToString()

	newRestrictionInput := &models.NewRestrictionInput{}
	newRestrictionInput.ForVersion = property.EventVersion()
	newRestrictionInput.Blackout = blackoutRestriction
	newRestrictionInput.Description = strconv.Itoa(today.Year())

	property, err := resolver.CreateRestriction(ctx, &struct {
		PropertyID string
		Input      *models.NewRestrictionInput
	}{
		PropertyID: property.PropertyID(),
		Input:      newRestrictionInput,
	})

	if err != nil {
		t.Fatal(err)
	}

	restrictions, err := property.Restrictions(&restrictionsArgs{})

	if err != nil {
		t.Fatal(err)
	}

	return property, blackoutRestriction, restrictions
}
