package persist

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"testing"

	"github.com/bjorge/friendlyreservations/platform"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/appengine/aetest"
)

type testEvent1 struct {
	Value        int
	VersionValue int
}

func (r *testEvent1) GetEventVersion() int {
	return r.VersionValue
}

func (r *testEvent1) SetEventVersion(Version int) {
	r.VersionValue = Version
}

type testEvent2 struct {
	Value        string
	VersionValue int
}

func (r *testEvent2) GetEventVersion() int {
	return r.VersionValue
}

func (r *testEvent2) SetEventVersion(Version int) {
	r.VersionValue = Version
}

type testEvent3 struct {
	Value        bool
	VersionValue int
}

func (r *testEvent3) GetEventVersion() int {
	return r.VersionValue
}

func (r *testEvent3) SetEventVersion(Version int) {
	r.VersionValue = Version
}

func TestCreateEmailDataStore(t *testing.T) {
	testCreateEmail(t, false)
}

func TestCreateEmailUnitTest(t *testing.T) {
	testCreateEmail(t, true)
}

func testCreateEmail(t *testing.T, unitTest bool) {

	var ctx context.Context
	var err error
	var done func()
	var persistedEmailStore platform.PersistedEmailStore
	if unitTest {
		persistedEmailStore = NewPersistedEmailStore(true)
	} else {
		ctx, done, err = aetest.NewContext()
		if err != nil {
			t.Fatal(err)
		}
		defer done()
		persistedEmailStore = NewPersistedEmailStore(false)
	}

	propertyID := "id123"
	email := "test@testing.com"

	record1, err := persistedEmailStore.CreateEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	if record1.Email != email {
		t.Fatal(errors.New("email in created record does not match"))
	}

	record2, err := persistedEmailStore.GetEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}
	if record1.EmailID != record2.EmailID {
		t.Fatal(errors.New("email id does not match"))
	}

	if record1.Email != record2.Email {
		t.Fatal(errors.New("emails do not match"))
	}

	ids, err := persistedEmailStore.GetPropertiesByEmail(ctx, email)

	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 1 {
		t.Fatal(errors.New("search for properties by email failed"))
	}

	if ids[0].PropertyID != propertyID {
		t.Fatal(errors.New("wrong propertyId from GetPropertiesByEmail"))
	}

	emailMap, err := persistedEmailStore.GetEmailMap(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	if emailMap[ids[0].EmailID].Email != email {
		t.Fatal(errors.New("email map does not work"))
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, email); !*exists {
		t.Fatal(errors.New("email should exist"))
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, "howdy!"); *exists {
		t.Fatal(errors.New("email should not exist"))
	}

}

func TestDeleteEmailUnitTest(t *testing.T) {
	testDeleteEmail(t, true)
}

func TestDeleteEmailDataStore(t *testing.T) {
	testDeleteEmail(t, false)
}

func testDeleteEmail(t *testing.T, unitTest bool) {

	var ctx context.Context
	var done func()
	var err error
	var persistedEmailStore platform.PersistedEmailStore
	if unitTest {
		persistedEmailStore = NewPersistedEmailStore(true)
	} else {
		ctx, done, err = aetest.NewContext()
		if err != nil {
			t.Fatal(err)
		}
		defer done()
		persistedEmailStore = NewPersistedEmailStore(false)
	}

	propertyID := "id123"
	email := "test@testing.com"

	if _, err = persistedEmailStore.CreateEmail(ctx, propertyID, email); err != nil {
		t.Fatal(err)
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, email); !*exists {
		t.Fatal(errors.New("email should exist"))
	}

	if err = persistedEmailStore.DeleteEmails(ctx, propertyID); err != nil {
		t.Fatal(err)
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, email); *exists {
		t.Fatal(errors.New("email should not exist"))
	}

}

func TestPersistEvents(t *testing.T) {
	persistEventsFunc(t, false)
}

func TestPersistEventsUnitTest(t *testing.T) {
	persistEventsFunc(t, true)
}

func persistEventsFunc(t *testing.T, unitTest bool) {

	var ctx context.Context
	var err error
	var done func()
	var persistedVersionedEvents platform.PersistedVersionedEvents
	var persistedPropertyList platform.PersistedPropertyList
	if unitTest {
		persistedVersionedEvents = NewPersistedVersionedEvents(true)
		persistedPropertyList = NewPersistedPropertyList(true)
	} else {
		ctx, done, err = aetest.NewContext()
		if err != nil {
			t.Fatal(err)
		}
		defer done()
		persistedVersionedEvents = NewPersistedVersionedEvents(false)
		persistedPropertyList = NewPersistedPropertyList(false)
	}

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

func TestDeletePropertyUnitTest(t *testing.T) {
	persistDeleteFunc(t, true)
}

func TestDeleteProperty(t *testing.T) {
	persistDeleteFunc(t, false)
}

func persistDeleteFunc(t *testing.T, unitTest bool) {

	var ctx context.Context
	var err error
	var done func()
	var persistedVersionedEvents platform.PersistedVersionedEvents
	var persistedPropertyList platform.PersistedPropertyList
	if unitTest {
		persistedVersionedEvents = NewPersistedVersionedEvents(true)
		persistedPropertyList = NewPersistedPropertyList(true)
	} else {
		ctx, done, err = aetest.NewContext()
		if err != nil {
			t.Fatal(err)
		}
		defer done()
		persistedVersionedEvents = NewPersistedVersionedEvents(false)
		persistedPropertyList = NewPersistedPropertyList(false)
	}

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

func TestRestoreEmailDataStore(t *testing.T) {
	testRestoreEmail(t, false)
}

func TestRestoreEmailUnitTest(t *testing.T) {
	testRestoreEmail(t, true)
}

func testRestoreEmail(t *testing.T, unitTest bool) {
	var ctx context.Context
	var err error
	var done func()
	var persistedEmailStore platform.PersistedEmailStore
	if unitTest {
		persistedEmailStore = NewPersistedEmailStore(true)
	} else {
		ctx, done, err = aetest.NewContext()
		if err != nil {
			t.Fatal(err)
		}
		defer done()
		persistedEmailStore = NewPersistedEmailStore(false)
	}

	propertyID := "id1234"
	email := "test1@testing.com"

	_, err = persistedEmailStore.CreateEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	record1, err := persistedEmailStore.GetEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	if record1.Email != email {
		t.Fatal(errors.New("email does not match"))
	}

	err = persistedEmailStore.DeleteEmails(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	emailBackup := make(map[string]string)
	emailBackup[record1.EmailID] = record1.Email
	err = persistedEmailStore.RestoreEmails(ctx, propertyID, emailBackup)
	if err != nil {
		t.Fatal(err)
	}

	record2, err := persistedEmailStore.GetEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	if record2.EmailID != record1.EmailID {
		t.Fatal(errors.New("email id does not match"))
	}

	if record2.Email != record1.Email {
		t.Fatal(errors.New("email does not match"))
	}
}
