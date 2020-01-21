package gaeplatform

import (
	"context"

	"google.golang.org/appengine/datastore"
)

var persistedPropertiesRootKind = "PERSISTED_PROPERTIES_PARENT_KIND"

var persistedPropertiesRootKindKey = "3734c846-7ea3-40b4-8241-9fd8136876fb"

func propertiesParentKey(ctx context.Context) (*datastore.Key, error) {
	// the parent key does not need to be persisted, see:
	// https://cloud.google.com/appengine/docs/standard/go/datastore/entities#Go_Ancestor_paths

	propertiesRootKey := datastore.NewKey(ctx, persistedPropertiesRootKind, persistedPropertiesRootKindKey, 0, nil)
	// if _, err := datastore.Put(ctx, propertiesRootKey, &emptyStruct{}); err != nil {
	// 	return nil, err
	// }

	return propertiesRootKey, nil
}
