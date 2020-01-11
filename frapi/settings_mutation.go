package frapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
)

// UpdateSettings is called to modify the current settings for a property
func (r *Resolver) UpdateSettings(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.UpdateSettingsInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Update Settings")

	// get the current property
	property, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, args.Input, property); duplicate || err != nil {
		if err == nil {
			return property, nil
		}
		return nil, err
	}

	// check the input values
	if !me.IsAdmin() {
		return nil, errors.New("only admins can update settings")
	}

	constraints, err := property.UpdateSettingsConstraints(ctx)
	if err != nil {
		return nil, err
	}

	stringArg, err := trim(args.Input.PropertyName)
	if err != nil {
		return nil, err
	}
	args.Input.PropertyName = *stringArg

	if len(args.Input.PropertyName) < int(constraints.PropertyNameMin()) {
		return nil, fmt.Errorf("property name too short")
	}

	if len(args.Input.PropertyName) > int(constraints.PropertyNameMax()) {
		return nil, fmt.Errorf("property name too long")
	}

	switch args.Input.Currency {
	case models.USD:
	case models.EUR:
		break
	default:
		return nil, fmt.Errorf("unknown currency %+v", args.Input.Currency)
	}

	if _, err := frdate.NewDateBuilder(args.Input.Timezone); err != nil {
		return nil, err
	}

	if args.Input.BalanceReminderIntervalDays < constraints.BalanceReminderIntervalDaysMin() ||
		args.Input.BalanceReminderIntervalDays > constraints.BalanceReminderIntervalDaysMax() {
		return nil, fmt.Errorf("BalanceReminderIntervalDays out of range %+v", args.Input.BalanceReminderIntervalDays)
	}

	if args.Input.MaxOutDays < constraints.MaxOutDaysMin() ||
		args.Input.MaxOutDays > constraints.MaxOutDaysMax() {
		return nil, fmt.Errorf("MaxOutDays out of range %+v", args.Input.MaxOutDays)
	}

	if args.Input.MemberRate < constraints.MemberRateMin() ||
		args.Input.MemberRate > constraints.MemberRateMax() {
		return nil, fmt.Errorf("MemberRate out of range %+v", args.Input.MemberRate)
	}

	if args.Input.NonMemberRate < constraints.NonMemberRateMin() ||
		args.Input.NonMemberRate > constraints.NonMemberRateMax() {
		return nil, fmt.Errorf("NonMemberRate out of range %+v", args.Input.NonMemberRate)
	}

	if args.Input.MinBalance < constraints.MinBalanceMin() ||
		args.Input.MinBalance > constraints.MinBalanceMax() {
		return nil, fmt.Errorf("MinBalance out of range %+v", args.Input.MinBalance)
	}

	if args.Input.ReservationReminderDaysBefore < constraints.ReservationReminderDaysBeforeMin() ||
		args.Input.ReservationReminderDaysBefore > constraints.ReservationReminderDaysBeforeMax() {
		return nil, fmt.Errorf("ReservationReminderDaysBefore out of range %+v", args.Input.ReservationReminderDaysBefore)
	}

	if args.Input.MinInDays < constraints.MinInDaysMin() ||
		args.Input.MinInDays > constraints.MinInDaysMax() {
		return nil, fmt.Errorf("MinInDays out of range %+v", args.Input.MinInDays)
	}

	// update the request with more information
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.AuthorUserId = me.UserID()

	// // persist the event
	// paramGroup := templates.Reservation
	// newNotificationInput := createNotificationRecord(NOTIFICATION_TARGET_ALL_MEMBERS, propertyResolver, templates.NEW_RESERVATION,
	// 	nil, &paramGroup, &args.Input.ReservationId)

	// propertyResolver, err = commitChanges(ctx, args.PropertyId, args.Input, newNotificationInput)
	property, err = commitChanges(ctx, args.PropertyID, property.EventVersion(), args.Input)

	// if err == nil {
	// 	// send the email notification
	// 	notifications, _ := propertyResolver.Notifications(&NotificationArgs{notificationId: &newNotificationInput.NotificationId})
	// 	sendEmail(ctx, propertyResolver, notifications[0])
	// }

	return property, err
}
