package models

// NewUserInputGQL is the GQL string for creating a new user
const NewUserInputGQL = `
# Information to create a new user.
input NewUserInput {
	forVersion: String!
	email: String!
	isAdmin: Boolean!
	isMember: Boolean!
	nickname: String!
}
`

// NewUserInput is the GQL structure for creating a new user
type NewUserInput struct {
	// Fields received from the client
	// Email from client, remove before persisting, only persist EmailId
	ForVersion int32
	Email      string
	IsAdmin    bool
	IsMember   bool
	Nickname   string

	// Extra fields persisted with the above
	IsSystem       bool
	UserId         string
	CreateDateTime string
	AuthorUserId   string
	EmailId        string
	State          UserState
	EventVersion   int32
}

// GetEventVersion returns the version of the event
func (r *NewUserInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *NewUserInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *NewUserInput) GetForVersion() int32 {
	return r.ForVersion
}

// UpdateUserInputGQL is the GQL string for updating user information
const UpdateUserInputGQL = `
# Update user
input UpdateUserInput {
	forVersion: String!
	email: String!
	isAdmin: Boolean!
	isMember: Boolean!
	nickname: String!
	state: UserState!
}
`

// UpdateUserInput is the GQL structure for updating user information
type UpdateUserInput struct {
	// Fields received from the client
	// Email from client, remove before persisting, only persist EmailId
	ForVersion int32
	Email      string
	IsAdmin    bool
	IsMember   bool
	Nickname   string
	State      UserState

	// Extra fields persisted with the above
	IsSystem       bool
	UserId         string
	UpdateDateTime string
	AuthorUserId   string
	EmailId        string
	EventVersion   int32
}

// GetEventVersion returns the version of the event
func (r *UpdateUserInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *UpdateUserInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *UpdateUserInput) GetForVersion() int32 {
	return r.ForVersion
}

// AcceptInvitationInputGQL is the GQL string for accepting a property invitation
const AcceptInvitationInputGQL = `
# Accept an invitation to join a property
input AcceptInvitationInput {
	forVersion: String!
	accept: Boolean!
}
`

// AcceptInvitationInput is the GQL structure for accepting a property invitation
type AcceptInvitationInput struct {
	// Fields received from the client
	ForVersion int32
	Accept     bool

	// Extra fields persisted with the above
	UpdateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

// GetEventVersion returns the version of the event
func (r *AcceptInvitationInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *AcceptInvitationInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *AcceptInvitationInput) GetForVersion() int32 {
	return r.ForVersion
}

// UpdateSystemUserInputGQL is the GQL string for updating user information
const UpdateSystemUserInputGQL = `
# Update system user
input UpdateSystemUserInput {
	forVersion: Int!
	email: String!
	nickname: String!
}
`

// UpdateSystemUserInput is the GQL structure for updating user information
type UpdateSystemUserInput struct {
	// Fields received from the client
	// Email from client, remove before persisting, only persist EmailId
	ForVersion int32
	Email      string
	Nickname   string

	// Extra fields persisted with the above
	UserID         string
	UpdateDateTime string
	AuthorUserID   string
	EmailID        string
	EventVersion   int32
}

// GetEventVersion returns the version of the event
func (r *UpdateSystemUserInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *UpdateSystemUserInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *UpdateSystemUserInput) GetForVersion() int32 {
	return r.ForVersion
}
