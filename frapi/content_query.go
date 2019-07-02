package frapi

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/templates"
)

const contentGQL = `
type Content {
	name: ContentName!
	rendered: String!
	template: String!
	comment: String!
	createDateTime: String!
	author: User!
	defaultTemplate: String!
	default: Boolean!
}

enum ContentName {
	ADMIN_HOME
	MEMBER_HOME
}
`

// Contents is called by the GQL framework to retrieve all the display content strings
func (r *PropertyResolver) Contents() ([]*ContentResolver, error) {

	r.rollupContent()

	// default home page content
	adminHome := &ContentResolver{name: models.ADMIN_HOME, property: r}
	memberHome := &ContentResolver{name: models.MEMBER_HOME, property: r}

	// get rollups
	ifaces := r.getRollups(&rollupArgs{}, contentsRollupType)
	for _, iface := range ifaces {
		rollup := iface.(*ContentRollup)
		if rollup.Input.Name == models.ADMIN_HOME {
			adminHome.rollup = rollup
		} else {
			memberHome.rollup = rollup
		}
	}

	l := []*ContentResolver{}
	l = append(l, adminHome)
	l = append(l, memberHome)

	return l, nil
}

// ContentResolver resolves a single content string
type ContentResolver struct {
	rollup   *ContentRollup
	name     models.ContentName
	property *PropertyResolver
}

// Name is called by the GQL framework, see contentGQL
func (r *ContentResolver) Name() models.ContentName {
	return r.name
}

// Rendered is called by the GQL framework, see contentGQL
func (r *ContentResolver) Rendered() (string, error) {
	// get defaults
	member, admin, templateParamGroups := templates.GetNotificationTemplate(templates.CurrentTemplateVersion, templates.HomePageContents)

	templateText := member
	if r.name == models.ADMIN_HOME {
		templateText = admin
	}

	// override default if set by user
	if r.rollup != nil {
		templateText = r.rollup.Input.Template
	}

	template, err := template.New("").Parse(templateText)
	if err != nil {
		return "", err
	}

	// the params to pass into the template
	paramsMap := make(map[string]interface{})
	me, err := r.property.Me()
	if err != nil {
		return "", err
	}

	for _, paramGroupName := range templateParamGroups {
		switch paramGroupName {
		case templates.Me:
			paramsMap[string(paramGroupName)] = me

		case templates.Ledger:
			last := int32(1)
			userID := me.UserID()
			userRecords, _ := r.property.Ledgers(&ledgersArgs{UserID: &userID, Last: &last})

			ledgers := userRecords[0].Records()

			// ledgers, err := r.property.Ledgers(&struct {
			// 	UserId  string
			// 	Last    *int32
			// 	Reverse *bool
			// }{
			// 	UserId: me.UserId(),
			// 	Last:   &last,
			// })
			// if err != nil {
			// 	return "", err
			// }
			if len(ledgers) != 1 {
				return "", errors.New("expected a ledger")
			}

			paramsMap[string(paramGroupName)] = ledgers[0]
		case templates.Settings:
			settings, err := r.property.Settings(&settingsArgs{})
			if err != nil {
				return "", err
			}
			paramsMap[string(paramGroupName)] = settings

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

// Template is called by the GQL framework, see contentGQL
func (r *ContentResolver) Template() string {
	if r.rollup != nil {
		return r.rollup.Input.Template
	}
	return r.DefaultTemplate()
}

// Comment is called by the GQL framework, see contentGQL
func (r *ContentResolver) Comment() string {
	if r.rollup != nil {
		return r.rollup.Input.Comment
	}
	return ""
}

// CreateDateTime is called by the GQL framework, see contentGQL
func (r *ContentResolver) CreateDateTime() string {
	if r.rollup != nil {
		return r.rollup.Input.CreateDateTime
	}
	return frdate.CreateDateTimeUTC()
}

// Author is called by the GQL framework, see contentGQL
func (r *ContentResolver) Author() (*UserResolver, error) {
	if r.rollup != nil {
		users := r.property.Users(&usersArgs{UserID: &r.rollup.Input.AuthorUserId})
		if len(users) > 0 {
			return users[0], nil
		}
	} else {
		users := r.property.Users(&usersArgs{})
		for _, user := range users {
			if user.IsSystem() {
				return user, nil
			}
		}
	}

	return nil, fmt.Errorf("user not found")
}

// DefaultTemplate is called by the GQL framework, see contentGQL
func (r *ContentResolver) DefaultTemplate() string {
	member, admin, _ := templates.GetNotificationTemplate(templates.CurrentTemplateVersion, templates.HomePageContents)
	switch r.name {
	case models.ADMIN_HOME:
		return admin
	case models.MEMBER_HOME:
		return member
	}
	return ""
}

// Default is called by the GQL framework, see contentGQL
func (r *ContentResolver) Default() bool {
	if r.rollup != nil {
		return false
	}
	return true
}
