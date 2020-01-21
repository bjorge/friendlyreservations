package platformtesting

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"testing"

	"github.com/bjorge/friendlyreservations/platform"
	uuid "github.com/satori/go.uuid"
)

// TestEventsCreate is called by the platform implementation testing code
func TestEventsCreate(ctx context.Context, t *testing.T, persistedVersionedEvents platform.PersistedVersionedEvents, persistedPropertyList platform.PersistedPropertyList) {

	gob.Register(&testEvent1{})
	gob.Register(&testEvent2{})
	gob.Register(&testEvent3{})

	version := &testEvent1{Value: 1}
	property := &testEvent2{Value: "one"}

	firstEvents := []platform.VersionedEvent{version, property}

	nextPropertyTransactionKey, err := persistedPropertyList.GetNextVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}

	propertyID := uuid.Must(uuid.NewV4()).String()
	nextEventTransactionKey, err := persistedVersionedEvents.CreateProperty(ctx, propertyID, firstEvents, persistedPropertyList, nextPropertyTransactionKey)
	if err != nil {
		t.Fatal(err)
	}

	// now the next property to create should have a different transaction id
	newPropertyTransactionKey, err := persistedPropertyList.GetNextVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if nextPropertyTransactionKey == newPropertyTransactionKey {
		t.Fatal(errors.New("new property transaction key should be different"))
	}

	version2 := &testEvent3{Value: true}
	ifaces := []platform.VersionedEvent{version2}

	nextTransactionID, err := persistedVersionedEvents.NewPropertyEvents(ctx, propertyID, nextEventTransactionKey, ifaces, false)

	if nextTransactionID != 3 {
		t.Fatalf("expected 3 but got %+v", nextTransactionID)
		t.Fatal(errors.New("expected 3 events"))
	}

	// get all the properties (should be only 1 now)
	ids, err := persistedPropertyList.GetProperties(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 1 {
		t.Fatal(errors.New("Expected only a single property"))
	}

	// get all events for the property
	events, err := persistedVersionedEvents.GetEvents(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != 3 {
		t.Fatal(fmt.Errorf("expected length to be 3 but instead got %+v", len(events)))
	}

	if events[0].GetEventVersion() != 0 {
		t.Fatal(fmt.Errorf("expected version to be 0, but it is: %+v", events[0].GetEventVersion()))
	}

	if events[2].GetEventVersion() != 2 {
		t.Fatal(fmt.Errorf("expected version to be 2, but it is: %+v", events[2].GetEventVersion()))
	}

	// get only last event for the property
	nextID, err := persistedVersionedEvents.GetNextEventID(ctx, propertyID, false)
	if err != nil {
		t.Fatal(err)
	}

	if nextID != 3 {
		t.Fatal(errors.New("Expected next event id to be 3"))
	}

	versionEvent, ok := events[0].(*testEvent1)

	if !ok {
		t.Fatal(errors.New("Expected first event to be version event"))
	}

	if versionEvent.Value != 1 {
		t.Fatal(errors.New("Expected version to be 1"))
	}

	// get all events for the property again (testing cache)
	events1, err := persistedVersionedEvents.GetEvents(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != len(events1) {
		t.Fatalf("cache error, expected %+v events but instead got %+v", len(events), len(events1))
	}
}

// TestEventsDelete is called by the platform implementation testing code
func TestEventsDelete(ctx context.Context, t *testing.T, persistedVersionedEvents platform.PersistedVersionedEvents, persistedPropertyList platform.PersistedPropertyList) {

	gob.Register(&testEvent1{})
	gob.Register(&testEvent2{})
	gob.Register(&testEvent3{})

	version := &testEvent1{Value: 1}
	property := &testEvent2{Value: "one"}

	firstEvents := []platform.VersionedEvent{version, property}

	nextPropertyTransactionKey, err := persistedPropertyList.GetNextVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}

	propertyID := uuid.Must(uuid.NewV4()).String()
	_, err = persistedVersionedEvents.CreateProperty(ctx, propertyID, firstEvents, persistedPropertyList, nextPropertyTransactionKey)
	if err != nil {
		t.Fatal(err)
	}

	_, err = persistedVersionedEvents.GetEvents(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	err = persistedVersionedEvents.DeleteProperty(ctx, propertyID, persistedPropertyList)
	if err != nil {
		t.Fatal(err)
	}

}
