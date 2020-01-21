package localplatform

import (
	"testing"

	"github.com/bjorge/friendlyreservations/platform_testing"
)

func TestCache(t *testing.T) {
	platformtesting.TestCache(nil, t, NewPersistedVersionedEvents())
}
