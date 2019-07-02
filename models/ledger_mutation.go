package models

// LedgerMutationGQL is the GQL string for updating user balance
const LedgerMutationGQL = `

# Information to create a new payment.
input UpdateBalanceInput {
	forVersion: Int!
	updateForUserId: String!
	amount: Int!
	increase: Boolean!
	description: String!
}
`

// UpdateBalanceInput is the go struct corresponding to the input GQL
type UpdateBalanceInput struct {
	// Fields received from the client
	ForVersion      int32
	UpdateForUserId string
	Amount          int32
	Description     string
	Increase        bool

	// Extra fields persisted with the above
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
	PaymentId      string
}

// GetEventVersion called to get input version
func (r *UpdateBalanceInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist layer to set the version
func (r *UpdateBalanceInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called to check for duplicate requests
func (r *UpdateBalanceInput) GetForVersion() int32 {
	return r.ForVersion
}
