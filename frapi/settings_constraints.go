package frapi

import (
	"context"

	"github.com/bjorge/friendlyreservations/utilities"
)

const settingsConstraintsGQL = `

type UpdateSettingsConstraints {
	propertyNameMin: Int!
	propertyNameMax: Int!
	memberRateMin: Int!
	memberRateMax: Int!
	nonMemberRateMin: Int!
	nonMemberRateMax: Int!
	minBalanceMin: Int!
	minBalanceMax: Int!
	maxOutDaysMin: Int!
	maxOutDaysMax: Int!
	minInDaysMin: Int!
	minInDaysMax: Int!
	reservationReminderDaysBeforeMin: Int!
	reservationReminderDaysBeforeMax: Int!
	balanceReminderIntervalDaysMin: Int! 
	balanceReminderIntervalDaysMax: Int!
	allowNewProperty: Boolean!
	allowPropertyImport: Boolean!
	allowPropertyDelete: Boolean!
	trialOn: Boolean!
	trialDays: Int!
}
`

// UpdateSettingsConstraints provides the retriever for constraint methods
type UpdateSettingsConstraints struct {
	ctx context.Context
}

// UpdateSettingsConstraints returns input constraints for updating settings
func (r *PropertyResolver) UpdateSettingsConstraints(ctx context.Context) (*UpdateSettingsConstraints, error) {

	return &UpdateSettingsConstraints{ctx}, nil
}

// UpdateSettingsConstraints returns input constraints for updating settings
func (r *Resolver) UpdateSettingsConstraints(ctx context.Context) (*UpdateSettingsConstraints, error) {

	return &UpdateSettingsConstraints{ctx}, nil
}

// PropertyNameMin returns min length
func (r *UpdateSettingsConstraints) PropertyNameMin() int32 { return 3 }

// PropertyNameMax returns max length
func (r *UpdateSettingsConstraints) PropertyNameMax() int32 { return 25 }

// MemberRateMin returns min value
func (r *UpdateSettingsConstraints) MemberRateMin() int32 { return 1 }

// MemberRateMax returns max value
func (r *UpdateSettingsConstraints) MemberRateMax() int32 { return 100000 }

// NonMemberRateMin returns min value
func (r *UpdateSettingsConstraints) NonMemberRateMin() int32 { return 1 }

// NonMemberRateMax returns max value
func (r *UpdateSettingsConstraints) NonMemberRateMax() int32 { return 100000 }

// MinBalanceMin returns min value
func (r *UpdateSettingsConstraints) MinBalanceMin() int32 { return -100000 }

// MinBalanceMax returns min value
func (r *UpdateSettingsConstraints) MinBalanceMax() int32 { return 100000 }

// MaxOutDaysMin returns min value
func (r *UpdateSettingsConstraints) MaxOutDaysMin() int32 { return 10 }

// MaxOutDaysMax returns max value
func (r *UpdateSettingsConstraints) MaxOutDaysMax() int32 { return 365 }

// MinInDaysMin returns min value
func (r *UpdateSettingsConstraints) MinInDaysMin() int32 { return -10 }

// MinInDaysMax returns min value
func (r *UpdateSettingsConstraints) MinInDaysMax() int32 { return 10 }

// ReservationReminderDaysBeforeMin returns min value
func (r *UpdateSettingsConstraints) ReservationReminderDaysBeforeMin() int32 { return 1 }

// ReservationReminderDaysBeforeMax returns min value
func (r *UpdateSettingsConstraints) ReservationReminderDaysBeforeMax() int32 { return 14 }

// BalanceReminderIntervalDaysMin returns min value
func (r *UpdateSettingsConstraints) BalanceReminderIntervalDaysMin() int32 { return 2 }

// BalanceReminderIntervalDaysMax returns min value
func (r *UpdateSettingsConstraints) BalanceReminderIntervalDaysMax() int32 { return 40 }

// AllowNewProperty is true if a new property creation is allowed
func (r *UpdateSettingsConstraints) AllowNewProperty() bool {
	if !utilities.AllowNewProperty {
		return false
	}

	// check that a user is logged in
	u := GetUser(r.ctx)
	if u == nil {
		return false
	}

	// get all the properties
	emailRecords, err := PersistedEmailStore.GetPropertiesByEmail(r.ctx, u.Email)
	Logger.LogDebugf("User %+v has %+v properties", u.Email, len(emailRecords))

	if err == nil {
		if len(emailRecords) >= 5 {
			Logger.LogDebugf("Create new property not allowed since user %+v in too many properties", u.Email)
			return false
		}
	}

	return true
}

// AllowPropertyDelete is true if a the property can be deleted
func (r *UpdateSettingsConstraints) AllowPropertyDelete() bool {
	return utilities.AllowDeleteProperty
}

// AllowPropertyImport is true if import of a property is allowed
func (r *UpdateSettingsConstraints) AllowPropertyImport() bool {
	return utilities.ImportFileName != ""
}

// TrialOn is true if the site is setup for trial properties only
func (r *UpdateSettingsConstraints) TrialOn() bool {
	return utilities.TrialDuration.Hours() > 0
}

// TrialDays returns the days at which a property will be deleted during a trial
func (r *UpdateSettingsConstraints) TrialDays() int32 {
	return int32(utilities.TrialDuration.Hours() / 24)
}
