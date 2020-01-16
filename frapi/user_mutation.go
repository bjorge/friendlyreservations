package frapi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// CreateUser is the gql call to create a new user record
func (r *Resolver) CreateUser(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.NewUserInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Create User")

	// get the current property
	propertyResolver, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, args.Input, propertyResolver); duplicate || err != nil {
		if err == nil {
			return propertyResolver, nil
		}
		return nil, err
	}

	// check the input values

	// trim the input
	args.Input.Email = strings.ToLower(strings.TrimSpace(args.Input.Email))
	args.Input.Nickname = strings.TrimSpace(args.Input.Nickname)

	constraints, err := propertyResolver.UpdateUserConstraints(ctx, &UpdateUserConstraintsArgs{})
	if err != nil {
		return nil, err
	}

	if len(args.Input.Nickname) < int(constraints.NicknameMin()) {
		return nil, errors.New("nickname too short: " + args.Input.Nickname)
	}

	if len(args.Input.Nickname) > int(constraints.NicknameMax()) {
		return nil, errors.New("nickname too long: " + args.Input.Nickname)
	}

	// check if the nickname already exists
	for _, name := range constraints.InvalidNicknames() {
		if args.Input.Nickname == *name {
			return nil, errors.New("nickname already exists: " + args.Input.Nickname)
		}
	}

	if len(args.Input.Email) < int(constraints.EmailMin()) {
		return nil, errors.New("email too short: " + args.Input.Email)
	}

	if len(args.Input.Email) > int(constraints.EmailMax()) {
		return nil, errors.New("email too long: " + args.Input.Email)
	}

	// check if the email already exists
	for _, email := range constraints.InvalidEmails() {
		if args.Input.Email == *email {
			return nil, errors.New("email already exists: " + args.Input.Email)
		}
	}

	// only an admin can add a user
	if !me.IsAdmin() {
		return nil, errors.New("only an admin can create a user")
	}

	// update the request with more information
	args.Input.IsSystem = false
	args.Input.UserId = utilities.NewGUID()
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.AuthorUserId = me.UserID()
	exists, err := PersistedEmailStore.EmailExists(ctx, args.PropertyID, args.Input.Email)
	if err != nil {
		return nil, err
	} else if *exists {
		return nil, errors.New("email already exists: " + args.Input.Email)
	}

	emailID, err := PersistedEmailStore.CreateEmail(ctx, args.PropertyID, args.Input.Email)
	if err != nil {
		return nil, err
	}
	args.Input.EmailId = emailID
	args.Input.State = models.WAITING_ACCEPT
	args.Input.Email = ""

	// persist the event
	return commitChanges(ctx, args.PropertyID, propertyResolver.EventVersion(), args.Input)
}

// UpdateUser is called to update a user record
func (r *Resolver) UpdateUser(ctx context.Context, args *struct {
	PropertyID string
	UserID     string
	Input      *models.UpdateUserInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Update User")

	propertyResolver, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, args.Input, propertyResolver); duplicate || err != nil {
		if err == nil {
			return propertyResolver, nil
		}
		return nil, err
	}

	// only an admin can update a user
	if !me.IsAdmin() {
		return nil, errors.New("only an admin can update a user")
	}

	users := propertyResolver.Users(&usersArgs{UserID: &args.UserID})
	if len(users) == 0 {
		return nil, errors.New("target user does not exist")
	}

	// trim the input
	args.Input.Email = strings.ToLower(strings.TrimSpace(args.Input.Email))
	args.Input.Nickname = strings.TrimSpace(args.Input.Nickname)

	constraints, err := propertyResolver.UpdateUserConstraints(ctx, &UpdateUserConstraintsArgs{UserID: &args.UserID})
	if err != nil {
		return nil, err
	}

	if len(args.Input.Nickname) < int(constraints.NicknameMin()) {
		return nil, errors.New("nickname too short: " + args.Input.Nickname)
	}

	if len(args.Input.Nickname) > int(constraints.NicknameMax()) {
		return nil, errors.New("nickname too long: " + args.Input.Nickname)
	}

	// check if the nickname already exists
	for _, name := range constraints.InvalidNicknames() {
		if args.Input.Nickname == *name {
			return nil, errors.New("nickname already exists: " + args.Input.Nickname)
		}
	}

	if len(args.Input.Email) < int(constraints.EmailMin()) {
		return nil, errors.New("email too short: " + args.Input.Email)
	}

	if len(args.Input.Email) > int(constraints.EmailMax()) {
		return nil, errors.New("email too long: " + args.Input.Email)
	}

	// check if the email already exists
	for _, email := range constraints.InvalidEmails() {
		if args.Input.Email == *email {
			return nil, errors.New("email already exists: " + args.Input.Email)
		}
	}

	if !args.Input.IsMember && !args.Input.IsAdmin {
		return nil, errors.New("must be an admin or a member")
	}

	user := users[0]

	if user.State() != args.Input.State {
		return nil, errors.New("cannot change user state yet from client")
	}

	// update the request with more information
	args.Input.IsSystem = false
	args.Input.UserId = args.UserID
	args.Input.UpdateDateTime = frdate.CreateDateTimeUTC()
	args.Input.AuthorUserId = me.UserID()

	if user.Email() == args.Input.Email {
		args.Input.EmailId = user.emailID()
		// do not store the actual email, only the email id
		args.Input.Email = ""
	} else {
		exists, err := PersistedEmailStore.EmailExists(ctx, args.PropertyID, args.Input.Email)
		if err != nil {
			return nil, err
		}

		if *exists {
			emailID, err := PersistedEmailStore.GetEmail(ctx, args.PropertyID, args.Input.Email)
			if err != nil {
				return nil, err
			}
			args.Input.EmailId = emailID
			args.Input.Email = ""
			args.Input.State = models.WAITING_ACCEPT
		} else {
			emailID, err := PersistedEmailStore.CreateEmail(ctx, args.PropertyID, args.Input.Email)
			if err != nil {
				return nil, err
			}
			args.Input.EmailId = emailID
			args.Input.Email = ""
			args.Input.State = models.WAITING_ACCEPT
		}
	}

	// persist the event
	return commitChanges(ctx, args.PropertyID, propertyResolver.EventVersion(), args.Input)
}

// AcceptInvitation is called when a user accepts an invitation to participate in a property
func (r *Resolver) AcceptInvitation(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.AcceptInvitationInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Accept Invitation")

	// get the current property
	propertyResolver, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, args.Input, propertyResolver); duplicate || err != nil {
		if err == nil {
			return propertyResolver, nil
		}
		return nil, err
	}

	if me.State() != models.WAITING_ACCEPT {
		return nil, fmt.Errorf("can only accept/decline from WAITING_ACCEPT, not from: %+v", me.State())
	}

	// update the request with more information
	args.Input.UpdateDateTime = frdate.CreateDateTimeUTC()
	args.Input.AuthorUserId = me.UserID()

	// persist the event
	return commitChanges(ctx, args.PropertyID, propertyResolver.EventVersion(), args.Input)
}
