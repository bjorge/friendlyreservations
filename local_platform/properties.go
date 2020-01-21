package localplatform

import (
	"context"
	"errors"

	"github.com/bjorge/friendlyreservations/platform"
)

type unitTestListImpl struct {
	PropertyList []string
}

// NewPersistedPropertyList is the factory method to create an property list store
func NewPersistedPropertyList() platform.PersistedPropertyList {
	return &unitTestListImpl{PropertyList: []string{}}
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
