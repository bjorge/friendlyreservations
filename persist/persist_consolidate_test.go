package persist

import (
	"context"
	"encoding/gob"
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/appengine/aetest"
)

func TestConsolidateDatastore(t *testing.T) {
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

	defaultsTest(ctx, t, persistedVersionedEvents)
	numRecordsTest(ctx, t, persistedVersionedEvents)
	maxSizeTest(ctx, t, persistedVersionedEvents)
	compressTest(ctx, t, persistedVersionedEvents)
}

func TestConsolidateUnitTest(t *testing.T) {
	var ctx context.Context

	var persistedVersionedEvents PersistedVersionedEvents
	persistedVersionedEvents = NewPersistedVersionedEvents(true)
	defaultsTest(ctx, t, persistedVersionedEvents)
	numRecordsTest(ctx, t, persistedVersionedEvents)
	maxSizeTest(ctx, t, persistedVersionedEvents)

}

func defaultsTest(ctx context.Context, t *testing.T, persistedVersionedEvents PersistedVersionedEvents) {

	_, events, numRecords := consolidateTest(ctx, t, persistedVersionedEvents)

	if _, ok := persistedVersionedEvents.(*unitTestImpl); ok {
		// the unit test just stores events into an array
		if numRecords != len(events) {
			t.Fatalf("Unit test expected %+v records but instead got %+v", len(events), numRecords)
		}
	} else if _, ok := persistedVersionedEvents.(*dataStoreImpl); ok {
		// the  real database code will consolidate events into less records for quicker access
		if numRecords != 5 {
			t.Fatalf("Persist test expected %+v records but instead got %+v", 5, numRecords)
		}
	}
}

func numRecordsTest(ctx context.Context, t *testing.T, persistedVersionedEvents PersistedVersionedEvents) {

	consolidateCompress = false
	consolidateNumRecords = 20

	_, events, numRecords := consolidateTest(ctx, t, persistedVersionedEvents)

	if _, ok := persistedVersionedEvents.(*unitTestImpl); ok {
		// the unit test just stores events into an array
		if numRecords != len(events) {
			t.Fatalf("Unit test expected %+v records but instead got %+v", len(events), numRecords)
		}
	} else if _, ok := persistedVersionedEvents.(*dataStoreImpl); ok {
		// the  real database code will consolidate events into less records for quicker access
		if numRecords != 6 {
			t.Fatalf("Persist test expected %+v records but instead got %+v", 6, numRecords)
		}
	}
}

func maxSizeTest(ctx context.Context, t *testing.T, persistedVersionedEvents PersistedVersionedEvents) {

	consolidateNumRecords = 10
	consolidateMaxSize = 500
	consolidateCompress = false

	_, events, numRecords := consolidateTest(ctx, t, persistedVersionedEvents)

	if _, ok := persistedVersionedEvents.(*unitTestImpl); ok {
		// the unit test just stores events into an array
		if numRecords != len(events) {
			t.Fatalf("Unit test expected %+v records but instead got %+v", len(events), numRecords)
		}
	} else if _, ok := persistedVersionedEvents.(*dataStoreImpl); ok {
		// the  real database code will consolidate events into less records for quicker access
		if numRecords != 11 {
			t.Fatalf("Persist test expected %+v records but instead got %+v", 11, numRecords)
		}
	}
}

func compressTest(ctx context.Context, t *testing.T, persistedVersionedEvents PersistedVersionedEvents) {

	consolidateNumRecords = 10
	consolidateMaxSize = 500
	consolidateCompress = true

	_, events, numRecords := consolidateTest(ctx, t, persistedVersionedEvents)

	if _, ok := persistedVersionedEvents.(*unitTestImpl); ok {
		// the unit test just stores events into an array
		if numRecords != len(events) {
			t.Fatalf("Unit test expected %+v records but instead got %+v", len(events), numRecords)
		}
	} else if _, ok := persistedVersionedEvents.(*dataStoreImpl); ok {
		// the  real database code will consolidate events into less records for quicker access
		if numRecords != 2 {
			t.Fatalf("Persist test expected %+v records but instead got %+v", 2, numRecords)
		}
	}
}

func consolidateTest(ctx context.Context, t *testing.T, persistedVersionedEvents PersistedVersionedEvents) (string, []VersionedEvent, int) {

	gob.Register(&testEvent1{})
	gob.Register(&testEvent2{})
	gob.Register(&testEvent3{})

	event1 := &testEvent1{Value: 1}
	event2 := &testEvent2{Value: "one"}

	firstEvents := []VersionedEvent{event1, event2}

	propertyID := uuid.Must(uuid.NewV4()).String()
	nextEventTransactionKey, err := persistedVersionedEvents.CreateProperty(ctx, propertyID, firstEvents, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	// ids so far are 0 and 1, so next one is 2
	nextID := 2
	// add a bunch of events
	for i := 0; i < 100; i++ {
		nextEventTransactionKey, err = persistedVersionedEvents.NewPropertyEvents(ctx, propertyID, nextEventTransactionKey, []VersionedEvent{&testEvent3{Value: true}}, false)
		if err != nil {
			t.Fatal(err)
		}
		nextID++
	}

	if nextEventTransactionKey != nextID {
		t.Fatalf("expected %+v but got %+v", nextID, nextEventTransactionKey)
	}

	// get all events for the property
	events, err := persistedVersionedEvents.GetEvents(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != nextID {
		t.Fatal(fmt.Errorf("expected length to be %+v but instead got %+v", nextID, len(events)))
	}

	numRecords, err := persistedVersionedEvents.NumRecords(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	return propertyID, events, numRecords
}
