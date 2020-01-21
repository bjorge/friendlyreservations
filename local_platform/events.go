package localplatform

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/bjorge/friendlyreservations/platform"
)

type cacheRecord struct {
	Version    int
	Compressed bool
	Value      []byte
}

// Event is the []byte gob-encoded array of events
type Event struct {
	ID    int
	Value interface{}
}

// PersistedPropertyEvents is the record that holds some property events
type PersistedPropertyEvents struct {
	Events     []byte
	First      int
	Last       int
	Compressed bool
	Type       int
}

func cacheRecordFromGob(gobData []byte) (*cacheRecord, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(gobData))
	record := &cacheRecord{}
	err := dec.Decode(&record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func gobFromCacheRecord(record *cacheRecord) ([]byte, error) {
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(record)
	if err != nil {
		return nil, err
	}
	return stream.Bytes(), err
}

const cacheKeyDelimiter = "_"

type unitTestImpl struct {
	Events []byte
	Cache  map[string]cacheRecord
}

// NewPersistedVersionedEvents is the factory method to create a persisted events store
func NewPersistedVersionedEvents() platform.PersistedVersionedEvents {
	return &unitTestImpl{Events: []byte{}, Cache: make(map[string]cacheRecord)}
}

func generateCacheKey(propertyID string, key string) string {
	cacheKey := fmt.Sprintf("%s%s%s", propertyID, cacheKeyDelimiter, key)
	return cacheKey
}

const versionedEventsCacheKeyName = "VersionedEvents"
const currentVersionCacheKeyName = "CurrentVersion"

func (r *unitTestImpl) CacheWrite(ctx context.Context, propertyID string, version int, key string, value []byte) error {
	cacheKey := generateCacheKey(propertyID, key)
	record := cacheRecord{Version: version, Value: value}
	r.Cache[cacheKey] = record

	return nil
}

func (r *unitTestImpl) CacheDelete(ctx context.Context, propertyID string, key string) error {
	cacheKey := generateCacheKey(propertyID, key)
	delete(r.Cache, cacheKey)

	return nil
}

func (r *unitTestImpl) CacheRead(ctx context.Context, propertyID string, keys []string) (int, map[string][]byte, error) {
	version := 0
	data := make(map[int]map[string][]byte)

	// default to empty set for default version 0
	data[version] = make(map[string][]byte)

	for _, key := range keys {
		cacheKey := generateCacheKey(propertyID, key)

		record, ok := r.Cache[cacheKey]
		if !ok {
			continue
		}

		if _, ok := data[record.Version]; !ok {
			data[record.Version] = make(map[string][]byte)
		}

		data[record.Version][key] = record.Value

		if version < record.Version {
			version = record.Version
		}
	}

	return version, data[version], nil
}

func (r *unitTestImpl) CreateProperty(ctx context.Context, propertyID string, events []platform.VersionedEvent, persistedPropertyList platform.PersistedPropertyList, nextPropertyListIndex int) (int, error) {
	// BUG(bjorge): check that transactionIndex is correct
	if persistedPropertyList != nil {
		err := persistedPropertyList.CreateProperty(ctx, propertyID, nextPropertyListIndex)
		if err != nil {
			return 0, err
		}
	}
	records := []platform.VersionedEvent{}
	for index, event := range events {
		event.SetEventVersion(index)
		records = append(records, event)
	}
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(records)
	if err != nil {
		return -1, err
	}
	r.Events = stream.Bytes()
	return len(events), nil
}

func (r *unitTestImpl) DeleteProperty(ctx context.Context, propertyID string, persistedPropertyList platform.PersistedPropertyList) error {
	if persistedPropertyList != nil {
		err := persistedPropertyList.DeleteProperty(ctx, propertyID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *unitTestImpl) NewPropertyEvents(ctx context.Context, propertyID string, transactionKey int, events []platform.VersionedEvent, inTransaction bool) (int, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(r.Events))
	records := []platform.VersionedEvent{}
	err := dec.Decode(&records)
	if err != nil {
		return -1, err
	}

	length := len(records)
	if transactionKey != length {
		return 0, errors.New("NewPropertyEvents transactionIndex is wrong")
	}
	for index, event := range events {
		event.SetEventVersion(length + index)
		records = append(records, event)
	}
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err = en.Encode(records)
	if err != nil {
		return -1, err
	}
	r.Events = stream.Bytes()
	r.CacheWrite(ctx, propertyID, len(records)-1, currentVersionCacheKeyName, []byte{})
	return len(records), nil

}
func (r *unitTestImpl) GetNextEventID(ctx context.Context, propertyID string, inTransaction bool) (int, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(r.Events))
	records := []platform.VersionedEvent{}
	err := dec.Decode(&records)
	if err != nil {
		return -1, err
	}

	return len(records), nil
}
func (r *unitTestImpl) NumRecords(ctx context.Context, propertyID string) (int, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(r.Events))
	records := []platform.VersionedEvent{}
	err := dec.Decode(&records)
	if err != nil {
		return -1, err
	}

	return len(records), nil
}

func eventsFromGob(gobData []byte) ([]platform.VersionedEvent, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(gobData))
	events := []platform.VersionedEvent{}
	err := dec.Decode(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func gobFromEvents(events []platform.VersionedEvent) ([]byte, error) {
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(events)
	if err != nil {
		return nil, err
	}
	return stream.Bytes(), err
}

func (r *unitTestImpl) GetEvents(ctx context.Context, propertyID string) ([]platform.VersionedEvent, error) {

	// get events from "cache"
	_, readData, err := r.CacheRead(ctx, propertyID, []string{versionedEventsCacheKeyName, currentVersionCacheKeyName})
	if err != nil {
		logging.LogDebugf("Error reading from cache: %+v", err)
	} else {
		if _, ok := readData[currentVersionCacheKeyName]; ok {
			// good, we have the current version
			if readValue, ok := readData[versionedEventsCacheKeyName]; ok {
				// and we have events at the current version
				events, err := eventsFromGob(readValue)
				if err != nil {
					logging.LogDebugf("Error converting cached gob to events: %+v", err)
				}
				return events, nil
			}
		}

	}

	// get events from "db"
	events, err := eventsFromGob(r.Events)
	if err != nil {
		logging.LogDebugf("Error converting db gob to events: %+v", err)
		return nil, err
	}

	// ok great we got events, now write back to cache
	gobData, err := gobFromEvents(events)
	err = r.CacheWrite(ctx, propertyID, events[len(events)-1].GetEventVersion(), versionedEventsCacheKeyName, gobData)
	err = r.CacheWrite(ctx, propertyID, events[len(events)-1].GetEventVersion(), currentVersionCacheKeyName, []byte{})
	if err != nil {
		logging.LogDebugf("Error writing gob to cache: %+v", err)
	}
	return events, nil

}
