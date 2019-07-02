package persist

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"

	"google.golang.org/appengine/log"
)

// VersionedEvent must be implemented by all event objects that are persisted
type VersionedEvent interface {
	GetEventVersion() int
	SetEventVersion(Version int)
}

// PersistedVersionedEvents is the interface for managing a property
type PersistedVersionedEvents interface {
	CreateProperty(ctx context.Context, propertyID string, events []VersionedEvent, persistedPropertyList PersistedPropertyList, nextPropertyListIndex int) (int, error)
	NewPropertyEvents(ctx context.Context, propertyID string, transactionKey int, events []VersionedEvent, inTransaction bool) (int, error)
	GetNextEventID(ctx context.Context, propertyID string, inTransaction bool) (int, error)
	GetEvents(ctx context.Context, propertyID string) ([]VersionedEvent, error)
	DeleteProperty(ctx context.Context, propertyID string, persistedPropertyList PersistedPropertyList) error
	NumRecords(ctx context.Context, propertyID string) (int, error)
	CacheWrite(ctx context.Context, propertyID string, version int, key string, value []byte) error
	CacheRead(ctx context.Context, propertyID string, keys []string) (int, map[string][]byte, error)
	CacheDelete(ctx context.Context, propertyID string, key string) error
}

type cacheRecord struct {
	Version    int
	Compressed bool
	Value      []byte
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

type dataStoreImpl struct{}
type unitTestImpl struct {
	Events []byte
	Cache  map[string]cacheRecord
}

// NewPersistedVersionedEvents is the factory method to create a persisted events store
func NewPersistedVersionedEvents(unitTest bool) PersistedVersionedEvents {
	if unitTest {
		return &unitTestImpl{Events: []byte{}, Cache: make(map[string]cacheRecord)}
	}
	return &dataStoreImpl{}
}

func generateCacheKey(propertyID string, key string) string {
	return fmt.Sprintf("%s%s%s", propertyID, cacheKeyDelimiter, key)
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

func (r *unitTestImpl) CreateProperty(ctx context.Context, propertyID string, events []VersionedEvent, persistedPropertyList PersistedPropertyList, nextPropertyListIndex int) (int, error) {
	// BUG(bjorge): check that transactionIndex is correct
	if persistedPropertyList != nil {
		err := persistedPropertyList.CreateProperty(ctx, propertyID, nextPropertyListIndex)
		if err != nil {
			return 0, err
		}
	}
	records := []VersionedEvent{}
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

func (r *unitTestImpl) DeleteProperty(ctx context.Context, propertyID string, persistedPropertyList PersistedPropertyList) error {
	if persistedPropertyList != nil {
		err := persistedPropertyList.DeleteProperty(ctx, propertyID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *unitTestImpl) NewPropertyEvents(ctx context.Context, propertyID string, transactionKey int, events []VersionedEvent, inTransaction bool) (int, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(r.Events))
	records := []VersionedEvent{}
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
	records := []VersionedEvent{}
	err := dec.Decode(&records)
	if err != nil {
		return -1, err
	}

	return len(records), nil
}
func (r *unitTestImpl) NumRecords(ctx context.Context, propertyID string) (int, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(r.Events))
	records := []VersionedEvent{}
	err := dec.Decode(&records)
	if err != nil {
		return -1, err
	}

	return len(records), nil
}

func eventsFromGob(gobData []byte) ([]VersionedEvent, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(gobData))
	events := []VersionedEvent{}
	err := dec.Decode(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func gobFromEvents(events []VersionedEvent) ([]byte, error) {
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(events)
	if err != nil {
		return nil, err
	}
	return stream.Bytes(), err
}

func (r *unitTestImpl) GetEvents(ctx context.Context, propertyID string) ([]VersionedEvent, error) {

	// get events from "cache"
	_, readData, err := r.CacheRead(ctx, propertyID, []string{versionedEventsCacheKeyName, currentVersionCacheKeyName})
	if err != nil {
		log.Debugf(ctx, "Error reading from cache: %+v", err)
	} else {
		if _, ok := readData[currentVersionCacheKeyName]; ok {
			// good, we have the current version
			if readValue, ok := readData[versionedEventsCacheKeyName]; ok {
				// and we have events at the current version
				events, err := eventsFromGob(readValue)
				if err != nil {
					log.Debugf(ctx, "Error converting cached gob to events: %+v", err)
				}
				return events, nil
			}
		}

	}

	// get events from "db"
	events, err := eventsFromGob(r.Events)
	if err != nil {
		log.Debugf(ctx, "Error converting db gob to events: %+v", err)
		return nil, err
	}

	// ok great we got events, now write back to cache
	gobData, err := gobFromEvents(events)
	err = r.CacheWrite(ctx, propertyID, events[len(events)-1].GetEventVersion(), versionedEventsCacheKeyName, gobData)
	err = r.CacheWrite(ctx, propertyID, events[len(events)-1].GetEventVersion(), currentVersionCacheKeyName, []byte{})
	if err != nil {
		log.Debugf(ctx, "Error writing gob to cache: %+v", err)
	}
	return events, nil

}

func (r *dataStoreImpl) CacheWrite(ctx context.Context, propertyID string, version int, key string, value []byte) error {
	compressed := false
	// appengine docs show a drop off in throughput when data > 1Kb, so compress if > 1Kb
	if len(value) > 1024 {
		stream := &bytes.Buffer{}
		err := func() error {
			zlibWriter := zlib.NewWriter(stream)
			defer zlibWriter.Close()
			en := gob.NewEncoder(zlibWriter)
			err := en.Encode(value)
			return err
		}()
		if err != nil {
			return err
		}
		value = stream.Bytes()
		compressed = true
	}

	// if size won't fit into memcache even after compression, then just forget it for now
	if len(value) > (1048576 - 104857) {
		return fmt.Errorf("cannot memcache events for propertyID %+v, too big", propertyID)
	}

	cacheKey := generateCacheKey(propertyID, key)
	record := cacheRecord{Version: version, Value: value, Compressed: compressed}
	log.Debugf(ctx, "Write to cache propertyID: %+v, version: %+v, key: %+v, size %+v", propertyID, version, cacheKey, len(value))

	gobData, err := gobFromCacheRecord(&record)
	if err != nil {
		return err
	}
	err = memcache.Set(ctx, &memcache.Item{Key: cacheKey, Value: gobData, Expiration: memcacheDuration})
	return err
}

func (r *dataStoreImpl) CacheDelete(ctx context.Context, propertyID string, key string) error {
	cacheKey := generateCacheKey(propertyID, key)
	err := memcache.Delete(ctx, cacheKey)
	return err
}

func (r *dataStoreImpl) CacheRead(ctx context.Context, propertyID string, keys []string) (int, map[string][]byte, error) {
	cacheKeys := []string{}
	for _, key := range keys {
		cacheKey := generateCacheKey(propertyID, key)
		cacheKeys = append(cacheKeys, cacheKey)
		//log.Debugf(ctx, "Attempt Read from cache propertyID: %+v, key: %+v", propertyID, cacheKey)

	}

	items, err := memcache.GetMulti(ctx, cacheKeys)
	if err != nil {
		return 0, nil, err
	}

	data := make(map[int]map[string][]byte)

	// default to empty set for default version 0
	version := 0
	data[version] = make(map[string][]byte)

	for _, key := range keys {
		cacheKey := generateCacheKey(propertyID, key)

		item, ok := items[cacheKey]
		if !ok {
			continue
		}

		record, err := cacheRecordFromGob(item.Value)
		if err != nil {
			log.Debugf(ctx, "Error getting value from memcache for key %+v", key)
			continue
		}

		if _, ok := data[record.Version]; !ok {
			data[record.Version] = make(map[string][]byte)
		}

		value := record.Value
		if record.Compressed {
			err = func() error {
				zlibReader, err := zlib.NewReader(bytes.NewBuffer(value))
				if err != nil {
					return err
				}
				defer zlibReader.Close()
				dec := gob.NewDecoder(zlibReader)
				events := []byte{}
				err = dec.Decode(&events)
				if err != nil {
					return err
				}
				value = events
				return nil
			}()
			if err != nil {
				return 0, nil, err
			}
		}

		data[record.Version][key] = value

		if version < record.Version {
			version = record.Version
		}
	}

	keysRead := []string{}
	for keyRead := range data[version] {
		keysRead = append(keysRead, keyRead)
	}

	log.Debugf(ctx, "Read from cache propertyID: %+v, version: %+v, keys read: %+v", propertyID, version, keysRead)

	return version, data[version], nil
}

func (r *dataStoreImpl) CreateProperty(ctx context.Context, propertyID string, events []VersionedEvent, persistedPropertyList PersistedPropertyList, nextPropertyListIndex int) (int, error) {

	opts := &datastore.TransactionOptions{XG: true}
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		var err1 error
		if persistedPropertyList != nil {
			err1 = persistedPropertyList.CreateProperty(ctx, propertyID, nextPropertyListIndex)
		}
		if err1 == nil {
			_, err1 = r.NewPropertyEvents(ctx, propertyID, 0, events, true)
		}
		return err1
	}, opts)
	if err != nil {
		return -1, err
	}
	return len(events), nil

}

func (r *dataStoreImpl) DeleteProperty(ctx context.Context, propertyID string, persistedPropertyList PersistedPropertyList) error {

	opts := &datastore.TransactionOptions{XG: true}
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		// get the root key of the property
		rootKey, err1 := propertyParentKey(ctx, propertyID)
		if err1 != nil {
			return err1
		}
		// get all events for the property in a transaction
		persistedPropertyEventsKind := "Y_EVENTS_KIND_" + propertyID
		// get all the keys
		keys := []*datastore.Key{}
		query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey).KeysOnly()
		for iterator := query.Run(ctx); ; {
			key, err1 := iterator.Next(nil)
			if err1 == datastore.Done {
				break
			}
			if err1 != nil {
				return err1
			}
			keys = append(keys, key)
		}

		for _, key := range keys {
			err1 = datastore.Delete(ctx, key)
			if err1 != nil {
				return err1
			}
		}

		if persistedPropertyList != nil {
			if err1 = persistedPropertyList.DeleteProperty(ctx, propertyID); err1 != nil {
				return err1
			}
		}

		// also remove from memcache
		cacheError := r.CacheDelete(ctx, propertyID, currentVersionCacheKeyName)
		if cacheError != nil {
			log.Warningf(ctx, "could not delete memcache key for propertyID: %+v", propertyID)
		}

		return err1
	}, opts)
	if err != nil {
		return err
	}
	return nil

}

func eventsToEntity(ctx context.Context, propertyID string, events []VersionedEvent, version int) (*datastore.Key, *PersistedPropertyEvents, int, error) {
	// add the next events to an array
	nextVersion := version
	for _, iface := range events {
		iface.SetEventVersion(nextVersion)
		nextVersion++
	}

	// encode the array into a gob
	stream := &bytes.Buffer{}
	if consolidateCompress {
		err := func() error {
			zlibWriter := zlib.NewWriter(stream)
			defer zlibWriter.Close()
			en := gob.NewEncoder(zlibWriter)
			err := en.Encode(events)
			return err
		}()
		if err != nil {
			return nil, nil, 0, err
		}
	} else {
		en := gob.NewEncoder(stream)
		err := en.Encode(events)
		if err != nil {
			return nil, nil, 0, err
		}
	}

	// create the entity, key and kind
	eventsEntity := &PersistedPropertyEvents{Compressed: consolidateCompress, Type: 1, First: version, Last: nextVersion - 1, Events: stream.Bytes()}
	rootKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return nil, nil, 0, err
	}
	persistedPropertyEventsKind := "Y_EVENTS_KIND_" + propertyID
	eventsKey := datastore.NewKey(ctx, persistedPropertyEventsKind, strconv.Itoa(version)+"_"+strconv.Itoa(nextVersion-1)+"_"+strconv.Itoa(len(stream.Bytes())), 0, rootKey)

	return eventsKey, eventsEntity, nextVersion, nil
}

// NewPropertyEvents returns next transaction key or error
func (r *dataStoreImpl) NewPropertyEvents(ctx context.Context, propertyID string, transactionKey int, events []VersionedEvent, inTransaction bool) (int, error) {

	eventsKey, eventsEntity, nextVersion, err := eventsToEntity(ctx, propertyID, events, transactionKey)
	if err != nil {
		return -1, err
	}

	// now store the entity
	if inTransaction {
		// this part of code only called during property creation

		// don't check on current version, checked in calling function
		if _, err = datastore.Put(ctx, eventsKey, eventsEntity); err != nil {
			return 0, err
		}
	} else {
		// not in a transaction, so create a transaction now
		opts := &datastore.TransactionOptions{XG: true}
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			// get the expected next id
			nextID, keys, err1 := r.getNextEventIDAndKeys(ctx, propertyID, true)
			if err1 != nil {
				return err1
			}
			// test that the caller is up to date with the latest id
			if transactionKey != nextID {
				return fmt.Errorf("expected id: %+v but got %+v", transactionKey, nextID)
			}

			// consolidate the previous records if necessary
			if err1 = r.consolidate(ctx, propertyID, keys); err1 != nil {
				return err1
			}

			// all ok, now store the entity
			if _, err1 = datastore.Put(ctx, eventsKey, eventsEntity); err1 != nil {
				return err1
			}

			cacheError := r.CacheWrite(ctx, propertyID, nextVersion-1, currentVersionCacheKeyName, []byte{})
			if cacheError != nil {
				log.Warningf(ctx, "could not store current version to memcache")
			}
			return err1
		}, opts)
		if err != nil {
			return -1, err
		}
	}

	return nextVersion, nil
}

// splitKey into first, last and size
func splitKey(key *datastore.Key) (int, int, int, error) {
	id := key.StringID()
	s := strings.Split(id, "_")
	size, err := strconv.Atoi(s[2])
	if err != nil {
		return 0, 0, 0, err
	}
	last, err := strconv.Atoi(s[1])
	if err != nil {
		return 0, 0, 0, err
	}

	first, err := strconv.Atoi(s[0])
	if err != nil {
		return 0, 0, 0, err
	}
	return first, last, size, nil
}

// consolidate is run under a transaction always
func (r *dataStoreImpl) consolidate(ctx context.Context, propertyID string, keys []*datastore.Key) error {

	// order the keys
	sort.Slice(keys, func(i, j int) bool {
		first1, _, _, _ := splitKey(keys[i])
		first2, _, _, _ := splitKey(keys[j])
		return first1 < first2
	})

	consolidateKeys := []*datastore.Key{}
	for _, key := range keys {
		_, _, size, err := splitKey(key)
		if err != nil {
			return err
		}

		// don't consolidate records that are already big
		if size >= consolidateMaxSize {
			continue
		}
		consolidateKeys = append(consolidateKeys, key)

		// only consolidate a few records at a time
		if len(consolidateKeys) >= consolidateNumRecords {
			break
		}
	}

	if len(consolidateKeys) < consolidateNumRecords {
		// not enough to consolidate
		return nil
	}

	// get the records
	events, err := r.getEventsInSet(ctx, propertyID, consolidateKeys)
	if err != nil {
		return err
	}

	consolidatedFirst, _, _, _ := splitKey(consolidateKeys[0])
	_, consolidatedLast, _, _ := splitKey(consolidateKeys[len(consolidateKeys)-1])
	eventsKey, eventsEntity, nextVersion, err := eventsToEntity(ctx, propertyID, events, consolidatedFirst)
	if err != nil {
		return err
	}

	if consolidatedLast+1 != nextVersion {
		return errors.New("wrong version after consolidation")
	}

	// now store the entity
	if _, err = datastore.Put(ctx, eventsKey, eventsEntity); err != nil {
		return err
	}

	// now delete the old entities that have now been consolidated
	if err = datastore.DeleteMulti(ctx, consolidateKeys); err != nil {
		return err
	}

	return nil
}

func (r *dataStoreImpl) GetNextEventID(ctx context.Context, propertyID string, inTransaction bool) (int, error) {
	rootKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return -1, err
	}

	persistedPropertyEventsKind := "Y_EVENTS_KIND_" + propertyID

	keys := []*datastore.Key{}
	if inTransaction {
		// get the expected next id
		query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey).KeysOnly()
		for iterator := query.Run(ctx); ; {
			key, err := iterator.Next(nil)
			if err == datastore.Done {
				break
			}
			if err != nil {
				return -1, err
			}
			keys = append(keys, key)
		}
	} else {
		// not in a transaction, so create a transaction now
		opts := &datastore.TransactionOptions{XG: true}
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			// get the expected next id
			var err1 error
			query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey).KeysOnly()
			for iterator := query.Run(ctx); ; {
				key, err1 := iterator.Next(nil)
				if err1 == datastore.Done {
					break
				}
				if err1 != nil {
					return err1
				}
				keys = append(keys, key)
			}
			return err1
		}, opts)
		if err != nil {
			return -1, err
		}
	}

	count := 0
	last := 0
	for _, key := range keys {
		i := key.StringID()
		s := strings.Split(i, "_")
		recordLast, err := strconv.Atoi(s[1])
		if err != nil {
			return 0, err
		}
		if recordLast > last {
			last = recordLast
		}
		count++
	}

	if count == 0 {
		return 0, errors.New("no events found")
	}

	return last + 1, nil
}

func (r *dataStoreImpl) getNextEventIDAndKeys(ctx context.Context, propertyID string, inTransaction bool) (int, []*datastore.Key, error) {
	rootKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return 0, nil, err
	}

	persistedPropertyEventsKind := "Y_EVENTS_KIND_" + propertyID

	keys := []*datastore.Key{}
	if inTransaction {
		// get the expected next id
		query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey).KeysOnly()
		for iterator := query.Run(ctx); ; {
			key, err := iterator.Next(nil)
			if err == datastore.Done {
				break
			}
			if err != nil {
				return 0, nil, err
			}
			keys = append(keys, key)
		}
	} else {
		// not in a transaction, so create a transaction now
		opts := &datastore.TransactionOptions{XG: true}
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			// get the expected next id
			var err1 error
			query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey).KeysOnly()
			for iterator := query.Run(ctx); ; {
				key, err1 := iterator.Next(nil)
				if err1 == datastore.Done {
					break
				}
				if err1 != nil {
					return err1
				}
				keys = append(keys, key)
			}
			return err1
		}, opts)
		if err != nil {
			return 0, nil, err
		}
	}

	count := 0
	last := 0
	for _, key := range keys {
		i := key.StringID()
		s := strings.Split(i, "_")
		recordLast, err := strconv.Atoi(s[1])
		if err != nil {
			return 0, nil, err
		}
		if recordLast > last {
			last = recordLast
		}
		count++
	}

	if count == 0 {
		return 0, nil, errors.New("no events found")
	}

	return last + 1, keys, nil
}

func (r *dataStoreImpl) NumRecords(ctx context.Context, propertyID string) (int, error) {
	rootKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return 0, err
	}

	persistedPropertyEventsKind := "Y_EVENTS_KIND_" + propertyID

	keys := []*datastore.Key{}

	// not in a transaction, so create a transaction now
	opts := &datastore.TransactionOptions{XG: true}
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		// get the expected next id
		var err1 error
		query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey).KeysOnly()
		for iterator := query.Run(ctx); ; {
			key, err1 := iterator.Next(nil)
			if err1 == datastore.Done {
				break
			}
			if err1 != nil {
				return err1
			}
			keys = append(keys, key)
		}
		return err1
	}, opts)
	if err != nil {
		return 0, err
	}

	return len(keys), nil

}

func eventsFromEntities(entities []PersistedPropertyEvents) ([]VersionedEvent, error) {

	// put events into a map to ensure no duplicates
	eventsMap := make(map[int]VersionedEvent)
	for _, eventsEntity := range entities {
		if eventsEntity.Type == 1 {

			if eventsEntity.Compressed {

				err := func() error {
					zlibReader, err := zlib.NewReader(bytes.NewBuffer(eventsEntity.Events))
					if err != nil {
						return err
					}
					defer zlibReader.Close()
					dec := gob.NewDecoder(zlibReader)
					events := []VersionedEvent{}
					err = dec.Decode(&events)
					if err != nil {
						return err
					}

					for _, event := range events {
						eventsMap[event.GetEventVersion()] = event
					}
					return nil
				}()

				if err != nil {
					return nil, err
				}
			} else {
				dec := gob.NewDecoder(bytes.NewBuffer(eventsEntity.Events))
				events := []VersionedEvent{}
				err := dec.Decode(&events)
				if err != nil {
					return nil, err
				}
				for _, event := range events {
					eventsMap[event.GetEventVersion()] = event
				}
			}
		} else {
			dec := gob.NewDecoder(bytes.NewBuffer(eventsEntity.Events))
			entityEvents := []Event{}
			err := dec.Decode(&entityEvents)
			if err != nil {
				return nil, err
			}

			for _, event := range entityEvents {
				versionedEvent, ok := event.Value.(VersionedEvent)
				if ok {
					eventsMap[versionedEvent.GetEventVersion()] = versionedEvent
				}
			}
		}

	}

	// get events from map
	events := []VersionedEvent{}
	for _, event := range eventsMap {
		events = append(events, event)
	}

	// order the events
	sort.Slice(events, func(i, j int) bool {
		return events[i].GetEventVersion() < events[j].GetEventVersion()
	})

	return events, nil
}

func (r *dataStoreImpl) GetEvents(ctx context.Context, propertyID string) ([]VersionedEvent, error) {

	// get events from memcache
	_, readData, err := r.CacheRead(ctx, propertyID, []string{versionedEventsCacheKeyName, currentVersionCacheKeyName})
	if err != nil {
		log.Warningf(ctx, "Error reading from cache: %+v", err)
	} else {
		if _, ok := readData[currentVersionCacheKeyName]; ok {
			if readValue, ok := readData[versionedEventsCacheKeyName]; ok {
				events, err := eventsFromGob(readValue)
				if err != nil {
					log.Warningf(ctx, "Error converting cached gob to events: %+v", err)
				}
				return events, nil
			}
		}
	}

	// get events from datastore
	rootKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return nil, err
	}
	// get all events for the property in a transaction
	persistedPropertyEventsKind := "Y_EVENTS_KIND_" + propertyID

	//entities := []PersistedPropertyEvents{}
	events := []VersionedEvent{}
	opts := &datastore.TransactionOptions{XG: true}
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		entities := []PersistedPropertyEvents{}
		query := datastore.NewQuery(persistedPropertyEventsKind).Ancestor(rootKey)
		var err1 error
		for iterator := query.Run(ctx); ; {
			var eventsEntity PersistedPropertyEvents
			_, err1 = iterator.Next(&eventsEntity)
			if err1 == datastore.Done {
				err1 = nil
				if len(entities) == 0 {
					log.Warningf(ctx, "non-existant property record requested: %+v", propertyID)
					return fmt.Errorf("no records found for id %+v", propertyID)
				}

				events, err1 = eventsFromEntities(entities)
				if err1 != nil {
					return err1
				}
				if len(events) != events[len(events)-1].GetEventVersion()+1 {
					log.Errorf(ctx, "number of events (%+v) does not match last version + 1 (%+v",
						len(events), events[len(events)-1].GetEventVersion()+1)
				}

				// ok great we got events, now write back to cache in this transaction
				gobData, cacheError := gobFromEvents(events)
				if cacheError != nil {
					log.Errorf(ctx, "Error getting gob from events: %+v", cacheError)
					break
				}
				cacheError = r.CacheWrite(ctx, propertyID, events[len(events)-1].GetEventVersion(), versionedEventsCacheKeyName, gobData)
				if cacheError != nil {
					log.Errorf(ctx, "Error writing events to cache: %+v", cacheError)
				}
				cacheError = r.CacheWrite(ctx, propertyID, events[len(events)-1].GetEventVersion(), currentVersionCacheKeyName, []byte{})
				if cacheError != nil {
					log.Errorf(ctx, "Error writing current version to cache: %+v", cacheError)
				}
				break
			}
			if err1 != nil {
				return err1
			}
			entities = append(entities, eventsEntity)
		}
		return err1
	}, opts)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *dataStoreImpl) getEventsInSet(ctx context.Context, propertyID string, keys []*datastore.Key) ([]VersionedEvent, error) {

	// TODO: get records from memcache
	entities := make([]PersistedPropertyEvents, len(keys))
	err := datastore.GetMulti(ctx, keys, entities)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		log.Warningf(ctx, "non-existant property record requested: %+v", propertyID)
		return nil, fmt.Errorf("no records found for id %+v", propertyID)
	}

	events, err := eventsFromEntities(entities)
	if err != nil {
		return nil, err
	}

	return events, nil

}
