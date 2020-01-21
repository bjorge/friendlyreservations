package gaeplatform

import (
	"testing"

	"github.com/bjorge/friendlyreservations/platform_testing"
	"google.golang.org/appengine/aetest"
)

func TestEmail(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	platformtesting.TestCreateEmail(ctx, t, NewPersistedEmailStore())
	platformtesting.TestDeleteEmail(ctx, t, NewPersistedEmailStore())
	platformtesting.TestRestoreEmail(ctx, t, NewPersistedEmailStore())

}
