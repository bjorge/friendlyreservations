package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
)

// UserRollup is the user event rollup structure
type UserRollup struct {
	UserID       string
	IsAdmin      bool
	IsMember     bool
	IsSystem     bool
	State        models.UserState
	Nickname     string
	EmailID      string
	EventVersion int32
}

// GetEventVersion returns the version of the rollup record
func (r *UserRollup) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *PropertyResolver) rollupUsers() {

	r.rollupMutexes[userRollupType].Lock()
	defer r.rollupMutexes[userRollupType].Unlock()

	if !r.rollupsExists(userRollupType) {

		for _, event := range r.property.Events {

			switch userEvent := event.(type) {
			case *models.NewUserInput:
				userRollup := &UserRollup{}
				userRollup.UserID = userEvent.UserId

				userRollup.IsSystem = userEvent.IsSystem
				userRollup.IsAdmin = userEvent.IsAdmin
				userRollup.IsMember = userEvent.IsMember
				userRollup.State = userEvent.State
				userRollup.Nickname = userEvent.Nickname

				userRollup.EmailID = userEvent.EmailId

				userRollup.EventVersion = userEvent.EventVersion

				r.addRollup(userEvent.UserId,
					userRollup, userRollupType)

			case *models.UpdateUserInput:
				rollups := r.getRollups(&rollupArgs{id: &userEvent.UserId}, userRollupType)
				if len(rollups) > 0 {
					rollup := rollups[0]
					if user, ok := rollup.(*UserRollup); ok {
						// make a copy
						userRollup := *user

						userRollup.UserID = userEvent.UserId

						userRollup.IsSystem = userEvent.IsSystem
						userRollup.IsAdmin = userEvent.IsAdmin
						userRollup.IsMember = userEvent.IsMember
						userRollup.State = userEvent.State
						userRollup.Nickname = userEvent.Nickname

						userRollup.EmailID = userEvent.EmailId

						userRollup.EventVersion = userEvent.EventVersion

						r.addRollup(userEvent.UserId,
							&userRollup, userRollupType)
					}
				}

			case *models.AcceptInvitationInput:
				// get the user
				rollups := r.getRollups(&rollupArgs{id: &userEvent.AuthorUserId}, userRollupType)
				if len(rollups) > 0 {
					rollup := rollups[0]
					if user, ok := rollup.(*UserRollup); ok {
						// make a copy
						userRollup := *user

						if userEvent.Accept {
							userRollup.State = models.ACCEPTED
						} else {
							userRollup.State = models.DECLINED
						}

						userRollup.EventVersion = userEvent.EventVersion

						r.addRollup(userEvent.AuthorUserId,
							&userRollup, userRollupType)
					}
				}

			}

		}
		cacheError := r.cacheRollup(userRollupType)
		if cacheError != nil {
			Logger.LogWarningf("cache write user rollups error: %+v", cacheError)
		}
	}
}
