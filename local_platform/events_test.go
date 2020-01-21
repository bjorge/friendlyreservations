package localplatform

import (
	"testing"

	"github.com/bjorge/friendlyreservations/platform_testing"
)

func TestEvents(t *testing.T) {
	platformtesting.TestEventsCreate(nil, t, NewPersistedVersionedEvents(), NewPersistedPropertyList())
	platformtesting.TestEventsDelete(nil, t, NewPersistedVersionedEvents(), NewPersistedPropertyList())
}
