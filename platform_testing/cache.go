package platformtesting

import (
	"context"
	"testing"

	"github.com/bjorge/friendlyreservations/platform"
)

// TestCache is called by a platform implementor to test the platform cache
func TestCache(ctx context.Context, t *testing.T, persistedVersionedEvents platform.PersistedVersionedEvents) {

	key := "theKey"
	value := "theValue"
	propertyID := "x"
	version := 1
	//data := make(map[string][]byte)
	//data[key] = []byte(value)

	err := persistedVersionedEvents.CacheWrite(ctx, propertyID, version, key, []byte(value))
	if err != nil {
		t.Fatal(err)
	}

	readVersion, readData, err := persistedVersionedEvents.CacheRead(ctx, propertyID, []string{key})

	if readVersion != version {
		t.Fatalf("cache version expected %+v but got %+v", version, readVersion)
	}

	if readData == nil {
		t.Fatalf("cache data is nil")
	}

	if readValue, ok := readData[key]; !ok {
		t.Fatalf("data does not contain key value")
	} else {
		readString := string(readValue)

		if readString != value {
			t.Fatalf("cache value expected %+v but got %+v", value, readValue)
		}
	}

}
