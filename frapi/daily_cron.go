package frapi

import (
	"context"
	"errors"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/platform"
	"github.com/bjorge/friendlyreservations/templates"
	"github.com/bjorge/friendlyreservations/utilities"
)

// DailyCron is called by the service (ex. appengine) once per day
func DailyCron(ctx context.Context) error {

	Logger.LogInfof("DailyCron start")

	resolver := &Resolver{}
	properties, err := resolver.cronProperties(ctx)

	if err != nil {
		Logger.LogErrorf("DailyCron error accessing properties: %+v", err)
		return err
	}

	settingsConstraints, err := resolver.UpdateSettingsConstraints(ctx)
	if err != nil {
		Logger.LogErrorf("DailyCron error accessing settings constraints: %+v", err)
		return err
	}

	trialOn := settingsConstraints.TrialOn()
	trialDays := settingsConstraints.TrialDays()

	// iterate through all the properties
	for _, property := range properties {
		settings, err := property.Settings(&settingsArgs{})
		if err != nil {
			Logger.LogErrorf("DailyCron error for property: %+v", err)
			continue
		}
		dateBuilder := frdate.MustNewDateBuilder(settings.Timezone())
		today := dateBuilder.MustNewDateTime(frdate.CreateDateTimeUTC())

		if trialOn {
			Logger.LogDebugf("DailyCron check trial period for propertyId: %+v", property.PropertyID())
			created := dateBuilder.MustNewDateTime(property.CreateDateTime())
			if created.AddDays(int(trialDays)).Before(today) {
				Logger.LogDebugf("DailyCron trial is over for propertyId: %+v, time to delete", property.PropertyID())
				// ok, trial is over! time to delete the property
				success, err := resolver.internalDeleteProperty(ctx, property.PropertyID())
				if err != nil {
					Logger.LogErrorf("DailyCron error deleting property: %+v", err)
					continue
				}
				Logger.LogDebugf("DailyCron delete propertyId status: %+v", success)
				continue
			}
		}

		Logger.LogDebugf("check balance for propertyId: %+v", property.PropertyID())

		// get the last notifications to make sure we don't send too many

		lastNotifications := property.lastNotifications(dateBuilder)

		// first search for users with negative balances
		lastActiveBalances := property.lastActiveLedgerBalances()

		for userID, balance := range lastActiveBalances {
			Logger.LogDebugf("DailyCron check balance for userId: %+v", userID)
			if balance < 0 {
				Logger.LogDebugf("DailyCron balance negative")
				notify := false
				if dateTime, ok := lastNotifications[templates.LowBalanceNotification][userID]; ok {
					// a notification has already been sent, check if too early to send another one
					nextNotification := dateTime.AddDays(int(settings.BalanceReminderIntervalDays()))
					if today.After(nextNotification) {
						Logger.LogDebugf("DailyCron time expired, send notification")
						notify = true
					}
				} else {
					// first time
					Logger.LogDebugf("DailyCron first time, send notification")
					notify = true
				}

				if notify {
					Logger.LogDebugf("DailyCron commit notification")
					paramGroup := templates.Ledger
					newNotificationInput := createNotificationRecord(notificationTargetMember, property, templates.LowBalanceNotification,
						&userID, &paramGroup, &userID)

					property, err = commitCronChanges(ctx, property.PropertyID(), newNotificationInput)

					if err != nil {
						Logger.LogErrorf("DailyCron error commit changes: %+v", err)
					} else {
						// send the email notification
						notifications, _ := property.Notifications(&notificationArgs{notificationID: &newNotificationInput.NotificationId})
						Logger.LogDebugf("DailyCron send notification email")
						sendEmail(ctx, property, notifications[0])
					}
				}
			}
		}
	}
	Logger.LogInfof("DailyCron end success")

	return nil
}

func (r *Resolver) cronProperties(ctx context.Context) ([]*PropertyResolver, error) {
	var l []*PropertyResolver

	// check that system user has been set
	if utilities.SystemEmail == "" {
		return nil, errors.New("system user email not set")
	}

	// get all the properties
	ids, err := persistedPropertyList.GetProperties(ctx)
	if err != nil {
		return nil, err
	}

	for _, propertyID := range ids {
		propertyResolver, err := currentBaseProperty(ctx, utilities.SystemEmail, propertyID)
		if err != nil {
			return nil, err
		}
		l = append(l, propertyResolver)
	}
	return l, nil
}

func commitCronChanges(ctx context.Context, propertyID string,
	events ...platform.VersionedEvent) (*PropertyResolver, error) {

	eventList := []platform.VersionedEvent{}
	for _, event := range events {
		eventList = append(eventList, event)
	}

	key, err := persistedVersionedEvents.GetNextEventID(ctx, propertyID, false)
	if err != nil {
		return nil, err
	}
	_, err = persistedVersionedEvents.NewPropertyEvents(ctx, propertyID, key, eventList, false)
	if err != nil {
		return nil, err
	}

	// get all events for the property
	propertyResolver, err := currentBaseProperty(ctx, utilities.SystemEmail, propertyID)
	if err != nil {
		return nil, err
	}

	return propertyResolver, err
}
