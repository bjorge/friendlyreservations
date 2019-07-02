package frapi

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type testStruct1 struct {
	a       int
	b       string
	Version int
}

func (r *testStruct1) GetEventVersion() int {
	return r.Version
}

func TestRollupUtilNew(t *testing.T) {
	property, _, _, _, _ := initAndCreateTestProperty(context.Background(), t)

	x1 := &testStruct1{a: 1, b: "one", Version: 1}

	//resolverMap := make(map[string][]models.VersionedRollup)

	testRollupType := notificationRollupType

	property.addRollup("id1", x1, testRollupType)

	s1, ok := property.property.Rollups[testRollupType]["id1"][0].(*testStruct1)
	if !ok {
		t.Fatal(errors.New("wrong struct came back"))
	}

	if *x1 != *s1 {
		t.Fatal(errors.New("did not store object to map"))
	}

	if x1.a != s1.a {
		t.Fatal(errors.New("content of object wrong"))
	}

	x2 := &testStruct1{a: 2, b: "two", Version: 2}
	property.addRollup("id1", x2, testRollupType)
	x3 := &testStruct1{a: 3, b: "three", Version: 3}
	property.addRollup("id1", x3, testRollupType)

	s3, _ := property.property.Rollups[testRollupType]["id1"][2].(*testStruct1)
	if s3.a != 3 {
		t.Fatal(errors.New("returned wrong object in map"))
	}

	id1 := "id1"
	resolvers := property.getRollups(&rollupArgs{id: &id1}, testRollupType)
	s3 = resolvers[0].(*testStruct1)
	if s3.a != 3 {
		t.Fatal(errors.New("returned wrong version of object when capping the maximum version"))
	}

	maxVersion := int32(2)
	resolvers = property.getRollups(&rollupArgs{id: &id1, maxVersion: &maxVersion}, testRollupType)
	s2 := resolvers[0].(*testStruct1)
	//s2 := GetObjectFromRollupMap("id1", 2, resolverMap).(*testStruct1)
	//t.Logf("returned: %+v", s2)
	if s2.a != 2 {
		t.Fatal(errors.New("returned wrong version of object when capping the maximum version"))
	}

	ifaces := property.getRollups(&rollupArgs{}, testRollupType)
	//ifaces := GetAllLastObjectsFromRollupMap(nil, resolverMap)
	if len(ifaces) != 1 {
		t.Fatal(fmt.Errorf("wrong number of interfaces returned, expected 1 but got %v", len(ifaces)))
	}

	y1 := &testStruct1{a: 1, b: "one"}
	property.addRollup("id2", y1, testRollupType)

	ifaces = property.getRollups(&rollupArgs{}, testRollupType)
	//ifaces = GetAllLastObjectsFromRollupMap(nil, resolverMap)
	if len(ifaces) != 2 {
		t.Fatal(errors.New("wrong number of interfaces returned, expected 2"))
	}
}
