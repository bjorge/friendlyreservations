package localplatform

import (
	"testing"

	"github.com/bjorge/friendlyreservations/platform_testing"
)

func TestEmail(t *testing.T) {
	platformtesting.TestCreateEmail(nil, t, NewPersistedEmailStore())
	platformtesting.TestDeleteEmail(nil, t, NewPersistedEmailStore())
	platformtesting.TestRestoreEmail(nil, t, NewPersistedEmailStore())

}
