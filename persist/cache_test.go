package persist

import (
	"context"
	"testing"

	"google.golang.org/appengine/aetest"
)

func TestCacheDatastore(t *testing.T) {
	var ctx context.Context
	var err error
	var done func()
	var persistedVersionedEvents PersistedVersionedEvents

	ctx, done, err = aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	persistedVersionedEvents = NewPersistedVersionedEvents(false)

	cacheTest(ctx, t, persistedVersionedEvents)
}

func TestCacheUnitTest(t *testing.T) {
	var ctx context.Context

	var persistedVersionedEvents PersistedVersionedEvents
	persistedVersionedEvents = NewPersistedVersionedEvents(true)
	cacheTest(ctx, t, persistedVersionedEvents)

}

func cacheTest(ctx context.Context, t *testing.T, persistedVersionedEvents PersistedVersionedEvents) {

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
