package frapi

import (
	"context"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
)

// NotificationRead marks a notification as having been read
func (r *Resolver) NotificationRead(ctx context.Context, args *struct {
	PropertyID     string
	NotificationID string
	ForVersion     int32
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Mark notification as read")

	property, me, err := currentProperty(ctx, args.PropertyID)

	if err != nil {
		return nil, err
	}

	notificationReadInput := &models.NotificationReadInput{}

	// TODO: validate that the notification exists...
	notificationReadInput.NotificationId = args.NotificationID
	notificationReadInput.ForVersion = args.ForVersion

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, notificationReadInput, property); duplicate || err != nil {
		if err == nil {
			return property, nil
		}
		return nil, err
	}

	// input looks good, now add extra internal values
	notificationReadInput.CreateDateTime = frdate.CreateDateTimeUTC()
	notificationReadInput.AuthorUserId = me.UserID()

	// persist the event
	return commitChanges(ctx, args.PropertyID, property.EventVersion(), notificationReadInput)
}
