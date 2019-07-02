package frapi

import (
	"context"
	"errors"
	"strings"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/templates"
	"github.com/bjorge/friendlyreservations/utilities"
)

// UpdateBalance changes the current balance for a user
func (r *Resolver) UpdateBalance(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.UpdateBalanceInput
}) (*PropertyResolver, error) {
	utilities.DebugLog(ctx, "Update Balance")

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
	if !me.IsAdmin() {
		return nil, errors.New("only admins can make payments")
	}

	constraints, err := propertyResolver.UpdateBalanceConstraints()
	if err != nil {
		return nil, err
	}

	if args.Input.Amount < constraints.AmountMin() {
		return nil, errors.New("amount too small")
	}

	if args.Input.Amount > constraints.AmountMax() {
		return nil, errors.New("amount too big")
	}

	args.Input.Description = strings.TrimSpace(args.Input.Description)
	if len(args.Input.Description) < int(constraints.DescriptionMin()) {
		return nil, errors.New("description too small")
	}

	if len(args.Input.Description) > int(constraints.DescriptionMax()) {
		return nil, errors.New("description too big")
	}

	users := propertyResolver.Users(&usersArgs{UserID: &args.Input.UpdateForUserId})
	if len(users) != 1 {
		return nil, errors.New("user does not exist")
	}

	// input looks good, now add extra internal values
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.PaymentId = utilities.NewGUID()
	args.Input.AuthorUserId = me.UserID()

	paramGroup := templates.Ledger
	newNotificationInput := createNotificationRecord(notificationTargetMember, propertyResolver, templates.BalanceChangeNotification,
		&args.Input.UpdateForUserId, &paramGroup, &args.Input.UpdateForUserId)

	propertyResolver, err = commitChanges(ctx, args.PropertyID, propertyResolver.EventVersion(), args.Input, newNotificationInput)

	if err == nil {
		// send the email notification
		notifications, _ := propertyResolver.Notifications(&notificationArgs{notificationID: &newNotificationInput.NotificationId})
		sendEmail(ctx, propertyResolver, notifications[0])
	}

	return propertyResolver, err
}
