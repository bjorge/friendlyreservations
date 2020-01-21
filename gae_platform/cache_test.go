package gaeplatform

import (
	"testing"

	"github.com/bjorge/friendlyreservations/platform_testing"
	"google.golang.org/appengine/aetest"
)

func TestCacheDatastore(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	platformtesting.TestCache(ctx, t, NewPersistedVersionedEvents())
}
