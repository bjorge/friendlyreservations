package models

// UpdateSettingsInputGQL is the GQL string for updating settings
const UpdateSettingsInputGQL = `
input UpdateSettingsInput {
	forVersion: String!
	propertyName: String!
	currency: Currency!
	memberRate: Int!
	allowNonMembers: Boolean!
	nonMemberRate: Int!
	timezone: String!
	minBalance: Int!
	maxOutDays: Int!
	minInDays: Int!
	reservationReminderDaysBefore: Int!
	balanceReminderIntervalDays: Int!
}
`

// UpdateSettingsInput is the gql input for updating settings
type UpdateSettingsInput struct {
	// Fields received from the client
	ForVersion                    int32
	PropertyName                  string
	Currency                      Currency
	MemberRate                    int32
	AllowNonMembers               bool
	NonMemberRate                 int32
	Timezone                      string
	MinBalance                    int32
	MaxOutDays                    int32
	MinInDays                     int32
	ReservationReminderDaysBefore int32
	BalanceReminderIntervalDays   int32

	// Extra fields persisted with the above
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

// GetEventVersion returns the version of the settings update event
func (r *UpdateSettingsInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *UpdateSettingsInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *UpdateSettingsInput) GetForVersion() int32 {
	return r.ForVersion
}
