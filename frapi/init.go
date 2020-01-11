package frapi

import (
	"encoding/gob"

	"github.com/bjorge/friendlyreservations/persist"
	"github.com/bjorge/friendlyreservations/platform"
)

// Logger is the logger for the platform implementation
var Logger platform.Logger

var persistedEmailStore = persist.NewPersistedEmailStore(false)
var persistedVersionedEvents = persist.NewPersistedVersionedEvents(false)
var persistedPropertyList = persist.NewPersistedPropertyList(false)

// init intializes the data structures for gob serialization
func init() {
	gob.Register(&PropertyExport{})
	gob.Register(&UserRollup{})
	gob.Register(&LedgerRollup{})
	gob.Register(&NotificationRollup{})
	gob.Register(&ReservationRollup{})
	gob.Register(&SettingsRollup{})
	gob.Register(&RestrictionRollup{})
	gob.Register(&ContentRollup{})
	gob.Register(&MembershipRollupRecord{})
}
