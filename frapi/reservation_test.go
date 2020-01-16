package frapi

import (
	"context"
	"testing"

	"github.com/bjorge/friendlyreservations/models"
)

// BUG(bjorge): test non member reservation and test rates and test some constraint failures

func TestReservationResolverSmokeTest(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	// get initial ledger records
	checkLedger(ctx, t, property, userID, 1, startLedgerEvent, 0, 0)

	t.Log("create a single reservation")
	property, reservations := createReservation(ctx, t, resolver, property, userID, today.AddDays(1).ToString(), today.AddDays(2).ToString())

	if len(reservations) != 1 {
		t.Fatalf("expected 1 reservation")
	}

	t.Log("create a second reservation")
	property, reservations = createReservation(ctx, t, resolver, property, userID, today.AddDays(4).ToString(), today.AddDays(5).ToString())

	if len(reservations) != 2 {
		t.Fatalf("expected 2 reservations")
	}

	for _, reservation := range reservations {
		t.Logf("start date: %+v", reservation.StartDate())
	}

	t.Log("create a reservation with a different user")

	secondUserEmail := "newone@a.out"
	property, newUser := createUser(ctx, t, resolver, property, secondUserEmail, "heythere")
	//utilities.SetTestUser(secondUserEmail)
	testUserEmail = secondUserEmail

	property, reservations = createReservation(ctx, t, resolver, property, newUser.UserID(), today.AddDays(6).ToString(), today.AddDays(7).ToString())

	if len(reservations) != 3 {
		t.Fatalf("expected 3 reservations")
	}

	t.Log("retrieve all reservations for only a single user")
	newUserID := newUser.UserID()
	reservations, err := property.Reservations(&reservationsArgs{UserID: &newUserID})
	if err != nil {
		t.Fatal(err)
	}
	if len(reservations) != 1 {
		t.Fatalf("expected 1 reservation but got %+v", len(reservations))
	}
}

func TestReservationResolverCancel(t *testing.T) {
	property, ctx, resolver, me, todayDate := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	checkin := todayDate.AddDays(1)
	checkout := checkin.AddDays(2)

	t.Log("create a single reservation")
	property, reservations := createReservation(ctx, t, resolver, property, userID, checkin.ToString(), checkout.ToString())

	if reservations[0].Canceled() {
		t.Fatalf("should not be canceled")
	}

	property, reservation := cancelReservation(ctx, t, resolver, property, reservations[0].ReservationID(), false, property.EventVersion())
	if !reservation.Canceled() {
		t.Fatalf("should be canceled")
	}

}

func TestCalendarDaysLowBalance(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	t.Log("TestCalendarDaysLowBalance")

	checkinDate := today.AddDays(1)
	checkoutDate := checkinDate.AddDays(100)
	//numDays := int32(4)
	t.Log("TestCalendarDaysLowBalance first should be ok")
	property, _ = createReservation(ctx, t, resolver, property, me.UserID(), checkinDate.ToString(), checkoutDate.ToString())

	t.Log("TestCalendarDaysLowBalance next bad balance")
	checkinDate = checkoutDate.AddDays(1)
	checkoutDate = checkinDate.AddDays(100)

	// Create reservation with low balance
	newReservationInput := &models.NewReservationInput{
		// Fields received from the client
		ReservedForUserId: me.UserID(),
		StartDate:         checkinDate.ToString(),
		EndDate:           checkoutDate.ToString(),
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

	if err == nil {
		t.Fatalf("Expected an error for the balance to be too low")
	}

}

func TestBlacklistConstraints(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	t.Log("TestBlacklistConstraints")

	property, _, _ = createBlackoutRestriction(ctx, t, resolver, property, today)

	userID := me.UserID()
	constraints, err := property.NewReservationConstraints(ctx, &NewReservationConstraintsArgs{UserID: &userID, UserType: MEMBER})
	if err != nil {
		t.Fatal(err)
	}

	if len(constraints.CheckinDisabled()) == 0 {
		t.Fatalf("Expected a blackout constraint")
	}

}
