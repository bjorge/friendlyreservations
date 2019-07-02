package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// NotificationRollup is internal structure to hold reservation information for resolvers
type NotificationRollup struct {
	Input            *models.NewNotificationInput
	ReaderUserIdsMap map[string]bool
	EventVersion     int32
	TargetUserIdsMap map[string]bool
}

// GetEventVersion returns the version of the rollup record
func (r *NotificationRollup) GetEventVersion() int {
	return int(r.Input.EventVersion)
}

func (r *PropertyResolver) rollupNotifications() {

	r.rollupMutexes[notificationRollupType].Lock()
	defer r.rollupMutexes[notificationRollupType].Unlock()

	// notifications are rolled up as they are read
	if !r.rollupsExists(notificationRollupType) {

		for _, event := range r.property.Events {
			if newNotificationEvent, ok := event.(*models.NewNotificationInput); ok {
				notificationRecord := &NotificationRollup{}
				notificationRecord.Input = newNotificationEvent
				notificationRecord.ReaderUserIdsMap = make(map[string]bool)
				notificationRecord.EventVersion = newNotificationEvent.EventVersion
				notificationRecord.TargetUserIdsMap = make(map[string]bool)
				for _, userID := range newNotificationEvent.ToUserIds {
					notificationRecord.TargetUserIdsMap[userID] = true
				}
				for _, userID := range newNotificationEvent.CcUserIds {
					notificationRecord.TargetUserIdsMap[userID] = true
				}

				r.addRollup(notificationRecord.Input.NotificationId,
					notificationRecord, notificationRollupType)
			}

			if notificationReadInput, ok := event.(*models.NotificationReadInput); ok {
				ifaces := r.getRollups(&rollupArgs{id: &notificationReadInput.NotificationId}, notificationRollupType)
				rollup, _ := ifaces[0].(*NotificationRollup)

				// make a copy
				notification := *rollup
				notification.ReaderUserIdsMap[notificationReadInput.AuthorUserId] = true
				notification.EventVersion = notificationReadInput.EventVersion

				r.addRollup(notification.Input.NotificationId,
					&notification, notificationRollupType)
			}
		}
		cacheError := r.cacheRollup(notificationRollupType)
		if cacheError != nil {
			utilities.LogWarningf(r.ctx, "cache write notification rollups error: %+v", cacheError)
		}
	}
}
