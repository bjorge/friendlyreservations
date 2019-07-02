package models

// NewReservationInputGQL is the GQL string for creating a reservation
const NewReservationInputGQL = `
# Information to create a new reservation.
input NewReservationInput {
	# the version of the property being updated
	forVersion: Int!
	reservedForUserId: String!
	startDate: String!
	endDate: String!
	member: Boolean!
	nonMemberName: String
	nonMemberInfo: String
	adminRequest: Boolean!
}
`

type NewReservationInput struct {
	// Fields received from the client
	ForVersion        int32
	ReservedForUserId string
	StartDate         string
	EndDate           string
	Member            bool
	NonMemberName     *string
	NonMemberInfo     *string
	AdminRequest      bool

	// Extra fields persisted with the above
	Rate           []DailyRate
	ReservationId  string
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

// GetEventVersion returns the version of the mutation event
func (r *NewReservationInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *NewReservationInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *NewReservationInput) GetForVersion() int32 {
	return r.ForVersion
}

// CancelReservationInput is called to cancel a reservation
type CancelReservationInput struct {
	// Fields received from the client
	ForVersion    int32
	ReservationId string
	AdminRequest  bool

	// Extra fields persisted with the above
	CreateDateTime    string
	ReservedForUserId string
	AuthorUserId      string
	EventVersion      int32
}

// GetEventVersion returns the version of the mutation event
func (r *CancelReservationInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *CancelReservationInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *CancelReservationInput) GetForVersion() int32 {
	return r.ForVersion
}
