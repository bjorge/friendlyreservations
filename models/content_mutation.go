package models

// NewContentInputGQL is the GQL string for updating content
const NewContentInputGQL = `
# Information to create new content.
input NewContentInput {
	forVersion: String!
	name: ContentName!
	template: String!
	comment: String!
}
`

// NewContentInput is the gql input for updating content
type NewContentInput struct {
	// Fields received from the client
	ForVersion int32
	Name       ContentName
	Template   string
	Comment    string

	// Extra fields persisted with the above
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

// GetEventVersion returns the version of the contents update event
func (r *NewContentInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *NewContentInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *NewContentInput) GetForVersion() int32 {
	return r.ForVersion
}
