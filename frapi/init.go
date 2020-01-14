package frapi

import (
	"encoding/gob"
	"fmt"

	"github.com/bjorge/friendlyreservations/config"
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

var destinationURI string

// init intializes the data structures for gob serialization
func init() {
	FrapiCookies = cookies.NewCookies()

	destinationURI = config.GetConfig("PLATFORM_DESTINATION_URI")
	if destinationURI == "" {
		panic(fmt.Errorf("PLATFORM_DESTINATION_URI is not set"))
	}

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
