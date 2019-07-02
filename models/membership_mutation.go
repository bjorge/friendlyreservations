package models

// UpdateMembershipStatusInputGQL is the GQL string for creating a new user
const UpdateMembershipStatusInputGQL = `
# Information to update membership status
input UpdateMembershipInput {
	# the version of the property being updated
	forVersion: Int!
	# this update is for this user
	updateForUserId: String!
	# the id of the membership restriction
	restrictionId: String!
	# true == purchase membership, false == optout from membership
	purchase: Boolean!
	# the administrator is making this update for some user
	adminUpdate: Boolean!
	# if adminUpdate is true, then a comment is required
	comment: String
}
`

type UpdateMembershipStatusInput struct {
	// Fields received from the client
	ForVersion      int32
	UpdateForUserId string
	RestrictionId   string
	Purchase        bool
	AdminUpdate     bool
	Comment         *string

	// Extra fields persisted with the above
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

func (r *UpdateMembershipStatusInput) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *UpdateMembershipStatusInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

func (r *UpdateMembershipStatusInput) GetForVersion() int32 {
	return r.ForVersion
}
