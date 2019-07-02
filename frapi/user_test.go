package frapi

import (
	"context"
	"errors"
	"testing"

	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

func TestUserResolver(t *testing.T) {
	property, _, _, me, _ := initAndCreateTestProperty(context.Background(), t)

	t.Logf("me email: %+v", me.Email())

	users := property.Users(allUsers())
	if len(users) != 2 {
		// expect system user and admin user
		t.Fatal(errors.New("TestUserResolver: expected a single user"))
	}

	// validate strings and presence of system user
	systemUserCount := 0
	for _, user := range users {
		validateString(t, user.Email())
		validateString(t, user.emailID())
		validateString(t, user.Nickname())
		validateUserState(t, user.State())
		validateString(t, user.UserID())
		if user.IsSystem() {
			systemUserCount++
		}
	}

	if systemUserCount != 1 {
		t.Fatal(errors.New("TestUserResolver: missing system user"))
	}

	// search by email
	users = property.Users(byEmail(defaultEmail))
	adminUser := users[0]
	if adminUser.Email() != defaultEmail {
		t.Fatal(errors.New("TestUserResolver: wrong email"))
	}

	// search by userId
	users = property.Users(singleUser(users[0].UserID()))
	adminUser = users[0]
	if adminUser.Email() != defaultEmail {
		t.Fatal(errors.New("TestUserResolver: wrong email"))
	}

	// validate other parameters
	validateBool(t, adminUser.IsAdmin(), true)
	validateBool(t, adminUser.IsMember(), defaultPropertyInput.IsMember)
	validateEventVersion(t, adminUser.GetEventVersion())

	// Get all versions
	users = property.Users(singleUserMaxVersion(users[0].UserID(), int32(100)))
	if len(users) != 1 {
		t.Fatal(errors.New("TestUserResolver: wrong number of user versions"))
	}
}

func TestInvitation(t *testing.T) {

	t.Logf("test lower case email")
	testInvitationForEmail(t, "test2@test.com")

	t.Logf("test upper case email")
	testInvitationForEmail(t, "TestThree@test.com")

}

func testInvitationForEmail(t *testing.T, email string) {
	property, ctx, resolver, me, _ := initAndCreateTestProperty(context.Background(), t)

	t.Logf("me email: %+v", me.Email())

	nickname := "test2"

	property, err := resolver.CreateUser(ctx, &struct {
		PropertyID string
		Input      *models.NewUserInput
	}{
		PropertyID: property.PropertyID(),
		Input: &models.NewUserInput{
			ForVersion: property.EventVersion(),
			Email:      email,
			Nickname:   nickname,
			IsAdmin:    false,
			IsMember:   true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	users := property.Users(allUsers())
	if len(users) != 3 {
		// expect system user and admin user
		t.Fatalf("TestUserResolver: expected a 3 users but got: %+v", len(users))
	}

	users = property.Users(byEmail(email))
	if len(users) != 1 {
		t.Fatalf("Expected the user to exist")
	}

	user := users[0]
	if user.State() != models.WAITING_ACCEPT {
		t.Fatalf("TestUserResolver: expected a WAITING_ACCEPT but got: %+v", user.State())
	}

	utilities.SetTestUser(email)
	property, err = resolver.AcceptInvitation(ctx, &struct {
		PropertyID string
		Input      *models.AcceptInvitationInput
	}{
		PropertyID: property.PropertyID(),
		Input: &models.AcceptInvitationInput{
			ForVersion: property.EventVersion(),
			Accept:     true,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	users = property.Users(byEmail(email))
	user = users[0]
	if user.State() != models.ACCEPTED {
		t.Fatalf("TestUserResolver: expected a ACCEPTED but got: %+v", user.State())
	}

}

func validateEventVersion(t *testing.T, value int) {
	if value < 0 {
		t.Fatal("unexpected event version")
	}
}
func validateBool(t *testing.T, value bool, expectedValue bool) {
	if value != expectedValue {
		t.Fatal("unexpected bool")
	}
}

func validateString(t *testing.T, text string) {
	if text == "" {
		t.Fatal(errors.New("unexpected empty string"))
	}
}

func validateUserState(t *testing.T, state models.UserState) {
	switch state {
	case models.ACCEPTED:
	case models.DISABLED:
	case models.WAITING_ACCEPT:
		break
	default:
		t.Fatal(errors.New("unexpected user state"))
	}
}

func allUsers() *usersArgs {
	return &usersArgs{}
}

func singleUser(userID string) *usersArgs {
	return &usersArgs{UserID: &userID}
}

func singleUserMaxVersion(userID string, maxVersion int32) *usersArgs {
	return &usersArgs{UserID: &userID, MaxVersion: &maxVersion}
}

func byEmail(email string) *usersArgs {
	return &usersArgs{Email: &email}
}
