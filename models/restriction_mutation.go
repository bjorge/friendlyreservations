package models

// NewRestrictionInputGQL is the GQL string for creating a new restriction
const NewRestrictionInputGQL = `
# Information to create a new restriction.
# Only one restriction must be specified.
input NewRestrictionInput {
	forVersion: Int!
	blackout: BlackoutRestrictionInput
	membership: MembershipRestrictionInput
	description: String!
}
`

type NewRestrictionInput struct {
	ForVersion  int32
	Blackout    *BlackoutRestriction
	Membership  *MembershipRestriction
	Description string

	// Extra fields persisted with the above
	RestrictionId  string
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

func (r *NewRestrictionInput) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *NewRestrictionInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *NewRestrictionInput) GetForVersion() int32 {
	return r.ForVersion
}

// BlackoutRestrictionInputGQL is the GQL string for creating a new blackout restriction
const BlackoutRestrictionInputGQL = `
# Specify a blackout period.
input BlackoutRestrictionInput {
	# Before start date no blackout.
	startDate: String!
	# End date and later no blackout.
	endDate: String!
}
`

type BlackoutRestriction struct {
	StartDate string
	EndDate   string
}

// MembershipRestrictionInputGQL is the GQL string for creating a new membership restriction
const MembershipRestrictionInputGQL = `
# Specify a blackout period.
input MembershipRestrictionInput {
	# Start pre-pay period
	prePayStartDate: String!
	# Start checkin date for membership period
	inDate: String!
	# End checkout date for membership period.
	outDate: String!
	# End checkout date grace period
	gracePeriodOutDate: String!
	# Payment amount for membership period
	amount: Int!
}
`

type MembershipRestriction struct {
	PrePayStartDate    string
	InDate             string
	OutDate            string
	GracePeriodOutDate string
	Amount             int32
}
