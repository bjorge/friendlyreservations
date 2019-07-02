package frapi

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"sync"

	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/persist"
	"github.com/bjorge/friendlyreservations/utilities"
)

type versionedRollup interface {
	GetEventVersion() int
}

type rollupType string

const (
	ledgerRollupType           rollupType = "LEDGER_ROLLUP"
	notificationRollupType     rollupType = "NOTIFICATION_ROLLUP"
	reservationRollupType      rollupType = "RESERVATION_ROLLUP"
	userRollupType             rollupType = "USER_ROLLUP"
	settingsRollupType         rollupType = "SETTINGS_ROLLUP"
	restrictionRollupType      rollupType = "RESTRICTION_ROLLUP"
	contentsRollupType         rollupType = "CONTENTS_ROLLUP"
	membershipStatusRollupType rollupType = "MEMBERSHIP_STATUS_ROLLUP"
)

var rollupTypes = [...]rollupType{
	ledgerRollupType,
	notificationRollupType,
	reservationRollupType,
	userRollupType,
	settingsRollupType,
	restrictionRollupType,
	contentsRollupType,
	membershipStatusRollupType}

// Property is the basic structure holding information for rollups
// The public fields can be cached (for current latest event version)
type Property struct {
	// sent back to client
	PropertyID     string
	CreateDateTime string

	EmailMap map[string]persist.PersistedEmail

	// reference the events pulled from the database
	Events []persist.VersionedEvent

	// map of rollups for different types of resolvers such as
	// reservations, notifications, etc.
	// each type of resolver is mapped from its id to versions of each id over time
	Rollups map[rollupType]map[string][]versionedRollup
}

// PropertyResolver is the resolver for a single property
type PropertyResolver struct {
	// persisted in memcache
	property *Property
	// created each request
	email         string
	rollupsMutex  sync.RWMutex
	rollupMutexes map[rollupType]*sync.Mutex
	ctx           context.Context
}

// Properties is called to retrieve all the properties for which the caller is a member or admin
func (r *Resolver) Properties(ctx context.Context) ([]*PropertyResolver, error) {
	var l []*PropertyResolver

	// check that a user is logged in
	u := utilities.GetUser(ctx)
	if u == nil {
		return nil, errors.New("user not logged in")
	}

	// get all the properties
	emailRecords, err := persistedEmailStore.GetPropertiesByEmail(ctx, u.Email)
	//ids, err := utilities.PersistedPropertyList.GetProperties(ctx)
	if err != nil {
		return nil, err
	}

	for _, emailRecord := range emailRecords {
		propertyResolver, err := currentBaseProperty(ctx, u.Email, emailRecord.PropertyID)
		if err != nil {
			return nil, err
		}

		users := propertyResolver.Users(&usersArgs{Email: &u.Email})

		if len(users) == 1 && users[0].State() != models.DISABLED && !users[0].IsSystem() {
			l = append(l, propertyResolver)
		}
	}
	return l, nil
}

// Property returns the information for a single property
func (r *Resolver) Property(ctx context.Context, args *struct {
	ID string
}) (*PropertyResolver, error) {

	propertyResolver, _, err := currentProperty(ctx, args.ID)

	if err != nil {
		return nil, err
	}

	return propertyResolver, nil
}

// PropertyID is the unique id of this property
func (r *PropertyResolver) PropertyID() string {
	return r.property.PropertyID
}

// EventVersion is the version of the last property mutation event
func (r *PropertyResolver) EventVersion() int32 {
	lastEvent := r.property.Events[len(r.property.Events)-1]

	return int32(lastEvent.GetEventVersion())
}

// CreateDateTime is the time stamp for when the property was created
func (r *PropertyResolver) CreateDateTime() string {
	return r.property.CreateDateTime
}

// Me is the information for the logged in user
func (r *PropertyResolver) Me() (*UserResolver, error) {
	users := r.Users(&usersArgs{Email: &r.email})

	if len(users) != 1 {
		return nil, errors.New("User not member of property")
	}

	return users[0], nil
}

type rollupArgs struct {
	maxVersion  *int32
	id          *string
	allVersions *bool
}

func getRollupVersion(maxVersion *int32, versions []versionedRollup) versionedRollup {
	var resolver versionedRollup
	if maxVersion == nil {
		// get the latest version
		length := len(versions)
		if length > 0 {
			resolver = versions[length-1]
		}
	} else {
		// get up to a specific version
		for _, item := range versions {
			if item.GetEventVersion() <= int(*maxVersion) {
				resolver = item
			} else {
				break
			}
		}
	}
	return resolver
}

func (r *PropertyResolver) rollupsExists(resolverType rollupType) bool {
	r.rollupsMutex.RLock()
	defer r.rollupsMutex.RUnlock()
	if r.property.Rollups == nil {
		return false
	}

	if _, ok := r.property.Rollups[resolverType]; !ok {
		return false
	}

	return true
}

// the structure is: resolver type -> resolver id -> ordered list of versions of the id
func (r *PropertyResolver) getRollups(args *rollupArgs, resolverType rollupType) []versionedRollup {
	r.rollupsMutex.RLock()
	defer r.rollupsMutex.RUnlock()

	resolvers := []versionedRollup{}

	if r.property.Rollups == nil {
		return resolvers
	}

	resolverMap := r.property.Rollups[resolverType]

	if resolverMap == nil {
		return resolvers
	}

	if args.id == nil {
		// get all the ids
		for _, versions := range resolverMap {
			if args.allVersions != nil && *args.allVersions {
				for _, version := range versions {
					resolvers = append(resolvers, version)
				}
			} else {
				resolver := getRollupVersion(args.maxVersion, versions)
				if resolver != nil {
					resolvers = append(resolvers, resolver)
				}
			}
		}
	} else {
		// get a single id
		id := *args.id
		if versions, ok := resolverMap[id]; ok {
			if args.allVersions != nil && *args.allVersions {
				if args.maxVersion == nil {
					resolvers = versions
				} else {
					// get up to a specific version
					for _, version := range versions {
						if version.GetEventVersion() <= int(*args.maxVersion) {
							resolvers = append(resolvers, version)
						}
					}
				}
			} else {
				resolver := getRollupVersion(args.maxVersion, versions)
				if resolver != nil {
					resolvers = append(resolvers, resolver)
				}
			}
		}
	}

	return resolvers
}

func (r *PropertyResolver) addRollup(id string, resolver versionedRollup,
	resolverType rollupType) error {

	r.rollupsMutex.Lock()
	defer r.rollupsMutex.Unlock()

	if r.property.Rollups == nil {
		r.property.Rollups = make(map[rollupType]map[string][]versionedRollup)
	}

	if _, ok := r.property.Rollups[resolverType]; !ok {
		r.property.Rollups[resolverType] = make(map[string][]versionedRollup)
	}

	resolverMap := r.property.Rollups[resolverType]

	if _, ok := resolverMap[id]; !ok {
		resolverMap[id] = []versionedRollup{}
	}

	// only accept increasing order
	// this policy will ensure unique entries in a simple deterministic way
	if length := len(resolverMap[id]); length > 0 {
		lastVersion := resolverMap[id][length-1].GetEventVersion()
		if lastVersion >= resolver.GetEventVersion() {
			return errors.New("can only add a version that is greater than last version")
		}
	}

	resolverMap[id] = append(resolverMap[id], resolver)
	return nil
}

func rollupsFromGob(gobData []byte) (map[string][]versionedRollup, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(gobData))
	rollups := make(map[string][]versionedRollup)
	err := dec.Decode(&rollups)
	if err != nil {
		return nil, err
	}
	return rollups, nil
}

func gobFromRollups(rollups map[string][]versionedRollup) ([]byte, error) {
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(rollups)
	if err != nil {
		return nil, err
	}
	return stream.Bytes(), err
}

func (r *PropertyResolver) cacheRollup(resolverType rollupType) error {
	// called at the end of a rollup
	r.rollupsMutex.Lock()
	defer r.rollupsMutex.Unlock()

	if r.property.Rollups == nil {
		// called before rollups has been initialized, just return
		return nil
	}

	if _, ok := r.property.Rollups[resolverType]; !ok {
		// if nothing was there to rollup, then this key does not exist yet, so just return
		return nil
	}

	resolverMap := r.property.Rollups[resolverType]

	// cache the rollup
	gobData, cacheError := gobFromRollups(resolverMap)
	if cacheError != nil {
		utilities.LogWarningf(r.ctx, "cache gob from rollups error: %+v", cacheError)
	} else {
		cacheError = persistedVersionedEvents.CacheWrite(r.ctx, r.PropertyID(), int(r.EventVersion()), string(resolverType), gobData)
		if cacheError != nil {
			utilities.LogWarningf(r.ctx, "cache rollups write error: %+v", cacheError)
		}
	}

	return nil
}
