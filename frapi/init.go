package frapi

import (
	"encoding/gob"

	"github.com/bjorge/friendlyreservations/cookies"
	"github.com/bjorge/friendlyreservations/logger"
	"github.com/bjorge/friendlyreservations/platform"
)

// Logger is the logger for the platform implementation
var Logger = logger.New()

// PersistedEmailStore manages the persisted user emails
var PersistedEmailStore platform.PersistedEmailStore

// PersistedVersionedEvents is the list of persisted events
var PersistedVersionedEvents platform.PersistedVersionedEvents

// PersistedPropertyList is the list of persisted properties
var PersistedPropertyList platform.PersistedPropertyList

// FrapiCookies contains helper functions for setting and getting cookies
var FrapiCookies *cookies.AuthCookies

// init intializes the data structures for gob serialization
func init() {
	FrapiCookies = cookies.NewCookies()

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
