package localplatform

import (
	"context"

	"github.com/bjorge/friendlyreservations/platform"
	uuid "github.com/satori/go.uuid"
)

// PersistedEmail is the structure used to store an email address and other PII information
type PersistedEmail struct {
	EmailID    string `datastore:"EmailId"`    // legacy name
	Email      string `datastore:"Email"`      // legacy name
	PropertyID string `datastore:"PropertyId"` // legacy name
}

type unitTestEmailImpl struct {
	propertyToEmailIds map[string]map[string]PersistedEmail
	emailToProperties  map[string][]PersistedEmail
}

// NewPersistedEmailStore is the factory method to create an email store
func NewPersistedEmailStore() platform.PersistedEmailStore {
	return &unitTestEmailImpl{
		propertyToEmailIds: make(map[string]map[string]PersistedEmail),
		emailToProperties:  make(map[string][]PersistedEmail),
	}
}

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
