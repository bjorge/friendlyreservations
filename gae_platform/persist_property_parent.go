package gaeplatform

import (
	"context"

	"google.golang.org/appengine/datastore"
)

var propertyParentKindPrefix = "X_PROPERTY_PARENT:"

func propertyParentKey(ctx context.Context, propertyID string) (*datastore.Key, error) {

	parentKind := propertyParentKindPrefix + propertyID

	// if value, ok := emailMap.Load(parentKind); ok {
	// 	return value.(*datastore.Key), nil
	// }

	parentKey := datastore.NewKey(ctx, parentKind, propertyID, 0, nil)
	// if _, err := datastore.Put(ctx, parentKey, &emptyStruct{}); err != nil {
	// 	return nil, err
	// }

	// emailMap.Store(parentKind, parentKey)

	return parentKey, nil
}

// func removePropertyParentKey(ctx context.Context, propertyId string) error {

// 	parentKind := propertyParentKindPrefix + propertyId

// 	emailMap.Delete(parentKind)

// 	return nil
// }
