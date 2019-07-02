package models

import "github.com/bjorge/friendlyreservations/templates"

// NewNotificationInput is created by the service, i.e. it is not a request from the client gql
type NewNotificationInput struct {
	AllNotifiedUserIds []string
	ToUserIds          []string
	CcUserIds          []string
	TemplateName       templates.TemplateName
	TemplateVersion    int32
	DefaultTemplate    bool
	TemplateParamData  map[templates.TemplateParamGroup]string
	NotificationId     string
	CreateDateTime     string
	AuthorUserId       string
	EventVersion       int32
	EmailSent          bool
}

// GetEventVersion returns the version of the mutation event
func (r *NewNotificationInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *NewNotificationInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// BUG(bjorge): read notification code has not been completed

type NotificationReadInput struct {
	// Fields received from the client
	ForVersion     int32
	NotificationId string

	// Extra fields persisted with the above
	CreateDateTime string
	AuthorUserId   string
	EventVersion   int32
}

// GetEventVersion returns the version of the mutation event
func (r *NotificationReadInput) GetEventVersion() int {
	return int(r.EventVersion)
}

// SetEventVersion is called by the persist code to set the event version
func (r *NotificationReadInput) SetEventVersion(Version int) {
	r.EventVersion = int32(Version)
}

// GetForVersion is called for duplicate suppression
func (r *NotificationReadInput) GetForVersion() int32 {
	return r.ForVersion
}
