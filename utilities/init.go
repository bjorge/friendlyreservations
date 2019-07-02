package utilities

import (
	"os"
	"time"
)

// SystemEmail is the default from email address
var SystemEmail string

// SystemName is the default name for the system email
var SystemName string

// TestUserEmail is the email for a test user, only used when testing
var TestUserEmail string

// AllowNewProperty enables a new property to be created
var AllowNewProperty bool

// AllowDeleteProperty enables a property to be deleted
var AllowDeleteProperty bool

// ImportFileName has the name of a file to import
var ImportFileName string

// SendMailDisabled when true disables sending emails
var SendMailDisabled bool

// AllowCrossDomainRequests enables cross domain requests (ex. development server)
var AllowCrossDomainRequests bool

// TrialDuration is the duration for trial accounts, deleted at the end of the trial
var TrialDuration time.Duration

func init() {
	SystemEmail = os.Getenv("DEFAULT_SYSTEM_EMAIL")
	SystemName = os.Getenv("DEFAULT_SYSTEM_NAME")
	TestUserEmail = os.Getenv("TEST_USER_EMAIL")
	AllowNewProperty = os.Getenv("ALLOW_NEW_PROPERTY") == "true"
	AllowDeleteProperty = os.Getenv("ALLOW_DELETE_PROPERTY") == "true"
	ImportFileName = os.Getenv("IMPORT_FILE_NAME")
	SendMailDisabled = os.Getenv("SEND_MAIL_DISABLED") == "true"
	AllowCrossDomainRequests = os.Getenv("ALLOW_CROSS_DOMAIN_REQUESTS") == "true"
	value := os.Getenv("TRIAL_DURATION")
	if value == "" {
		// default 0
		TrialDuration, _ = time.ParseDuration("0h")
	} else {
		TrialDuration, _ = time.ParseDuration(value)
	}
}
