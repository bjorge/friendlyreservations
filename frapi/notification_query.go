package frapi

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"sort"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/templates"
)

const notificationGQL = `
type Notification {
	emailSent: Boolean!
	to: [User]!
	cc: [User]!
	subject: String!
	body: String!
	createDateTime: String!
	author: User!
	notificationId: String!
	templateName: String!
	templateVersion: Int!
	defaultTemplate: Boolean!
	read: Boolean!
}
`

type notificationArgs struct {
	UserID         *string
	Reverse        *bool
	notificationID *string
}

// return the last notification date for each notification type and user
// (internal call)
func (r *PropertyResolver) lastNotifications(dateBuilder *frdate.DateBuilder) map[templates.TemplateName]map[string]*frdate.DateTime {

	// rollup
	r.rollupNotifications()

	// get the rollups
	ifaces := r.getRollups(&rollupArgs{}, notificationRollupType)

	sort.Slice(ifaces, func(i, j int) bool {
		return ifaces[i].GetEventVersion() < ifaces[j].GetEventVersion()
	})

	// check when last notifications of each type were sent
	notificationsMap := make(map[templates.TemplateName]map[string]*frdate.DateTime)
	for _, iface := range ifaces {
		rollup := iface.(*NotificationRollup)
		name := rollup.Input.TemplateName
		//time := rollup.Input.CreateDateTime
		userIds := rollup.Input.ToUserIds
		for _, userID := range userIds {
			if _, ok := notificationsMap[name]; !ok {
				notificationsMap[name] = make(map[string]*frdate.DateTime)
			}
			dateTime := dateBuilder.MustNewDateTime(rollup.Input.CreateDateTime)
			notificationsMap[name][userID] = dateTime
		}
	}

	return notificationsMap
}

// Notifications is called to return the list of all past notifications
func (r *PropertyResolver) Notifications(args *notificationArgs) ([]*NotificationResolver, error) {
	var l []*NotificationResolver

	// rollup
	r.rollupNotifications()

	ifaces := r.getRollups(&rollupArgs{id: args.notificationID}, notificationRollupType)
	for _, iface := range ifaces {
		resolver := &NotificationResolver{}
		resolver.property = r
		resolver.args = args
		resolver.rollup = iface.(*NotificationRollup)
		l = append(l, resolver)
	}

	if args.UserID != nil {
		userNotifications := []*NotificationResolver{}
		for _, resolver := range l {
			if _, ok := resolver.rollup.TargetUserIdsMap[*args.UserID]; ok {
				userNotifications = append(userNotifications, resolver)
			}
		}
		l = userNotifications
	}

	reverseOrder := false
	if args.Reverse != nil {
		reverseOrder = *args.Reverse
	}

	sort.Slice(l, func(i, j int) bool {
		if reverseOrder {
			return l[i].GetEventVersion() > l[j].GetEventVersion()
		}
		return l[i].GetEventVersion() < l[j].GetEventVersion()
	})

	return l, nil
}

// NotificationResolver resolves a single notification
type NotificationResolver struct {
	rollup   *NotificationRollup
	property *PropertyResolver
	args     *notificationArgs
}

// NotificationID is the id of the notification
func (r *NotificationResolver) NotificationID() string {
	return r.rollup.Input.NotificationId
}

// GetEventVersion is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) GetEventVersion() int {
	return int(r.rollup.EventVersion)
}

// To is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) To() []*UserResolver {
	l := []*UserResolver{}
	for _, userID := range r.rollup.Input.ToUserIds {
		userResolvers := r.property.Users(&usersArgs{
			UserID:     &userID,
			MaxVersion: &r.rollup.Input.EventVersion,
		})
		if len(userResolvers) > 0 {
			l = append(l, userResolvers[0])
		}
	}

	return l
}

func (r *NotificationResolver) toEmailFormat() []string {
	l := []string{}
	for _, userID := range r.rollup.Input.ToUserIds {
		userResolvers := r.property.Users(&usersArgs{
			UserID:     &userID,
			MaxVersion: &r.rollup.Input.EventVersion,
		})
		if len(userResolvers) > 0 {
			user := userResolvers[0]
			email := fmt.Sprintf("%s <%s>", user.Nickname(), user.Email())
			l = append(l, email)
		}
	}

	return l
}

// Cc is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) Cc() []*UserResolver {
	l := []*UserResolver{}
	for _, userID := range r.rollup.Input.CcUserIds {
		userResolvers := r.property.Users(&usersArgs{
			UserID:     &userID,
			MaxVersion: &r.rollup.Input.EventVersion,
		})
		if len(userResolvers) > 0 {
			l = append(l, userResolvers[0])
		}
	}

	return l
}

func (r *NotificationResolver) ccEmailFormat() []string {
	l := []string{}
	for _, userID := range r.rollup.Input.CcUserIds {
		userResolvers := r.property.Users(&usersArgs{
			UserID:     &userID,
			MaxVersion: &r.rollup.Input.EventVersion,
		})
		if len(userResolvers) > 0 {
			user := userResolvers[0]
			email := fmt.Sprintf("%s <%s>", user.Nickname(), user.Email())
			l = append(l, email)
		}
	}

	return l
}

func (r *NotificationResolver) templateHelper(subject bool) (string, error) {

	subjectTemplate, bodyTemplate, templateParamGroups := templates.GetNotificationTemplate(int(r.rollup.Input.TemplateVersion), r.rollup.Input.TemplateName)
	templateText := subjectTemplate
	if !subject {
		templateText = bodyTemplate
	}

	template, err := template.New("").Parse(templateText)
	if err != nil {
		return "", err
	}

	// the params to pass into the template
	paramsMap := make(map[string]interface{})

	//for paramName, paramValue := range r.notification.Input.TemplateParamData {
	for _, paramGroupName := range templateParamGroups {
		switch paramGroupName {
		case templates.Settings:
			settings, err := r.property.Settings(&settingsArgs{MaxVersion: &r.rollup.Input.EventVersion})
			if err != nil {
				return "", err
			}
			paramsMap[string(paramGroupName)] = settings

		case templates.Reservation:
			reservationID := r.rollup.Input.TemplateParamData[templates.Reservation]
			reservations, err := r.property.Reservations(&reservationsArgs{
				MaxVersion:    &r.rollup.Input.EventVersion,
				ReservationID: &reservationID,
			})

			if err != nil {
				return "", err
			}
			paramsMap[string(paramGroupName)] = reservations[0]

		case templates.Ledger:
			userID := r.rollup.Input.TemplateParamData[templates.Ledger]
			last := int32(1)
			userRecords, _ := r.property.Ledgers(&ledgersArgs{UserID: &userID, Last: &last, MaxVersion: &r.rollup.Input.EventVersion})

			userRecord := userRecords[0]
			ledgers := userRecord.Records()

			if len(ledgers) != 1 {
				return "", errors.New("expected a ledger")
			}

			ledger := ledgers[0]

			paramsMap[string(paramGroupName)] = ledger

			paramsMap["User"] = userRecord.User()

		case templates.Decimal:
			paramsMap[string(paramGroupName)] = &struct{ Format amountFormat }{Format: decimal}

		}
	}

	var buffer bytes.Buffer
	if err := template.Execute(&buffer, paramsMap); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// Subject is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) Subject() (string, error) {
	return r.templateHelper(true)
}

// Body is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) Body() (string, error) {
	return r.templateHelper(false)
}

// TemplateName is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) TemplateName() string {
	return string(r.rollup.Input.TemplateName)
}

// TemplateVersion is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) TemplateVersion() int32 {
	return r.rollup.Input.TemplateVersion
}

// DefaultTemplate is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) DefaultTemplate() bool {
	return r.rollup.Input.DefaultTemplate
}

// Read is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) Read() bool {
	// only return true if a user is targeted, otherwise all the notifications are returned
	// (example in an admin call)
	if r.args.UserID == nil {
		return false
	}
	if _, ok := r.rollup.ReaderUserIdsMap[*r.args.UserID]; ok {
		return true
	}
	return false
}

// Author is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) Author() *UserResolver {
	userResolvers := r.property.Users(&usersArgs{
		UserID:     &r.rollup.Input.AuthorUserId,
		MaxVersion: &r.rollup.Input.EventVersion,
	})
	if len(userResolvers) > 0 {
		return userResolvers[0]
	}
	return nil
}

func (r *NotificationResolver) authorEmailFormat() *string {
	userResolvers := r.property.Users(&usersArgs{
		UserID:     &r.rollup.Input.AuthorUserId,
		MaxVersion: &r.rollup.Input.EventVersion,
	})
	if len(userResolvers) > 0 {
		user := userResolvers[0]
		email := fmt.Sprintf("%s <%s>", user.Nickname(), user.Email())
		return &email
	}
	return nil
}

// EmailSent is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) EmailSent() bool {
	return r.rollup.Input.EmailSent
}

// CreateDateTime is called by the GQL framework, see notificationGQL
func (r *NotificationResolver) CreateDateTime() string {
	return r.rollup.Input.CreateDateTime
}
