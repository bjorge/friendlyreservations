package frapi

import (
	"fmt"
	"sort"

	"github.com/bjorge/friendlyreservations/frdate"
)

const membershipStatusConstraintsGQL = `

enum MembershipState {
	# membership has not been purchased or set to opt out
	OPEN
	# membership has been purchased
	PURCHASED
	# user has opted out
	OPTOUT
}

type MembershipStatusConstraintsRecord {
	status: MembershipState!
	info: RestrictionRecord!
	reservationCount: Int!
	# cannot opt out if reservations exist
	optOutAllowed: Boolean!
	# cannot purchase if over minimum balance
	purchaseAllowed: Boolean!
}

type MembershipStatusConstraints {
	user: User!
	memberships: [MembershipStatusConstraintsRecord]!
}

`

type membershipStatusConstraintsArgs struct {
	UserID     *string
	MaxVersion *int32
}

// MembershipStatusConstraints is called by the GQL framewor, see contentGQL
func (r *PropertyResolver) MembershipStatusConstraints(args *membershipStatusConstraintsArgs) ([]*MembershipStatusConstraintsResolver, error) {

	// validate input
	if args.UserID != nil {
		users := r.Users(&usersArgs{UserID: args.UserID})
		if len(users) != 1 {
			return nil, fmt.Errorf("user for id %+v not found", *args.UserID)
		}
	}

	if args.MaxVersion != nil {
		if *args.MaxVersion < 0 {
			return nil, fmt.Errorf("maxVersion cannot be less than 0")
		}
	}

	// rollup
	r.rollupMembershipStatus()

	// resolve
	return r.resolveMembershipStatus(args)
}

// UserMembershipRecord contains the information for resolving a user membership
type UserMembershipRecord struct {
	restrictionID    string
	userID           string
	status           MembershipState
	inDate           *frdate.Date
	outDate          *frdate.Date
	prePayStartDate  *frdate.Date
	purchaseAllowed  *bool
	optOutAllowed    *bool
	reservationCount int
}

// MembershipRecordResolver resolves one of the memberships for a user
type MembershipRecordResolver struct {
	membership *UserMembershipRecord
	property   *PropertyResolver
	userID     string
}

// MembershipStatusConstraintsResolver resolves the memberships for a user
type MembershipStatusConstraintsResolver struct {
	userID         string
	allMemberShips []*UserMembershipRecord
	property       *PropertyResolver
}

// Memberships is called by the GQL framework, see membershipStatusGQL
func (r *MembershipStatusConstraintsResolver) Memberships() []*MembershipRecordResolver {

	resolvers := []*MembershipRecordResolver{}
	for _, membership := range r.allMemberShips {
		resolver := &MembershipRecordResolver{}
		resolver.membership = membership
		resolver.property = r.property
		resolver.userID = r.userID
		resolvers = append(resolvers, resolver)
	}

	// descending order so that newest on the top of the list
	sort.Slice(resolvers, func(i, j int) bool {
		return resolvers[i].membership.inDate.After(resolvers[j].membership.inDate)
	})
	return resolvers
}

// User is called by the GQL framework, see membershipStatusGQL
func (r *MembershipStatusConstraintsResolver) User() *UserResolver {
	users := r.property.Users(&usersArgs{UserID: &r.userID})
	return users[0]
}

// Info is called by the GQL framework, see membershipStatusGQL
func (r *MembershipRecordResolver) Info() *RestrictionRecordResolver {
	restrictions, _ := r.property.Restrictions(&restrictionsArgs{RestrictionID: &r.membership.restrictionID})
	return restrictions[0]
}

// Status is called by the GQL framework, see membershipStatusGQL
func (r *MembershipRecordResolver) Status() MembershipState {
	return r.membership.status
}

// ReservationCount is called by the GQL framework, see membershipStatusGQL
func (r *MembershipRecordResolver) ReservationCount() int32 {
	return int32(r.membership.reservationCount)
}

var bFalse = false
var bTrue = true

// OptOutAllowed is called by the GQL framework, see membershipStatusGQL
func (r *MembershipRecordResolver) OptOutAllowed() bool {
	return *r.membership.optOutAllowed
}
func setOptOutAllowed(allMemberShips []*UserMembershipRecord) {

	foundOpen := false
	for _, resolver := range allMemberShips {
		if foundOpen {
			// don't allow updates after first OPEN record
			resolver.optOutAllowed = &bFalse
		} else {
			switch resolver.status {
			case OPEN:
				if !foundOpen {
					foundOpen = true
					resolver.optOutAllowed = &bTrue
				}
			case OPTOUT:
				resolver.optOutAllowed = &bFalse
			case PURCHASED:
				resolver.optOutAllowed = &bTrue
			default:
				resolver.optOutAllowed = &bFalse
			}
		}
		if resolver.reservationCount > 0 {
			resolver.optOutAllowed = &bFalse
		}
	}
}

// PurchaseAllowed is called by the GQL framework, see membershipStatusGQL
func setPurchaseAllowed(lowBalance bool, allMemberShips []*UserMembershipRecord, today *frdate.Date) {

	foundOpen := false
	for _, resolver := range allMemberShips {
		if foundOpen {
			// don't allow updates after first OPEN record
			resolver.purchaseAllowed = &bFalse
		} else {
			switch resolver.status {
			case OPEN:
				resolver.purchaseAllowed = &bTrue
				if !foundOpen {
					foundOpen = true
				}
			case OPTOUT:
				resolver.purchaseAllowed = &bTrue
			case PURCHASED:
				resolver.purchaseAllowed = &bFalse
			}
		}
		// if before prepay start, then not allowed
		if today.Before(resolver.prePayStartDate) || lowBalance {
			resolver.purchaseAllowed = &bFalse
		}
	}

}

// PurchaseAllowed is called by the GQL framework, see membershipStatusGQL
func (r *MembershipRecordResolver) PurchaseAllowed() bool {
	return *r.membership.purchaseAllowed
}

func (r *PropertyResolver) resolveMembershipStatus(args *membershipStatusConstraintsArgs) ([]*MembershipStatusConstraintsResolver, error) {

	// get all the rollups
	versionedRollups := r.getRollups(&rollupArgs{maxVersion: args.MaxVersion}, membershipStatusRollupType)

	// get a datebuilder from the property timezone
	settings, err := r.Settings(&settingsArgs{})
	if err != nil {
		return nil, err
	}
	dateBuilder := frdate.MustNewDateBuilder(settings.Timezone())

	// calculate the number of reservations per membership period per user
	reservations, err := r.Reservations(&reservationsArgs{Order: ASCENDING, MaxVersion: args.MaxVersion})
	if err != nil {
		return nil, err
	}

	// userID -> membershipID -> count
	reservationCountMap := make(map[string]map[string]int)
	for _, reservation := range reservations {
		if reservation.Canceled() {
			continue
		}
		rInDate := dateBuilder.MustNewDate(reservation.StartDate())
		rOutDate := dateBuilder.MustNewDate(reservation.EndDate())
		for _, versionedRollup := range versionedRollups {
			rollup := versionedRollup.(*MembershipRollupRecord)
			mInDate := dateBuilder.MustNewDate(rollup.InDate)
			mOutDate := dateBuilder.MustNewDate(rollup.OutDate)
			if frdate.DateOverlap(rInDate, rOutDate, mInDate, mOutDate) {
				uID := reservation.ReservedFor().UserID()
				if _, ok := reservationCountMap[uID]; !ok {
					reservationCountMap[uID] = make(map[string]int)
				}
				mID := rollup.RestrictionID
				if _, ok := reservationCountMap[uID][mID]; !ok {
					reservationCountMap[uID][mID] = 0
				}
				reservationCountMap[uID][mID]++
			}
		}
	}

	// get all the users
	users := r.Users(&usersArgs{})

	resolvers := []*MembershipStatusConstraintsResolver{}

	// for each user...
	for _, user := range users {
		// return only a single user
		if args.UserID != nil {
			if user.UserID() != *args.UserID {
				continue
			}
		}

		resolver := &MembershipStatusConstraintsResolver{}
		resolver.userID = user.UserID()
		resolver.property = r

		for _, versionedRollup := range versionedRollups {
			rollup := versionedRollup.(*MembershipRollupRecord)

			record := &UserMembershipRecord{}
			record.userID = resolver.userID
			record.restrictionID = rollup.RestrictionID
			record.inDate = dateBuilder.MustNewDate(rollup.InDate)
			record.outDate = dateBuilder.MustNewDate(rollup.OutDate)
			record.prePayStartDate = dateBuilder.MustNewDate(rollup.PrePayStartDate)

			if status, ok := rollup.Users[user.UserID()]; ok {
				record.status = status
			} else {
				record.status = OPEN
			}

			// set the reservation count in the record
			if count, ok := reservationCountMap[resolver.userID][record.restrictionID]; ok {
				record.reservationCount = count
			} else {
				record.reservationCount = 0
			}

			resolver.allMemberShips = append(resolver.allMemberShips, record)
		}

		if len(resolver.allMemberShips) == 0 {
			continue
		}

		sort.Slice(resolver.allMemberShips, func(i, j int) bool {
			return resolver.allMemberShips[i].inDate.Before(resolver.allMemberShips[j].inDate)
		})

		setOptOutAllowed(resolver.allMemberShips)

		ledgers, err := r.Ledgers(&ledgersArgs{Reverse: &bTrue, UserID: &resolver.userID})
		if err != nil {
			return nil, err
		}
		balance := ledgers[0].Records()[0].balanceInternal().Raw()
		minBalance := settings.minBalanceInternal().Raw()

		setPurchaseAllowed(balance < minBalance, resolver.allMemberShips, dateBuilder.Today())

		resolvers = append(resolvers, resolver)

	}

	return resolvers, nil
}
