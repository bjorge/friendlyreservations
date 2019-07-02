package models

// events sent from the browser (persisted to datastore)
type NewVersionEvent struct {
	Version      int
	EventVersion int32
}

func (r *NewVersionEvent) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *NewVersionEvent) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
)

type NewPropertyInput struct {
	// Fields received from the client
	PropertyName    string
	Currency        Currency
	MemberRate      int32
	AllowNonMembers bool
	NonMemberRate   int32
	IsMember        bool
	NickName        string
	Timezone        string

	// Extra fields persisted with the above
	// BUG(bjorge): remove propertyId from here - misleading
	//   since after export/import it will be different
	PropertyId     string
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

func (r *NewPropertyInput) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *NewPropertyInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

type DeletePropertyInput struct {
	// Fields received from the client
	PropertyId string
}

type ContentName string

const (
	ADMIN_HOME  ContentName = "ADMIN_HOME"
	MEMBER_HOME ContentName = "MEMBER_HOME"
)
