package persist

import (
	"context"
	"errors"

	"github.com/bjorge/friendlyreservations/platform"
	"google.golang.org/appengine/datastore"
)

type dataStoreListImpl struct{}
type unitTestListImpl struct {
	PropertyList []string
}

// NewPersistedPropertyList is the factory method to create an property list store
func NewPersistedPropertyList(unitTest bool) platform.PersistedPropertyList {
	if unitTest {
		return &unitTestListImpl{PropertyList: []string{}}
	}
	return &dataStoreListImpl{}
}

func (r *unitTestListImpl) CreateProperty(ctx context.Context, propertyID string, idx int) error {
	length := len(r.PropertyList)
	if idx != length {
		return errors.New("CreatePropertyRecords wrong idx")
	}
	r.PropertyList = append(r.PropertyList, propertyID)
	return nil
}

func (r *unitTestListImpl) DeleteProperty(ctx context.Context, propertyID string) error {
	newList := []string{}
	for _, property := range r.PropertyList {
		if property != propertyID {
			newList = append(newList, property)
		}
	}
	r.PropertyList = newList
	return nil
}

func (r *unitTestListImpl) GetNextVersion(ctx context.Context) (int, error) {
	return len(r.PropertyList), nil
}

func (r *unitTestListImpl) GetProperties(ctx context.Context) ([]string, error) {
	return r.PropertyList, nil
}

func (r *dataStoreListImpl) CreateProperty(ctx context.Context, propertyID string, idx int) error {

	// todo: do a multi-query of this query and next
	// create/update the properties root entity
	propertiesRootKey, err := propertiesParentKey(ctx)
	if err != nil {
		return err
	}

	// now query for the last index in the table
	nextIndex, err := r.GetNextVersion(ctx)
	if err != nil {
		return err
	}

	// if the last index does not match, this means that the client either
	// has the wrong index or the client has attempted the request twice and there
	// is a transaction error, or two clients have simultaneously tried to create
	// a property
	// in any error case the client should refresh and try again (if needed)
	if idx != nextIndex {
		return errors.New("new property index not correct")
	}

	propertyLookup := &PersistedProperties{ParentKey: propertyID, TransactionIndex: idx}

	key := datastore.NewKey(ctx, persistedPropertiesKind, propertyID, 0, propertiesRootKey)

	if _, err := datastore.Put(ctx, key, propertyLookup); err != nil {
		return err
	}

	if _, err := propertyParentKey(ctx, propertyID); err != nil {
		return err
	}

	return nil
}

func (r *dataStoreListImpl) DeleteProperty(ctx context.Context, propertyID string) error {

	// create/update the properties root entity
	propertiesRootKey, err := propertiesParentKey(ctx)
	if err != nil {
		return err
	}

	key := datastore.NewKey(ctx, persistedPropertiesKind, propertyID, 0, propertiesRootKey)

	if err = datastore.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

func (r *dataStoreListImpl) GetNextVersion(ctx context.Context) (int, error) {
	// query for the last index in the table
	// todo: get from cache first
	//propertiesRootKey := datastore.NewKey(ctx, persistedPropertiesRootKind, PersistedPropertiesRootKindKey, 0, nil)
	propertiesRootKey, err := propertiesParentKey(ctx)
	if err != nil {
		return 0, err
	}
	lastIdxQuery := datastore.NewQuery(persistedPropertiesKind).Ancestor(propertiesRootKey).Order("-i").Limit(1)

	// todo: consider using GetAll() instead
	var x PersistedProperties
	iterator := lastIdxQuery.Run(ctx)
	_, err = iterator.Next(&x)
	if err == datastore.Done {
		return 0, nil // return the first index for an empty set, 0
	}
	if err != nil {
		return 0, err // return an error
	}
	return x.TransactionIndex + 1, nil // return the next available index
}

// GetProperties returns a list of all the properties
func (r *dataStoreListImpl) GetProperties(ctx context.Context) ([]string, error) {
	propertiesRootKey, err := propertiesParentKey(ctx)
	if err != nil {
		return nil, err
	}
	// propertiesRootKey := datastore.NewKey(ctx, persistedPropertiesRootKind, PersistedPropertiesRootKindKey, 0, nil)
	query := datastore.NewQuery(persistedPropertiesKind).Ancestor(propertiesRootKey).Order("i")

	ids := []string{}
	for iterator := query.Run(ctx); ; {
		var entity PersistedProperties
		_, err := iterator.Next(&entity)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, entity.ParentKey)
	}

	return ids, nil
}
