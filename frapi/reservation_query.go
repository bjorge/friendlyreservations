package frapi

import (
	"fmt"
	"sort"

	"github.com/bjorge/friendlyreservations/frdate"

	"github.com/bjorge/friendlyreservations/models"
)

// BUG(bjorge): consider having separate create author and cancel author
const reservationGQL = `
type Reservation {
	reservationId: String!
	createDateTime: String!
	reservedFor: User!
	author: User!
	startDate: String!
	endDate: String!
	member: Boolean!
	nonMemberName: String
	nonMemberInfo: String
	rate: [DailyRate!]!
	amount: Int!
	canceled: Boolean!
}

type DailyRate {
	date: String!
	amount: Int!
}

enum OrderDirection {
	ASCENDING
	DESCENDING
  }

`

// OrderDirection is the sorting order for the reservations by checkin date
type OrderDirection string

const (
	// ASCENDING is oldest to newest reservations by checkin date
	ASCENDING OrderDirection = "ASCENDING"
	// DESCENDING is newest to oldest reservations by checkin date
	DESCENDING OrderDirection = "DESCENDING"
)

type reservationsArgs struct {
	ReservationID *string
	MaxVersion    *int32
	UserID        *string
	Order         OrderDirection
}

// DailyRateResolver resolves a reservation daily rate
type DailyRateResolver struct {
	dailyRate *models.DailyRate
}

// Date is the date of the reservation
func (r *DailyRateResolver) Date() string {
	return r.dailyRate.Date
}

// Amount is the cost of the reservation for the date
func (r *DailyRateResolver) Amount() int32 {
	return r.dailyRate.Amount
}

// Reservations is the gql call to return the list of reservations
func (r *PropertyResolver) Reservations(args *reservationsArgs) ([]*ReservationResolver, error) {

	// validate input for query
	if args.MaxVersion != nil && *args.MaxVersion <= 0 {
		return nil, fmt.Errorf("max version arg must be greater than 0")
	}

	if args.Order == "" {
		args.Order = ASCENDING
	}

	r.rollupReservations()

	// get rollups (with common filters applied)
	var l []*ReservationResolver
	ifaces := r.getRollups(&rollupArgs{id: args.ReservationID, maxVersion: args.MaxVersion}, reservationRollupType)
	for _, iface := range ifaces {
		resolver := &ReservationResolver{}
		resolver.property = r
		resolver.args = args
		resolver.rollup = iface.(*ReservationRollup)
		l = append(l, resolver)
	}

	// additional userid filter
	if args.UserID != nil {
		var single []*ReservationResolver
		for _, reservation := range l {
			if reservation.ReservedFor().UserID() != *args.UserID {
				continue
			}
			single = append(single, reservation)
		}
		l = single
	}

	// ordering
	settings, err := r.Settings(&settingsArgs{MaxVersion: args.MaxVersion})
	if err != nil {
		return nil, err
	}
	sort.Slice(l, func(i, j int) bool {
		firstStart := frdate.MustNewDate(settings.Timezone(), l[i].StartDate())
		secondStart := frdate.MustNewDate(settings.Timezone(), l[j].StartDate())
		if args.Order == ASCENDING {
			return firstStart.Before(secondStart)
		}
		return secondStart.Before(firstStart)
	})

	return l, nil
}

// ReservationResolver is the reservation resolver receiver
type ReservationResolver struct {
	rollup   *ReservationRollup
	property *PropertyResolver
	args     *reservationsArgs
}

// ReservationID returns the reservation id
func (r *ReservationResolver) ReservationID() string {
	return r.rollup.Input.ReservationId
}

// CreateDateTime returns the reservation create time stamp
func (r *ReservationResolver) CreateDateTime() string {
	return r.rollup.Input.CreateDateTime
}

// UpdateDateTime returns the time stamp for an update to the reservation (ex. cancel)
func (r *ReservationResolver) UpdateDateTime() string {
	return r.rollup.UpdateDateTime
}

// ReservedFor denotes who will reserve the reservation
func (r *ReservationResolver) ReservedFor() *UserResolver {
	userResolvers := r.property.Users(&usersArgs{
		UserID:     &r.rollup.Input.ReservedForUserId,
		MaxVersion: r.args.MaxVersion,
	})
	if len(userResolvers) > 0 {
		return userResolvers[0]
	}
	return nil
}

// Author denotes who made the reservation (same as ReservedFor or an admin)
func (r *ReservationResolver) Author() *UserResolver {
	userResolvers := r.property.Users(&usersArgs{
		UserID:     &r.rollup.Input.AuthorUserId,
		MaxVersion: r.args.MaxVersion,
	})
	if len(userResolvers) > 0 {
		return userResolvers[0]
	}
	return nil
}

// StartDate is the checkin date
func (r *ReservationResolver) StartDate() string {
	return r.rollup.Input.StartDate
}

// EndDate is the checkout date
func (r *ReservationResolver) EndDate() string {
	return r.rollup.Input.EndDate
}

// Member is true if the reservation is for the ReservedFor member,
// otherwise the reservation is for a non-member
func (r *ReservationResolver) Member() bool {
	return r.rollup.Input.Member
}

// NonMemberName is required if the Member flag is false
func (r *ReservationResolver) NonMemberName() *string {
	return r.rollup.Input.NonMemberName
}

// NonMemberInfo is required if the Member flag is false
func (r *ReservationResolver) NonMemberInfo() *string {
	return r.rollup.Input.NonMemberInfo
}

// Canceled is true if the reservation has been canceled
func (r *ReservationResolver) Canceled() bool {
	return r.rollup.Canceled
}

// Rate is a list of the daily price for the reservation
func (r *ReservationResolver) Rate() []*DailyRateResolver {
	var l []*DailyRateResolver
	for _, item := range r.rollup.Input.Rate {
		dailyRate := item
		l = append(l, &DailyRateResolver{&dailyRate})
	}

	return l

}

// Amount is the sum of the daily Rates
func (r *ReservationResolver) Amount() int32 {
	var amount int32
	for _, item := range r.rollup.Input.Rate {
		amount += item.Amount
	}
	return amount
}
