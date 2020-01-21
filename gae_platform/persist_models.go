package gaeplatform

// todo: use these
// type PropertyTransactionKey int
// type PropertyId string
// type EventTransactionKey int

type emptyStruct struct {
}

// THE MASTER TABLE OF PROPERTIES

var persistedPropertiesKind = "PERSISTED_PROPERTIES_KIND"

// PersistedProperties is the structure used to a property id record (the id is in the key)
type PersistedProperties struct {
	ParentKey        string `datastore:"k,noindex"`
	TransactionIndex int    `datastore:"i"` // indexed in order to get the last one
}

// THE TABLE OF EACH SET OF PROPERTY EVENTS
// type PersistedPropertyEventsRoot struct {
// }

// PersistedPropertyEvents is the record that holds some property events
type PersistedPropertyEvents struct {
	Events     []byte `datastore:"e,noindex"` // make sure smaller than 1MB (datastore limit)
	First      int    `datastore:"f,noindex"`
	Last       int    `datastore:"l,noindex"`
	Compressed bool   `datastore:"c,noindex"`
	Type       int    `datastore:"t,noindex"`
}

// Event is the []byte gob-encoded array of events
type Event struct {
	ID    int
	Value interface{}
}
