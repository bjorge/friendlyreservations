package persist

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
type unitTestEmailImpl struct {
	propertyToEmailIds map[string]map[string]PersistedEmail
	emailToProperties  map[string][]PersistedEmail
}

// NewPersistedEmailStore is the factory method to create an email store
func NewPersistedEmailStore(unitTest bool) platform.PersistedEmailStore {
	if !unitTest {
		return &dataStoreEmailImpl{}
	}
	return &unitTestEmailImpl{
		propertyToEmailIds: make(map[string]map[string]PersistedEmail),
		emailToProperties:  make(map[string][]PersistedEmail),
	}

}

var persistedEmailsKind = "PERSISTED_EMAILS_KIND"

var emailKeyDelimiter = ":"

var emailKeyPrefix = "EMAIL_KEY"

func (r *unitTestEmailImpl) CreateEmail(ctx context.Context, propertyID string, email string) (string, error) {
	_, ok := r.propertyToEmailIds[propertyID]
	if !ok {
		r.propertyToEmailIds[propertyID] = make(map[string]PersistedEmail)
	}

	emailID := uuid.Must(uuid.NewV4()).String()
	persistedEmail := PersistedEmail{EmailID: emailID, Email: email, PropertyID: propertyID}
	r.propertyToEmailIds[propertyID][emailID] = persistedEmail

	_, ok = r.emailToProperties[email]
	if !ok {
		r.emailToProperties[email] = []PersistedEmail{}
	}
	r.emailToProperties[email] = append(r.emailToProperties[email], persistedEmail)

	return emailID, nil
}
func (r *unitTestEmailImpl) GetPropertiesByEmail(ctx context.Context, email string) ([]string, error) {
	if _, ok := r.emailToProperties[email]; ok {
		properties := []string{}
		for _, record := range r.emailToProperties[email] {
			properties = append(properties, record.PropertyID)
		}
		return properties, nil
	}
	return []string{}, nil
}
func (r *unitTestEmailImpl) GetEmailMap(ctx context.Context, propertyID string) (map[string]string, error) {
	if _, ok := r.propertyToEmailIds[propertyID]; ok {
		emailMap := make(map[string]string)
		for _, record := range r.propertyToEmailIds[propertyID] {
			emailMap[record.EmailID] = record.Email
		}
		return emailMap, nil
	}
	return nil, nil
}

func (r *unitTestEmailImpl) RestoreEmails(ctx context.Context, propertyID string, persistedEmails map[string]string) error {

	for emailID, email := range persistedEmails {
		_, ok := r.propertyToEmailIds[propertyID]
		if !ok {
			r.propertyToEmailIds[propertyID] = make(map[string]PersistedEmail)
		}

		// email := record.Email
		// emailId := record.EmailId
		persistedEmail := PersistedEmail{EmailID: emailID, Email: email, PropertyID: propertyID}
		r.propertyToEmailIds[propertyID][emailID] = persistedEmail

		_, ok = r.emailToProperties[email]
		if !ok {
			r.emailToProperties[email] = []PersistedEmail{}
		}
		r.emailToProperties[email] = append(r.emailToProperties[email], persistedEmail)
	}
	return nil
}

func (r *unitTestEmailImpl) GetEmail(ctx context.Context, propertyID string, email string) (string, error) {
	if _, ok := r.emailToProperties[email]; ok {
		properties := r.emailToProperties[email]
		for _, persistedEmail := range properties {
			if persistedEmail.PropertyID == propertyID {
				return persistedEmail.EmailID, nil
			}
		}
	}
	return "", nil

}
func (r *unitTestEmailImpl) EmailExists(ctx context.Context, propertyID string, email string) (*bool, error) {
	records, ok := r.emailToProperties[email]
	if !ok {
		return &ok, nil
	}
	ok = false
	for _, record := range records {
		if record.PropertyID == propertyID {
			ok = true
			return &ok, nil
		}
	}
	return &ok, nil
}

func (r *unitTestEmailImpl) DeleteEmails(ctx context.Context, propertyID string) error {
	if mapEmailIds, ok := r.propertyToEmailIds[propertyID]; ok {
		for _, record1 := range mapEmailIds {
			newList := []PersistedEmail{}
			email := record1.Email
			for _, record2 := range r.emailToProperties[email] {
				if record2.PropertyID != propertyID {
					newList = append(newList, record2)
				}
			}
			r.emailToProperties[email] = newList
		}
		delete(r.propertyToEmailIds, propertyID)
	}
	return nil
}

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
