package frapi

import (
	"context"
	"errors"
	"testing"

	"github.com/bjorge/friendlyreservations/models"
)

var defaultPropertyInput = &models.NewPropertyInput{
	PropertyName:    "Test Property",
	Currency:        models.EUR,
	MemberRate:      4000,
	AllowNonMembers: true,
	NonMemberRate:   8000,
	IsMember:        true,
	NickName:        "Mr. Admin",
	Timezone:        "America/Los_Angeles",
}

var defaultEmail = "a@b.com"

func TestSmokeTest1(t *testing.T) {
	// Just make sure a property can be created to start with
	_, ctx, resolver, _, _ := initAndCreateTestProperty(context.Background(), t)

	properties, err := resolver.Properties(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(properties) != 1 {
		t.Fatal(errors.New("Properties: expected 1 property"))
	}

}

func TestHappyPath(t *testing.T) {
	// skip error checking and just roll through all the methods
	property, ctx, resolver, _, today := initAndCreateTestProperty(context.Background(), t)

	// Query properties
	properties, _ := resolver.Properties(ctx)
	if len(properties) != 1 {
		t.Fatal(errors.New("Properties: expected 1 property"))
	}

	// Query property
	property, _ = resolver.Property(ctx, &struct{ ID string }{ID: property.PropertyID()})

	me, _ := property.Me()

	// Create reservation
	newReservationInput := &models.NewReservationInput{
		// Fields received from the client
		ForVersion:        property.EventVersion(),
		ReservedForUserId: me.UserID(),
		StartDate:         today.AddDays(1).ToString(),
		EndDate:           today.AddDays(2).ToString(),
		Member:            true,
	}

	property, err := resolver.CreateReservation(ctx,
		&struct {
			PropertyID string
			Input      *models.NewReservationInput
		}{
			PropertyID: property.PropertyID(),
			Input:      newReservationInput,
		})

	if err != nil {
		t.Fatal(err)
	}

	// Query reservations
	reservations, _ := property.Reservations(&reservationsArgs{})

	if len(reservations) != 1 {
		t.Fatal(errors.New("TestCreateReservation: expected 1 reservation"))
	}

	// Create user

}
