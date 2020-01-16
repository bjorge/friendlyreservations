package frapi

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/platform"
)

const emailMapCacheKeyName = "EmailMap"

// DuplicateDetectionEvent must be implemented by all client requests
type DuplicateDetectionEvent interface {
	GetForVersion() int32
}

func emailMapFromGob(gobData []byte) (map[string]string, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(gobData))
	emailMap := make(map[string]string)
	err := dec.Decode(&emailMap)
	if err != nil {
		return nil, err
	}
	return emailMap, nil
}

func gobFromEmailMap(emailMap map[string]string) ([]byte, error) {
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(emailMap)
	if err != nil {
		return nil, err
	}
	return stream.Bytes(), err
}

func currentBaseProperty(ctx context.Context, email string, propertyID string) (*PropertyResolver, error) {

	property := &Property{}

	// get all events for the property
	var err error
	property.Events, err = PersistedVersionedEvents.GetEvents(ctx, propertyID)
	if err != nil {
		return nil, err
	}

	if len(property.Events) == 0 {
		return nil, errors.New("Property not found")
	}

	property.Rollups = make(map[rollupType]map[string][]versionedRollup)

	var created *string
	for _, event := range property.Events {
		if newPropertyInput, ok := event.(*models.NewPropertyInput); ok {
			created = &newPropertyInput.CreateDateTime
		}
	}
	if created == nil {
		return nil, fmt.Errorf("CreateDateTime not found for property %+v", propertyID)
	}
	property.CreateDateTime = *created

	// start setting up the property resolver struct
	property.PropertyID = propertyID
	propertyResolver := &PropertyResolver{
		property:      property,
		email:         email,
		rollupMutexes: make(map[rollupType]*sync.Mutex),
		ctx:           ctx,
	}

	// get the email map and rollups from the cache
	cacheReadKeys := []string{emailMapCacheKeyName}
	for _, rollup := range rollupTypes {
		cacheReadKeys = append(cacheReadKeys, string(rollup))
	}
	cachedVersion, readData, cacheError := PersistedVersionedEvents.CacheRead(ctx, propertyID, cacheReadKeys)
	if cacheError != nil {
		Logger.LogWarningf("cache error: %+v", cacheError)
	}
	cacheUpToDate := false
	if cacheError == nil && int32(cachedVersion) == propertyResolver.EventVersion() {
		cacheUpToDate = true
	} else {
		// dummy empty cache data on error/out of date for easier processing below
		readData = make(map[string][]byte)
	}

	// get email map from cache
	if readValue, ok := readData[emailMapCacheKeyName]; ok && cacheUpToDate {
		property.EmailMap, cacheError = emailMapFromGob(readValue)
		if cacheError != nil {
			Logger.LogWarningf("cache email map from gob error: %+v", cacheError)
		}
	}

	if property.EmailMap == nil {
		// get email map from db since not cached
		property.EmailMap, err = PersistedEmailStore.GetEmailMap(ctx, propertyID)
		if err != nil {
			return nil, err
		}
		// cache the email map
		gobData, cacheError := gobFromEmailMap(property.EmailMap)
		if cacheError != nil {
			Logger.LogWarningf("cache gob from email map error: %+v", cacheError)
		} else {
			PersistedVersionedEvents.CacheWrite(ctx, propertyID, int(propertyResolver.EventVersion()), emailMapCacheKeyName, gobData)
		}
	}

	// initialize the rollups
	for _, rollup := range rollupTypes {
		// get rollup from the cache if it is there
		if readValue, ok := readData[string(rollup)]; ok && cacheUpToDate {
			property.Rollups[rollup], cacheError = rollupsFromGob(readValue)
			if cacheError != nil {
				Logger.LogWarningf("cache rollups from gob error: %+v", cacheError)
			}
		}
		// initialize the rollup mutex
		propertyResolver.rollupMutexes[rollup] = &sync.Mutex{}
	}

	return propertyResolver, nil
}

func currentProperty(ctx context.Context, propertyID string) (*PropertyResolver, *UserResolver, error) {

	u := GetUser(ctx)
	if u == nil {
		return nil, nil, errors.New("user not logged in")
	}

	property, err := currentBaseProperty(ctx, u.Email, propertyID)
	if err != nil {
		return nil, nil, err
	}

	// get the user and update the property with user records
	me, err := property.Me()
	if err != nil {
		return nil, nil, err
	}

	return property, me, nil
}

// VersionedInput is an input based on the client view of the current version
type VersionedInput interface {
	GetForVersion() int32
}

func isDuplicate(ctx context.Context, request DuplicateDetectionEvent, property *PropertyResolver) (bool, error) {
	if request.GetForVersion() == property.EventVersion() {
		// the normal case, the client wants to mutate the current version, so no duplicate check required!
		return false, nil
	}

	// the exception case, the client wants to mutate an older version, this may be because
	// the client is not updated (i.e. some other client has made a change), or this may be a duplicate
	// request from the client. So check for a duplicate, otherwise the request can be processed normally.
	if request.GetForVersion() <= 0 {
		return false, fmt.Errorf("for version cannot be <= 0, it is %v", request.GetForVersion())
	}

	if request.GetForVersion() > property.EventVersion() {
		return false, fmt.Errorf("for version cannot be > current version, it is %v", request.GetForVersion())
	}

	// iterate through all the events for a duplicate (from the last one to first)
	for i := len(property.property.Events) - 1; i > 0; i-- {
		event := property.property.Events[i]

		// a duplicate will have the same type (ex. update settings event)
		if reflect.TypeOf(event) == reflect.TypeOf(request) {

			// a client event that supports duplicate suppression will implment GetForVersion
			// (i.e. internal events like notification event may not)
			// this should always be true because of the type check above...
			if duplicateDetectionEvent, ok := event.(DuplicateDetectionEvent); ok {

				// if the forVersion is the same, then this request is a duplicate
				if duplicateDetectionEvent.GetForVersion() == request.GetForVersion() {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func commitChanges(ctx context.Context, propertyID string, eventVersion int32,
	events ...platform.VersionedEvent) (*PropertyResolver, error) {

	eventList := []platform.VersionedEvent{}
	for _, event := range events {
		eventList = append(eventList, event)
	}

	key := int(eventVersion) + 1

	// key, err := utilities.PersistedVersionedEvents.GetNextEventId(ctx, propertyId, false)
	// if err != nil {
	// 	return nil, err
	// }
	_, err := PersistedVersionedEvents.NewPropertyEvents(ctx, propertyID, key, eventList, false)
	if err != nil {
		return nil, err
	}

	// get all events for the property
	propertyResolver, _, err := currentProperty(ctx, propertyID)

	return propertyResolver, err
}

func trim(arg string) (*string, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return nil, fmt.Errorf("input string empty")
	}
	return &trimmed, nil
}
