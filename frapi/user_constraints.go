package frapi

import (
	"context"

	"github.com/bjorge/friendlyreservations/utilities"
)

const updateUserConstraintsGQL = `

type UpdateUserConstraints {
	nicknameMin: Int!
	nicknameMax: Int!
	emailMin: Int!
	emailMax: Int!
	invalidNicknames: [String]!
	invalidEmails: [String]!
}
`

// UpdateUserConstraints provides the retriever for contraint methods
type UpdateUserConstraints struct {
	args  *UpdateUserConstraintsArgs
	users []*UserResolver
}

// UpdateUserConstraintsArgs holds the arguments for the UpdateUserConstraints GQL call
type UpdateUserConstraintsArgs struct {
	UserID *string
}

// UpdateUserConstraints returns input constraints for updating user
func (r *PropertyResolver) UpdateUserConstraints(ctx context.Context, args *UpdateUserConstraintsArgs) (*UpdateUserConstraints, error) {

	users := r.Users(&usersArgs{})
	return &UpdateUserConstraints{args, users}, nil
}

// UpdateUserConstraints returns input constraints for updating user
func (r *Resolver) UpdateUserConstraints() (*UpdateUserConstraints, error) {

	return &UpdateUserConstraints{}, nil
}

// NicknameMin returns min length of nickname
func (r *UpdateUserConstraints) NicknameMin() int32 { return 3 }

// NicknameMax returns max length of nickname
func (r *UpdateUserConstraints) NicknameMax() int32 { return 25 }

// EmailMin returns min length of email string
func (r *UpdateUserConstraints) EmailMin() int32 { return 3 }

// EmailMax returns max length of email string
func (r *UpdateUserConstraints) EmailMax() int32 { return 35 }

// InvalidNicknames is a list of all the nicknames that cannot be used
func (r *UpdateUserConstraints) InvalidNicknames() []*string {
	names := []*string{}
	if r.users == nil || len(r.users) == 0 {
		names = append(names, &utilities.SystemName)
	} else if r.args.UserID == nil {
		for _, user := range r.users {
			name := user.Nickname()
			names = append(names, &name)
		}
	} else {
		for _, user := range r.users {
			if user.UserID() != *r.args.UserID {
				name := user.Nickname()
				names = append(names, &name)
			}
		}
	}
	return names
}

// InvalidEmails is the list of emails that cannot be used
func (r *UpdateUserConstraints) InvalidEmails() []*string {
	emails := []*string{}
	if r.users == nil || len(r.users) == 0 {
		emails = append(emails, &utilities.SystemEmail)
	} else if r.args.UserID == nil {
		for _, user := range r.users {
			email := user.Email()
			emails = append(emails, &email)
		}
	} else {
		for _, user := range r.users {
			if user.UserID() != *r.args.UserID {
				email := user.Email()
				emails = append(emails, &email)
			}
		}
	}
	return emails
}
