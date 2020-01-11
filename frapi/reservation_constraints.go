package frapi

import (
	"context"
	"fmt"
	"strconv"


	"github.com/bjorge/friendlyreservations/frdate"
)

const reservationConstraintsGQL = `

enum ConstraintsUserType {
	ADMIN
	MEMBER
	NONMEMBER
}

type CalendarDisabledRange {
	before: String
	after: String
	from: String
	to: String
}

# The reservation characteristics for a day in the calendar
type NewReservationConstraints {
	newReservationAllowed: Boolean!
	checkinDisabled: [CalendarDisabledRange]!
	checkoutDisabled: [CalendarDisabledRange]!
	nonMemberNameMin: Int!
	nonMemberNameMax: Int!
	nonMemberInfoMin: Int!
	nonMemberInfoMax: Int!
}

`

const cancelReservationConstraintsGQL = `

# List of reservations ids that can be canceled
type CancelReservationConstraints {
	cancelReservationAllowed: [String]!
}

`

// ConstraintsUserType is the user type requesting the new or cancel reservation request
type ConstraintsUserType string

const (
	// ADMIN means an admin is making the request
	ADMIN ConstraintsUserType = "ADMIN"
	// MEMBER means a member is making the request
	MEMBER ConstraintsUserType = "MEMBER"
	// NONMEMBER means a member is making the request for a non-member
	NONMEMBER ConstraintsUserType = "NONMEMBER"
)

// CalendarDisabledRange is a range of disabled dates for making reservation,
// only BeforeDate+AfterDate, or FromDate or ToDate can be set, not any together
type CalendarDisabledRange struct {
	BeforeDate *frdate.Date
	AfterDate  *frdate.Date
	FromDate   *frdate.Date
	ToDate     *frdate.Date
}

func disabledRanges(checkIn *frdate.Date, checkOut *frdate.Date) (*CalendarDisabledRange, *CalendarDisabledRange, error) {
	lastIn := checkOut.AddDays(-1)
	firstOut := checkIn.AddDays(1)
	inRange := &CalendarDisabledRange{FromDate: checkIn, ToDate: lastIn}
	outRange := &CalendarDisabledRange{FromDate: firstOut, ToDate: checkOut}

	return inRange, outRange, nil
}

// NewReservationConstraints are constraints for a new reservation
type NewReservationConstraints struct {
	newReservationAllowed bool
	checkinDisabled       []*CalendarDisabledRange
	checkoutDisabled      []*CalendarDisabledRange
}

// NewReservationConstraintsArgs are the arguments for retrieving the constraints
type NewReservationConstraintsArgs struct {
	UserID   *string
	UserType ConstraintsUserType
}

// CancelReservationConstraints are constraints for a canceling reservation
type CancelReservationConstraints struct {
	cancelAllowed []*string
}

// CancelReservationConstraintsArgs are the arguments for retrieving the constraints
type CancelReservationConstraintsArgs struct {
	UserID   *string
	UserType ConstraintsUserType
}

// CancelReservationConstraints is called to retrieve the cancel reservation constraints
func (r *PropertyResolver) CancelReservationConstraints(ctx context.Context, args *CancelReservationConstraintsArgs) (*CancelReservationConstraints, error) {

	if args.UserType != ADMIN && args.UserID == nil {
		return nil, fmt.Errorf("non-admin user type requires user id")
	}

	me, err := r.Me()
	if err != nil {
		return nil, err
	}

	if args.UserType == ADMIN && !me.IsAdmin() {
		return nil, fmt.Errorf("admin request for non-admin")
	}

	if args.UserType == NONMEMBER {
		return nil, fmt.Errorf("nonmember request not allowed")
	}

	if args.UserType == MEMBER && args.UserID == nil {
		return nil, fmt.Errorf("member request requires userId")
	}

	if args.UserID != nil {
		users := r.Users(&usersArgs{})
		found := false
		for _, user := range users {
			if user.UserID() == *args.UserID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("userId does not match a valid user")
		}

	}

	reservations, err := r.Reservations(&reservationsArgs{})
	if err != nil {
		return nil, err
	}

	settings, err := r.Settings(&settingsArgs{})
	if err != nil {
		return nil, err
	}

	dateBuilder := frdate.MustNewDateBuilder(settings.Timezone())

	// allowed cancel reservation ids
	cancelAllowed := []*string{}
	for _, reservation := range reservations {
		if !reservation.Canceled() {
			reservationID := reservation.ReservationID()
			if args.UserType == ADMIN {
				// allow admins to cancel any reservation
				cancelAllowed = append(cancelAllowed, &reservationID)
			} else {
				// members can only cancel their own future reservations
				reservationIn := dateBuilder.MustNewDate(reservation.StartDate())
				today := dateBuilder.Today()
				if !reservationIn.Before(today) && reservation.ReservedFor().UserID() == *args.UserID {
					cancelAllowed = append(cancelAllowed, &reservationID)
				}
			}
		}
	}

	return &CancelReservationConstraints{cancelAllowed}, nil
}

// CancelReservationAllowed returns the reservation ids that the caller is allowed to cancel
func (r *CancelReservationConstraints) CancelReservationAllowed() []*string {
	return r.cancelAllowed
}

// NewReservationConstraints is called to retrieve new reservation constraints
func (r *PropertyResolver) NewReservationConstraints(ctx context.Context, args *NewReservationConstraintsArgs) (*NewReservationConstraints, error) {

	if args.UserType != ADMIN && args.UserID == nil {
		return nil, fmt.Errorf("non-admin user type requires user id")
	}

	if args.UserID != nil {
		users := r.Users(&usersArgs{})
		found := false
		for _, user := range users {
			if user.UserID() == *args.UserID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("userId does not match a valid user")
		}

	}

	Logger.LogDebugf("NewReservationConstraints: start")

	newReservationConstraints := &NewReservationConstraints{}
	newReservationConstraints.newReservationAllowed = true
	newReservationConstraints.checkinDisabled = []*CalendarDisabledRange{}
	newReservationConstraints.checkoutDisabled = []*CalendarDisabledRange{}

	settings, err := r.Settings(&settingsArgs{})
	if err != nil {
		return nil, err
	}

	if args.UserType == NONMEMBER && !settings.AllowNonMembers() {
		Logger.LogDebugf("NewReservationConstraints: settings do not allow non-members")
		newReservationConstraints.newReservationAllowed = false
		return newReservationConstraints, nil
	}

	// Ledger balance check
	if args.UserType != ADMIN {
		last := int32(1)

		userRecords, _ := r.Ledgers(&ledgersArgs{UserID: args.UserID, Last: &last})

		ledgers := userRecords[0].Records()

		balance, _ := strconv.Atoi(ledgers[0].balanceInternal().NoDecimal())
		minBalance, _ := strconv.Atoi(settings.minBalanceInternal().NoDecimal())

		if balance < minBalance {
			newReservationConstraints.newReservationAllowed = false
			return newReservationConstraints, nil
		}
	}

	dateBuilder, _ := frdate.NewDateBuilder(settings.Timezone())

	// Settings check
	if args.UserType != ADMIN {
		today := dateBuilder.Today()

		// disable checkin for the past
		disableInBefore := today.AddDays(int(settings.MinInDays()))
		disabledInRange := &CalendarDisabledRange{BeforeDate: disableInBefore}
		newReservationConstraints.checkinDisabled = append(newReservationConstraints.checkinDisabled, disabledInRange)

		// once user has selected the checkin, the calendar will disable final checkout days
		disableOutAfter := today.AddDays(int(settings.MaxOutDays()))
		disabledOutRange := &CalendarDisabledRange{AfterDate: disableOutAfter}
		newReservationConstraints.checkoutDisabled = append(newReservationConstraints.checkoutDisabled, disabledOutRange)

		// last checkin is one day before last checkout
		disabledOutRange = &CalendarDisabledRange{AfterDate: disableOutAfter.AddDays(-1)}
		newReservationConstraints.checkinDisabled = append(newReservationConstraints.checkinDisabled, disabledOutRange)

	}

	reservations, err := r.Reservations(&reservationsArgs{})
	if err != nil {
		return nil, err
	}

	// Reservation ranges
	for _, reservation := range reservations {
		if !reservation.Canceled() {
			reservationIn, _ := dateBuilder.NewDate(reservation.StartDate())
			reservationOut, _ := dateBuilder.NewDate(reservation.EndDate())
			in, out, err := disabledRanges(reservationIn, reservationOut)
			if err != nil {
				return nil, err
			}

			newReservationConstraints.checkinDisabled = append(newReservationConstraints.checkinDisabled, in)
			newReservationConstraints.checkoutDisabled = append(newReservationConstraints.checkoutDisabled, out)
		}
	}

	// Blackout ranges
	if args.UserType != ADMIN {
		restrictions, err := r.Restrictions(&restrictionsArgs{})
		if err != nil {
			return nil, err
		}

		for _, restriction := range restrictions {
			Logger.LogDebugf("NewReservationConstraints: found a restriction: %+v", restriction.Description())
			blackoutRestriction, ok := restriction.Restriction().ToBlackoutRestriction()
			if ok {
				blackoutIn := dateBuilder.MustNewDate(blackoutRestriction.StartDate())
				blackoutOut := dateBuilder.MustNewDate(blackoutRestriction.EndDate())
				in, out, err := disabledRanges(blackoutIn, blackoutOut)
				if err != nil {
					return nil, err
				}

				Logger.LogDebugf("NewReservationConstraints: adding blackout restriction")

				newReservationConstraints.checkinDisabled = append(newReservationConstraints.checkinDisabled, in)
				newReservationConstraints.checkoutDisabled = append(newReservationConstraints.checkoutDisabled, out)
			}
		}
	}

	// Membership ranges
	if args.UserType != ADMIN {
		Logger.LogDebugf("NewReservationConstraints: looking for membership restrictions")
		dateBuilder, _ := frdate.NewDateBuilder(settings.Timezone())

		membershipStatusList, err := r.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: args.UserID})
		if err != nil {
			return nil, err
		}
		if len(membershipStatusList) > 0 {
			memberships := membershipStatusList[0].Memberships()

			var in *frdate.Date
			var grace *frdate.Date

			purchased := false
			for _, membership := range memberships {
				if membership.Status() == PURCHASED {
					purchased = true
					membershipRestriction, _ := membership.Info().Restriction().ToMembershipRestriction()
					if in == nil {
						in = dateBuilder.MustNewDate(membershipRestriction.inDateInternal())
					} else {
						newIn := dateBuilder.MustNewDate(membershipRestriction.inDateInternal())
						if newIn.Before(in) {
							in = newIn
						}
					}
					if grace == nil {
						grace = dateBuilder.MustNewDate(membershipRestriction.internalGracePeriodOutDate())
					} else {
						newGrace := dateBuilder.MustNewDate(membershipRestriction.internalGracePeriodOutDate())
						if newGrace.After(grace) {
							grace = newGrace
						}
					}
				}
			}

			if !purchased {
				newReservationConstraints.newReservationAllowed = false
				return newReservationConstraints, nil
			}
			inConstraint := &CalendarDisabledRange{BeforeDate: in}
			outConstraint := &CalendarDisabledRange{AfterDate: grace}
			newReservationConstraints.checkinDisabled = append(newReservationConstraints.checkinDisabled, inConstraint)
			newReservationConstraints.checkoutDisabled = append(newReservationConstraints.checkoutDisabled, outConstraint)

			// last checkin is one day before last checkout
			outConstraint = &CalendarDisabledRange{AfterDate: grace.AddDays(-1)}
			newReservationConstraints.checkinDisabled = append(newReservationConstraints.checkinDisabled, outConstraint)
		}
	}

	return newReservationConstraints, nil
}

// NewReservationAllowed is true if the caller is allowed to create a new reservation
func (r *NewReservationConstraints) NewReservationAllowed() bool {
	return r.newReservationAllowed
}

// CheckinDisabled returns the ranges of dates that cannot be a new reservation checkin date
func (r *NewReservationConstraints) CheckinDisabled() []*CalendarDisabledRange {
	return r.checkinDisabled
}

// CheckoutDisabled returns the ranges of dates that cannot be a new reservation checkout date
func (r *NewReservationConstraints) CheckoutDisabled() []*CalendarDisabledRange {
	return r.checkoutDisabled
}

// NonMemberNameMin is the minimum length of a non member name
func (r *NewReservationConstraints) NonMemberNameMin() int32 { return 3 }

// NonMemberNameMax is the maximum length of a non member name
func (r *NewReservationConstraints) NonMemberNameMax() int32 { return 25 }

// NonMemberInfoMin is the minimum length of non member information
func (r *NewReservationConstraints) NonMemberInfoMin() int32 { return 3 }

// NonMemberInfoMax is the maximum length of non member information
func (r *NewReservationConstraints) NonMemberInfoMax() int32 { return 45 }

// Before if set (not nil) allows dates that fall before this date,
// After is always set if Before is set
func (r *CalendarDisabledRange) Before() *string {
	if r.BeforeDate == nil {
		return nil
	}
	return r.BeforeDate.ToStringPtr()
}

// After if set (not nil) allows dates that fall after this date,
// Before is always set if After is set
func (r *CalendarDisabledRange) After() *string {
	if r.AfterDate == nil {
		return nil
	}
	return r.AfterDate.ToStringPtr()
}

// From if set (not nil) allows dates from this date into the future
func (r *CalendarDisabledRange) From() *string {
	if r.FromDate == nil {
		return nil
	}
	return r.FromDate.ToStringPtr()
}

// To if set allows dates before and including this date
func (r *CalendarDisabledRange) To() *string {
	if r.ToDate == nil {
		return nil
	}
	return r.ToDate.ToStringPtr()
}
