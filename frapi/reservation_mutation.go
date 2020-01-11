package frapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/templates"
	"github.com/bjorge/friendlyreservations/utilities"
)

// CancelReservation is called to cancel a reservation
func (r *Resolver) CancelReservation(ctx context.Context, args *struct {
	PropertyID    string
	ForVersion    int32
	ReservationID string
	AdminRequest  *bool
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Cancel Reservation")

	adminRequest := false
	if args.AdminRequest != nil && *args.AdminRequest {
		adminRequest = true
	}

	// get the current property
	property, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	if adminRequest && !me.IsAdmin() {
		return nil, errors.New("cancel request for admin but user is not an admin")
	}

	// get the reservation
	reservations, err := property.Reservations(&reservationsArgs{ReservationID: &args.ReservationID})
	if len(reservations) != 1 {
		return nil, fmt.Errorf("reservation not found for id: %+v", args.ReservationID)
	}

	// get the cancel constraints
	var cancelConstraints *CancelReservationConstraints
	if adminRequest {
		cancelConstraints, err = property.CancelReservationConstraints(ctx, &CancelReservationConstraintsArgs{UserType: ADMIN})
	} else {
		userID := me.UserID()
		cancelConstraints, err = property.CancelReservationConstraints(ctx, &CancelReservationConstraintsArgs{UserType: MEMBER, UserID: &userID})
	}
	if err != nil {
		return nil, err
	}

	cancelAllowed := false
	for _, id := range cancelConstraints.CancelReservationAllowed() {
		if *id == args.ReservationID {
			cancelAllowed = true
		}
	}

	if !cancelAllowed {
		return nil, errors.New("cancel reservation is not allowed")
	}

	// store the event!
	cancelReservationInput := &models.CancelReservationInput{}
	cancelReservationInput.ForVersion = args.ForVersion
	cancelReservationInput.AdminRequest = adminRequest
	cancelReservationInput.AuthorUserId = me.UserID()
	cancelReservationInput.ReservationId = args.ReservationID
	cancelReservationInput.CreateDateTime = frdate.CreateDateTimeUTC()
	cancelReservationInput.ReservedForUserId = reservations[0].ReservedFor().UserID()

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, cancelReservationInput, property); duplicate || err != nil {
		if err == nil {
			return property, nil
		}
		return nil, err
	}

	// persist the event
	paramGroup := templates.Reservation
	newNotificationInput := createNotificationRecord(notificationTargetAllMembers, property, templates.CancelReservationNotification,
		nil, &paramGroup, &args.ReservationID)

	property, err = commitChanges(ctx, args.PropertyID, property.EventVersion(), cancelReservationInput, newNotificationInput)

	if err == nil {
		// send the email notification
		notifications, _ := property.Notifications(&notificationArgs{notificationID: &newNotificationInput.NotificationId})
		sendEmail(ctx, property, notifications[0])
	}

	return property, err
}

// CreateReservation is called to create a new reservation
func (r *Resolver) CreateReservation(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.NewReservationInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Create Reservation")

	// get the current property
	propertyResolver, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, args.Input, propertyResolver); duplicate || err != nil {
		if err == nil {
			return propertyResolver, nil
		}
		return nil, err
	}

	if args.Input.AdminRequest && !me.IsAdmin() {
		return nil, errors.New("admin request for non-admin")
	}

	// check the input values
	if me.UserID() != args.Input.ReservedForUserId && !args.Input.AdminRequest {
		return nil, errors.New("cannot reserve for another user if not an admin request")
	}

	users := propertyResolver.Users(&usersArgs{UserID: &args.Input.ReservedForUserId})
	if len(users) != 1 {
		return nil, fmt.Errorf("unknown user with user id (%+v)",
			args.Input.ReservedForUserId)
	}

	settings, _ := propertyResolver.Settings(&settingsArgs{})
	b, err := frdate.NewDateBuilder(settings.Timezone())
	if err != nil {
		return nil, err
	}

	checkIn, err := b.NewDate(args.Input.StartDate)
	if err != nil {
		return nil, err
	}

	checkOut, err := b.NewDate(args.Input.EndDate)
	if err != nil {
		return nil, err
	}

	if checkOut.Before(checkIn) || checkIn.ToString() == checkOut.ToString() {
		return nil, fmt.Errorf("check out date (%+v) must be after check in date (%+v)",
			checkOut.ToString(), checkIn.ToString())
	}

	var constraints *NewReservationConstraints
	if args.Input.AdminRequest {
		constraints, err = propertyResolver.NewReservationConstraints(ctx, &NewReservationConstraintsArgs{UserType: ADMIN})
	} else if args.Input.Member {
		constraints, err = propertyResolver.NewReservationConstraints(ctx, &NewReservationConstraintsArgs{UserID: &args.Input.ReservedForUserId, UserType: MEMBER})
	} else {
		constraints, err = propertyResolver.NewReservationConstraints(ctx, &NewReservationConstraintsArgs{UserID: &args.Input.ReservedForUserId, UserType: NONMEMBER})
	}

	if err != nil {
		return nil, err
	}

	if !args.Input.Member && args.Input.NonMemberName == nil {
		return nil, fmt.Errorf("non member name is required")
	} else if !args.Input.Member {
		stringArg, err := trim(*args.Input.NonMemberName)
		if err != nil {
			return nil, err
		}
		if len(*stringArg) < int(constraints.NonMemberNameMin()) {
			return nil, fmt.Errorf("non member name too short: %+v", *stringArg)
		}
		if len(*stringArg) > int(constraints.NonMemberNameMax()) {
			return nil, fmt.Errorf("non member name too long: %+v", *stringArg)
		}
		args.Input.NonMemberName = stringArg
	}

	if !args.Input.Member && args.Input.NonMemberInfo == nil {
		return nil, fmt.Errorf("non member info is required")
	} else if !args.Input.Member {
		stringArg, err := trim(*args.Input.NonMemberInfo)
		if err != nil {
			return nil, err
		}
		if len(*stringArg) < int(constraints.NonMemberInfoMin()) {
			return nil, fmt.Errorf("non member info too short: %+v", *stringArg)
		}
		if len(*stringArg) > int(constraints.NonMemberInfoMax()) {
			return nil, fmt.Errorf("non member info too long: %+v", *stringArg)
		}
		args.Input.NonMemberInfo = stringArg
	}

	// check if requested dates are allowed
	disabled, err := propertyResolver.newReservationDisabled(ctx, checkIn, checkOut, constraints)
	if err != nil {
		return nil, err
	}
	if disabled {
		return nil, fmt.Errorf("reservation dates not allowed in range %+v to %+v, check balance and/or restrictions",
			checkIn.ToString(), checkOut.ToString())
	}

	// update the request with more information
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.ReservationId = utilities.NewGUID()
	args.Input.AuthorUserId = me.UserID()

	args.Input.Rate = []models.DailyRate{}
	days, _ := frdate.DaysList(checkIn, checkOut, false)
	for _, day := range days {
		if args.Input.Member {
			args.Input.Rate = append(args.Input.Rate, models.DailyRate{Amount: settings.memberRateInternal(), Date: day.ToString()})
		} else {
			args.Input.Rate = append(args.Input.Rate, models.DailyRate{Amount: settings.nonMemberRateInternal(), Date: day.ToString()})
		}
	}

	// persist the event
	paramGroup := templates.Reservation
	newNotificationInput := createNotificationRecord(notificationTargetAllMembers, propertyResolver, templates.NewReservationNotification,
		nil, &paramGroup, &args.Input.ReservationId)

	propertyResolver, err = commitChanges(ctx, args.PropertyID, propertyResolver.EventVersion(), args.Input, newNotificationInput)

	if err == nil {
		// send the email notification
		notifications, _ := propertyResolver.Notifications(&notificationArgs{notificationID: &newNotificationInput.NotificationId})
		sendEmail(ctx, propertyResolver, notifications[0])
	}

	return propertyResolver, err
}

// internal method
func (r *PropertyResolver) newReservationDisabled(ctx context.Context, checkIn *frdate.Date, checkOut *frdate.Date, constraints *NewReservationConstraints) (bool, error) {

	if !constraints.NewReservationAllowed() {
		return true, nil
	}

	inRanges := constraints.CheckinDisabled()
	outRanges := constraints.CheckoutDisabled()

	lastIn := checkOut.AddDays(-1)
	firstOut := checkIn.AddDays(1)

	for _, inRange := range inRanges {
		if inRange.From() != nil && inRange.To() != nil {
			if frdate.DateOverlap(checkIn, lastIn, inRange.FromDate, inRange.ToDate) {
				return true, nil
			}
		}
		if inRange.After() != nil {
			// example disabled
			// --IIIIII---
			// ------A----
			if lastIn.After(inRange.AfterDate) {
				return true, nil
			}
		}
		if inRange.Before() != nil {
			// example disabled
			// --IIIIII---
			// ---B-------
			if checkIn.Before(inRange.BeforeDate) {
				return true, nil
			}
		}
	}
	for _, outRange := range outRanges {
		if outRange.From() != nil && outRange.To() != nil {
			if frdate.DateOverlap(firstOut, checkOut, outRange.FromDate, outRange.ToDate) {
				return true, nil
			}
		}
		if outRange.After() != nil {
			// example disabled
			// --OOOOOO---
			// ------A----
			if checkOut.After(outRange.AfterDate) {
				return true, nil
			}
		}
		if outRange.Before() != nil {
			// example disabled
			// --OOOOOO---
			// ---B-------
			if firstOut.Before(outRange.BeforeDate) {
				return true, nil
			}
		}
	}

	return false, nil
}
