package frapi

import (
	"context"
	"testing"
)

func TestRestrictions(t *testing.T) {
	property, ctx, resolver, _, today := initAndCreateTestProperty(context.Background(), t)

	t.Log("TestBlackoutRestriction")

	property, _, _ = createBlackoutRestriction(ctx, t, resolver, property, today)

	restrictions, err := property.Restrictions(&restrictionsArgs{})
	if err != nil {
		t.Fatal(err)
	}

	if len(restrictions) != 1 {
		t.Fatalf("expected a blackout restriction")
	}

	if _, ok := restrictions[0].Restriction().ToBlackoutRestriction(); !ok {
		t.Fatalf("did not get a blackout restriction")
	}

	property, _, _ = createMembershipRestriction(ctx, t, resolver, property, today.AddDays(370))

	restrictions, _ = property.Restrictions(&restrictionsArgs{})

	if len(restrictions) != 2 {
		t.Fatalf("expected 2 restrictions")
	}

	if _, ok := restrictions[0].Restriction().ToBlackoutRestriction(); !ok {
		t.Fatalf("did not get a blackout restriction")
	}

	if _, ok := restrictions[1].Restriction().ToMembershipRestriction(); !ok {
		t.Fatalf("did not get a membership restriction")
	}

	// if _, ok := restrictions[0].Restriction().ToBlackoutRestriction(); !ok {
	// 	t.Fatalf("did not get a blackout restriction")
	// }
}

// func TestRestrictionResolver(t *testing.T) {

// 	firstEnabledDay := "2018-02-02"
// 	blockoutStart := "2018-02-04"
// 	blockoutEnd := "2018-02-06"
// 	lastEnabledDay := "2018-02-12"

// 	// setup the property with some restrictions
// 	property := &models.Property{}
// 	testUserId := "abc"
// 	testRestrictionId1 := "xyz"
// 	testRestrictionId2 := "zyx"
// 	testRestrictionId3 := "yzx"

// 	newUserInput := &models.NewUserInput{
// 		Nickname:     "testUser",
// 		UserId:       testUserId,
// 		EventVersion: 0,
// 	}

// 	restriction3 := &models.NewRestrictionInput{
// 		Blackout:      &models.BlackoutRestriction{StartDate: blockoutStart, EndDate: blockoutEnd},
// 		RestrictionId: testRestrictionId3,
// 		AuthorUserId:  testUserId,
// 		EventVersion:  3,
// 	}

// 	property.Events = []platform.VersionedEvent{newUserInput, restriction1, restriction2, restriction3}

// 	propertyResolver := &PResolver{property}
// 	restrictions, err := propertyResolver.Restrictions(nil, &struct {
// 		RestrictionId *string
// 		MaxVersion    *int32
// 	}{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(restrictions) != 3 {
// 		t.Logf("restrictions len is: %+v", len(restrictions))
// 		t.Fatal("Expected three restrictions")
// 	}

// 	// make sure results are sorted in EventVersion order
// 	if _, ok := restrictions[0].Restriction().ToFirstDayRestriction(); !ok {
// 		t.Fatal("ToFirstDayRestriction failed")
// 	}
// 	if _, ok := restrictions[1].Restriction().ToLastDayRestriction(); !ok {
// 		t.Fatal("ToLastDayRestriction failed")
// 	}
// 	if _, ok := restrictions[2].Restriction().ToBlackoutRestriction(); !ok {
// 		t.Fatal("ToBlackoutRestriction failed")
// 	}

// 	// ok now let's get just a single restriction
// 	restrictions, err = propertyResolver.Restrictions(nil, &struct {
// 		RestrictionId *string
// 		MaxVersion    *int32
// 	}{RestrictionId: &testRestrictionId1})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(restrictions) != 1 {
// 		t.Fatal("Expected a single restriction")
// 	}

// }
