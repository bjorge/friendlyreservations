package platform

import "context"

// PersistedPropertyList is the interface for managing property ids in the system
type PersistedPropertyList interface {
	CreateProperty(ctx context.Context, propertyID string, version int) error
	GetNextVersion(ctx context.Context) (int, error)
	GetProperties(ctx context.Context) ([]string, error)
	DeleteProperty(ctx context.Context, propertyID string) error
}

// VersionedEvent must be implemented by all event objects that are persisted
type VersionedEvent interface {
	GetEventVersion() int
	SetEventVersion(Version int)
}

// PersistedVersionedEvents is the interface for managing a property
type PersistedVersionedEvents interface {
	CreateProperty(ctx context.Context, propertyID string, events []VersionedEvent, persistedPropertyList PersistedPropertyList, nextPropertyListIndex int) (int, error)
	NewPropertyEvents(ctx context.Context, propertyID string, transactionKey int, events []VersionedEvent, inTransaction bool) (int, error)
	GetNextEventID(ctx context.Context, propertyID string, inTransaction bool) (int, error)
	GetEvents(ctx context.Context, propertyID string) ([]VersionedEvent, error)
	DeleteProperty(ctx context.Context, propertyID string, persistedPropertyList PersistedPropertyList) error
	NumRecords(ctx context.Context, propertyID string) (int, error)
	CacheWrite(ctx context.Context, propertyID string, version int, key string, value []byte) error
	CacheRead(ctx context.Context, propertyID string, keys []string) (int, map[string][]byte, error)
	CacheDelete(ctx context.Context, propertyID string, key string) error
}

// PersistedEmailStore is the interface for accessing the store of email addresses
type PersistedEmailStore interface {
	CreateEmail(ctx context.Context, propertyID string, email string) (string, error)
	GetPropertiesByEmail(ctx context.Context, email string) ([]string, error)
	GetEmailMap(ctx context.Context, propertyID string) (map[string]string, error)
	RestoreEmails(ctx context.Context, propertyID string, persistedEmails map[string]string) error
	GetEmail(ctx context.Context, propertyID string, email string) (string, error)

	// BUG(bjorge): change *bool to be bool in EmailExists()

	EmailExists(ctx context.Context, propertyID string, email string) (*bool, error)
	DeleteEmails(ctx context.Context, propertyID string) error
}

// An EmailAttachment represents an email attachment.
type EmailAttachment struct {
	// Name must be set to a valid file name.
	Name      string
	Data      []byte
	ContentID string
}

// An EmailMessage represents an email message.
// Addresses may be of any form permitted by RFC 822.
type EmailMessage struct {
	Sender  string
	ReplyTo string // may be empty

	To, Cc, Bcc []string

	Subject string

	Body     string
	HTMLBody string

	Attachments []EmailAttachment
}

// SendMail is the interface for sending out emails
type SendMail interface {
	// Send sends an email message.
	Send(ctx context.Context, msg *EmailMessage) error
}

// Logger is the interface for logging
type Logger interface {
	LogDebugf(format string, args ...interface{})
	LogErrorf(format string, args ...interface{})
	LogInfof(format string, args ...interface{})
	LogWarningf(format string, args ...interface{})
}
