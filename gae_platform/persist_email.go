package gaeplatform

import (
	"context"
	"errors"
	"strings"

	"github.com/bjorge/friendlyreservations/platform"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/appengine/datastore"
)

// PersistedEmail is the structure used to store an email address and other PII information
type PersistedEmail struct {
	EmailID    string `datastore:"EmailId"`    // legacy name
	Email      string `datastore:"Email"`      // legacy name
	PropertyID string `datastore:"PropertyId"` // legacy name
}

// BUG(bjorge): change array items []PersistedEmail to array of pointers

type dataStoreEmailImpl struct{}

// NewPersistedEmailStore is the factory method to create an email store
func NewPersistedEmailStore() platform.PersistedEmailStore {
	return &dataStoreEmailImpl{}
}

var persistedEmailsKind = "PERSISTED_EMAILS_KIND"

var emailKeyDelimiter = ":"

var emailKeyPrefix = "EMAIL_KEY"

func emailRecordKey(ctx context.Context, propertyID string, email string) (*datastore.Key, *datastore.Key, error) {
	// base the key on the email address and propertyId in order to force uniqueness
	parentKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return nil, nil, err
	}

	trimmedEmail := strings.ToLower(strings.TrimSpace(email))

	stringID := emailKeyPrefix + emailKeyDelimiter + propertyID + emailKeyDelimiter + trimmedEmail
	if len(stringID) > 500 {
		return nil, nil, errors.New("email is too long")
	}

	key := datastore.NewKey(ctx, persistedEmailsKind, stringID, 0, parentKey)
	return key, parentKey, nil
}

func (r *dataStoreEmailImpl) GetPropertiesByEmail(ctx context.Context, email string) ([]string, error) {
	trimmedEmail := strings.ToLower(strings.TrimSpace(email))

	query := datastore.NewQuery(persistedEmailsKind).Filter("Email =", trimmedEmail)
	records := []string{}
	for iterator := query.Run(ctx); ; {
		entity := &PersistedEmail{}
		_, err := iterator.Next(entity)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, entity.PropertyID)
	}

	return records, nil
}

// GetEmailMap returns a map of email id to email object
func (r *dataStoreEmailImpl) GetEmailMap(ctx context.Context, propertyID string) (map[string]string, error) {
	emailMap := make(map[string]string)

	parentKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return nil, err
	}

	x := &PersistedEmail{}
	opts := &datastore.TransactionOptions{XG: true}
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		query := datastore.NewQuery(persistedEmailsKind).Ancestor(parentKey)
		var err1 error
		for iterator := query.Run(ctx); ; {
			_, err1 = iterator.Next(x)
			if err1 == nil {
				emailMap[x.EmailID] = x.Email
			} else if err1 == datastore.Done {
				err1 = nil
				break
			} else {
				break
			}
		}
		return err1
	}, opts)
	if err != nil {
		return nil, err
	}
	return emailMap, nil
}

func (r *dataStoreEmailImpl) RestoreEmails(ctx context.Context, propertyID string, persistedEmails map[string]string) error {

	for emailID, email := range persistedEmails {
		//email := record.Email
		key, _, err := emailRecordKey(ctx, propertyID, email)
		if err != nil {
			return err
		}
		//emailId := record.EmailId
		record := &PersistedEmail{Email: email, EmailID: emailID, PropertyID: propertyID}
		opts := &datastore.TransactionOptions{XG: true}
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			_, err = datastore.Put(ctx, key, record)
			return err
		}, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *dataStoreEmailImpl) GetEmail(ctx context.Context, propertyID string, email string) (string, error) {
	trimmedEmail := strings.ToLower(strings.TrimSpace(email))

	key, parentKey, err := emailRecordKey(ctx, propertyID, trimmedEmail)
	if err != nil {
		return "", err
	}

	record := &PersistedEmail{}
	opts := &datastore.TransactionOptions{XG: true}
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		query := datastore.NewQuery(persistedEmailsKind).Ancestor(parentKey).Filter("__key__ =", key)
		iterator := query.Run(ctx)
		_, err1 := iterator.Next(record)
		return err1
	}, opts)
	if err != nil {
		return "", err
	}
	return record.EmailID, nil
}

func (r *dataStoreEmailImpl) EmailExists(ctx context.Context, propertyID string, email string) (*bool, error) {
	_, err := r.GetEmail(ctx, propertyID, email)
	var exists = true
	var doesNotExist = false
	if err == nil {
		return &exists, nil
	} else if err == datastore.Done {
		return &doesNotExist, nil
	} else {
		return nil, err
	}
}

func (r *dataStoreEmailImpl) DeleteEmails(ctx context.Context, propertyID string) error {

	parentKey, err := propertyParentKey(ctx, propertyID)
	if err != nil {
		return err
	}

	keys := []*datastore.Key{}
	query := datastore.NewQuery(persistedEmailsKind).Ancestor(parentKey).KeysOnly()
	for iterator := query.Run(ctx); ; {
		key, err := iterator.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		keys = append(keys, key)
	}

	for _, key := range keys {
		datastore.Delete(ctx, key)
	}

	return nil

}

func (r *dataStoreEmailImpl) CreateEmail(ctx context.Context, propertyID string, email string) (string, error) {

	existingEmail, err := r.GetEmail(ctx, propertyID, email)
	if err == nil {
		return existingEmail, nil
	} else if err == datastore.Done {
		key, _, err := emailRecordKey(ctx, propertyID, email)
		if err != nil {
			return "", err
		}
		emailID := uuid.Must(uuid.NewV4()).String()
		record := &PersistedEmail{Email: email, EmailID: emailID, PropertyID: propertyID}
		opts := &datastore.TransactionOptions{XG: true}
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			_, err = datastore.Put(ctx, key, record)
			return err
		}, opts)
		if err != nil {
			return "", err
		}
		return emailID, nil
	} else {
		return "", err
	}
}
