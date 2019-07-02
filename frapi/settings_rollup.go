package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// SettingsRollup is an internal struct used duing rollup of settings
type SettingsRollup struct {
	PropertyName                  string
	Currency                      models.Currency
	MemberRate                    int32
	AllowNonMembers               bool
	NonMemberRate                 int32
	Timezone                      string
	MinBalance                    int32
	MaxOutDays                    int32
	MinInDays                     int32
	EventVersion                  int32
	ReservationReminderDaysBefore int32
	BalanceReminderIntervalDays   int32
}

// GetEventVersion returns version of rollup item
func (r *SettingsRollup) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *PropertyResolver) rollupSettings() {

	r.rollupMutexes[settingsRollupType].Lock()
	defer r.rollupMutexes[settingsRollupType].Unlock()

	if !r.rollupsExists(settingsRollupType) {
		// go through all the events and process the settings ones
		for _, event := range r.property.Events {

			switch settingsEvent := event.(type) {

			case *models.NewPropertyInput:

				settings := &SettingsRollup{}
				settings.PropertyName = settingsEvent.PropertyName
				settings.Currency = settingsEvent.Currency
				settings.MemberRate = settingsEvent.MemberRate
				settings.AllowNonMembers = settingsEvent.AllowNonMembers
				settings.NonMemberRate = settingsEvent.NonMemberRate
				settings.Timezone = settingsEvent.Timezone
				settings.MaxOutDays = 365
				settings.MinInDays = 0
				settings.MinBalance = -100000
				settings.ReservationReminderDaysBefore = 3
				settings.BalanceReminderIntervalDays = 2

				r.addRollup(settingsID,
					settings, settingsRollupType)

			case *models.UpdateSettingsInput:

				id := settingsID
				ifaces := r.getRollups(&rollupArgs{id: &id}, settingsRollupType)

				lastSettings := ifaces[0].(*SettingsRollup)

				// make a copy
				settings := *lastSettings

				settings.PropertyName = settingsEvent.PropertyName
				settings.Currency = settingsEvent.Currency
				settings.MemberRate = settingsEvent.MemberRate
				settings.AllowNonMembers = settingsEvent.AllowNonMembers
				settings.NonMemberRate = settingsEvent.NonMemberRate
				settings.Timezone = settingsEvent.Timezone
				settings.MaxOutDays = settingsEvent.MaxOutDays
				settings.MinInDays = settingsEvent.MinInDays
				settings.MinBalance = settingsEvent.MinBalance
				settings.ReservationReminderDaysBefore = settingsEvent.ReservationReminderDaysBefore
				settings.BalanceReminderIntervalDays = settingsEvent.BalanceReminderIntervalDays

				settings.EventVersion = settingsEvent.EventVersion

				r.addRollup(settingsID,
					&settings, settingsRollupType)
			}
		}
		cacheError := r.cacheRollup(settingsRollupType)
		if cacheError != nil {
			utilities.LogWarningf(r.ctx, "cache write setting rollups error: %+v", cacheError)
		}
	}
}
