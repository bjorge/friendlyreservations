package utilities

import (
	"time"

	"github.com/bjorge/friendlyreservations/config"
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
	SystemEmail = config.GetConfig("DEFAULT_SYSTEM_EMAIL")
	SystemName = config.GetConfig("DEFAULT_SYSTEM_NAME")
	TestUserEmail = config.GetConfig("TEST_USER_EMAIL")
	AllowNewProperty = config.GetConfig("ALLOW_NEW_PROPERTY") == "true"
	AllowDeleteProperty = config.GetConfig("ALLOW_DELETE_PROPERTY") == "true"
	ImportFileName = config.GetConfig("IMPORT_FILE_NAME")
	SendMailDisabled = config.GetConfig("SEND_MAIL_DISABLED") == "true"
	AllowCrossDomainRequests = config.GetConfig("ALLOW_CROSS_DOMAIN_REQUESTS") == "true"
	value := config.GetConfig("TRIAL_DURATION")
	if value == "" {
		// default 0
		TrialDuration, _ = time.ParseDuration("0h")
	} else {
		TrialDuration, _ = time.ParseDuration(value)
	}
}
