package frapi

import (
	"context"
	"errors"
	"strconv"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/persist"
	"github.com/bjorge/friendlyreservations/templates"
	"github.com/bjorge/friendlyreservations/utilities"
)

type deletePropertyArgs struct {
	PropertyID string
}

// DeleteProperty is called from gql to delete a property
func (r *Resolver) DeleteProperty(ctx context.Context, args *deletePropertyArgs) (bool, error) {

	if !utilities.AllowDeleteProperty {
		return false, errors.New("Delete property not allowed")
	}

	// get the current property
	_, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return false, err
	}

	// check the input values
	if !me.IsAdmin() {
		return false, errors.New("only admins can delete the property")
	}

	return r.internalDeleteProperty(ctx, args.PropertyID)
}

// internalDeleteProperty is called from DeleteProperty and the cron job
func (r *Resolver) internalDeleteProperty(ctx context.Context, propertyID string) (bool, error) {

	err := persistedVersionedEvents.DeleteProperty(ctx, propertyID, persistedPropertyList)
	if err != nil {
		return false, err
	}
	err = persistedEmailStore.DeleteEmails(ctx, propertyID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateProperty is called from gql to create a property
func (r *Resolver) CreateProperty(ctx context.Context, args *struct{ Input *models.NewPropertyInput }) (*PropertyResolver, error) {

	// check that a user is logged in
	u := utilities.GetUser(ctx)
	if u == nil {
		return nil, errors.New("user not logged in")
	}

	constraints, err := r.UpdateSettingsConstraints(ctx)
	if err != nil {
		return nil, err
	}

	if !constraints.AllowNewProperty() {
		return nil, errors.New("create property not allowed")
	}

	if u.Email == utilities.SystemEmail {
		return nil, errors.New("cannot create a property with system email")
	}

	// check input parameters
	if args.Input.PropertyName == "" {
		return nil, errors.New("propery name is missing")
	}
	if args.Input.NickName == "" {
		return nil, errors.New("nickname is missing")
	}

	if _, err := frdate.NewDateBuilder(args.Input.Timezone); err != nil {
		return nil, err
	}

	// create the first events
	nextPropertyTransactionKey, err := persistedPropertyList.GetNextVersion(ctx)
	if err != nil {
		return nil, err
	}
	propertyID := strconv.Itoa(nextPropertyTransactionKey)
	//propertyId := utilities.NewGuid()

	userID := utilities.NewGUID()
	args.Input.PropertyId = propertyID
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.AuthorUserId = userID
	firstEvents, err := CreateNewPropertyEvents(ctx, u.Email, userID, args.Input)

	_, err = persistedVersionedEvents.CreateProperty(ctx, propertyID, firstEvents, persistedPropertyList, nextPropertyTransactionKey)
	if err != nil {
		return nil, err
	}

	// get all events for the property
	propertyResolver, _, err := currentProperty(ctx, propertyID)
	if err != nil {
		return nil, err
	}

	return propertyResolver, nil
}

// CreateNewPropertyEvents sets up a new property with its first default events
func CreateNewPropertyEvents(ctx context.Context, email string, userID string, newProperty *models.NewPropertyInput) ([]persist.VersionedEvent, error) {

	// default version
	newVersion := &models.NewVersionEvent{Version: 1}

	emailRecord, err := persistedEmailStore.CreateEmail(ctx, newProperty.PropertyId, email)
	if err != nil {
		return nil, err
	}

	newUser := &models.NewUserInput{Nickname: newProperty.NickName, EmailId: emailRecord.EmailID, UserId: userID,
		State: models.ACCEPTED, IsAdmin: true, IsMember: newProperty.IsMember, IsSystem: false, CreateDateTime: frdate.CreateDateTimeUTC(),
		AuthorUserId: userID, Email: ""}

	emailRecordSystemUser, err := persistedEmailStore.CreateEmail(ctx, newProperty.PropertyId, utilities.SystemEmail)
	if err != nil {
		return nil, err
	}

	userIDSystem := utilities.NewGUID()

	newSystemUser := &models.NewUserInput{Nickname: "Friendly Reservations", EmailId: emailRecordSystemUser.EmailID, UserId: userIDSystem,
		State: models.ACCEPTED, IsAdmin: false, IsMember: false, IsSystem: true, CreateDateTime: frdate.CreateDateTimeUTC(),
		AuthorUserId: userID, Email: ""}

	newNotification := &models.NewNotificationInput{}
	newNotification.AuthorUserId = userIDSystem
	newNotification.CcUserIds = []string{}
	newNotification.ToUserIds = []string{userID}
	newNotification.TemplateName = templates.NewPropertyNotification
	newNotification.CreateDateTime = frdate.CreateDateTimeUTC()
	newNotification.NotificationId = utilities.NewGUID()
	newNotification.TemplateVersion = int32(templates.CurrentTemplateVersion)
	newNotification.DefaultTemplate = true
	newNotification.TemplateParamData = make(map[templates.TemplateParamGroup]string)

	// return all the first property events!
	firstEvents := []persist.VersionedEvent{newVersion, newUser, newSystemUser, newProperty, newNotification}

	return firstEvents, nil
}
