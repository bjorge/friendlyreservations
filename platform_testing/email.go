package platformtesting

import (
	"context"
	"errors"
	"testing"

	"github.com/bjorge/friendlyreservations/platform"
)

// TestCreateEmail is called by the platform implementation testing code
func TestCreateEmail(ctx context.Context, t *testing.T, persistedEmailStore platform.PersistedEmailStore) {

	var err error

	propertyID := "id123"
	email := "test@testing.com"

	emailID1, err := persistedEmailStore.CreateEmail(ctx, propertyID, email)
	if err != nil {
		t.Log("CreateEmail failed")
		t.Fatal(err)
	}

	t.Logf("CreateEmail emailID is: %v", emailID1)

	emailID2, err := persistedEmailStore.GetEmail(ctx, propertyID, email)
	if err != nil {
		t.Log("GetEmail failed")
		t.Fatal(err)
	}
	if emailID1 != emailID2 {
		t.Fatal(errors.New("email ids do not match"))
	}

	ids, err := persistedEmailStore.GetPropertiesByEmail(ctx, email)

	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 1 {
		t.Fatal(errors.New("search for properties by email failed"))
	}

	if ids[0] != propertyID {
		t.Fatal(errors.New("wrong propertyId from GetPropertiesByEmail"))
	}

	emailMap, err := persistedEmailStore.GetEmailMap(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	if emailMap[emailID1] != email {
		t.Logf("expected: %v got: %v", email, emailMap[emailID1])
		t.Fatal(errors.New("email map does not work"))
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, email); !*exists {
		t.Fatal(errors.New("email should exist"))
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, "howdy!"); *exists {
		t.Fatal(errors.New("email should not exist"))
	}

}

// TestDeleteEmail is called by the platform implementation testing code
func TestDeleteEmail(ctx context.Context, t *testing.T, persistedEmailStore platform.PersistedEmailStore) {

	propertyID := "id123"
	email := "test@testing.com"

	if _, err := persistedEmailStore.CreateEmail(ctx, propertyID, email); err != nil {
		t.Fatal(err)
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, email); !*exists {
		t.Fatal(errors.New("email should exist"))
	}

	if err := persistedEmailStore.DeleteEmails(ctx, propertyID); err != nil {
		t.Fatal(err)
	}

	if exists, _ := persistedEmailStore.EmailExists(ctx, propertyID, email); *exists {
		t.Fatal(errors.New("email should not exist"))
	}

}

// TestRestoreEmail is called by the platform implementation testing code
func TestRestoreEmail(ctx context.Context, t *testing.T, persistedEmailStore platform.PersistedEmailStore) {

	propertyID := "id1234"
	email := "test1@testing.com"

	emailID1, err := persistedEmailStore.CreateEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	emailID2, err := persistedEmailStore.GetEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	if emailID1 != emailID2 {
		t.Fatal(errors.New("email does not match"))
	}

	err = persistedEmailStore.DeleteEmails(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}

	emailBackup := make(map[string]string)
	emailBackup[emailID1] = email
	err = persistedEmailStore.RestoreEmails(ctx, propertyID, emailBackup)
	if err != nil {
		t.Fatal(err)
	}

	emailID3, err := persistedEmailStore.GetEmail(ctx, propertyID, email)
	if err != nil {
		t.Fatal(err)
	}

	if emailID1 != emailID3 {
		t.Fatal(errors.New("email id does not match"))
	}

	emailMap, err := persistedEmailStore.GetEmailMap(ctx, propertyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(emailMap) == 0 {
		t.Fatal(errors.New("after restore email map is empty"))
	}

	if emailMap[emailID1] != email {
		t.Fatal(errors.New("email does not match after restore"))
	}
}
