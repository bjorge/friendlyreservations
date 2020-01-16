package frapi

import (
	"strings"

	"github.com/bjorge/friendlyreservations/models"
)

const userGQL = `

enum UserState {
	WAITING_ACCEPT
	ACCEPTED
	DISABLED
	DECLINED
}

type User {
	# the user id
	userId: String!
	# the member email
	email: String!
	# whether this user is an admin
	isAdmin: Boolean!
	# whether this user is a member
	isMember: Boolean!
	# whether this user is a system user
	isSystem: Boolean!
	# the current state of the user
	state: UserState!
	# the name of the user
	nickname: String!
}
`

type usersArgs struct {
	UserID     *string
	Email      *string
	MaxVersion *int32
}

// Users is the gql call to retrieve users for the property
func (r *PropertyResolver) Users(args *usersArgs) []*UserResolver {

	// rollup user events
	r.rollupUsers()

	// run through the common filters
	var l []*UserResolver
	ifaces := r.getRollups(&rollupArgs{id: args.UserID, maxVersion: args.MaxVersion}, userRollupType)
	for _, iface := range ifaces {
		resolver := &UserResolver{}
		resolver.rollup = iface.(*UserRollup)
		resolver.property = r
		resolver.args = args
		l = append(l, resolver)
	}

	// add the email filter
	if args.Email != nil {
		lowerCaseEmail := strings.ToLower(strings.TrimSpace(*args.Email))
		args.Email = &lowerCaseEmail
		emailList := []*UserResolver{}
		for _, resolver := range l {
			if resolver.Email() == *args.Email {
				emailList = append(emailList, resolver)
			}
		}
		l = emailList
	}

	return l
}

// UserResolver is the receiver for user resolver calls
type UserResolver struct {
	rollup   *UserRollup
	property *PropertyResolver
	args     *usersArgs
}

// UserID is the id of the user
func (r *UserResolver) UserID() string {
	return r.rollup.UserID
}

// Email is the email of the user
func (r *UserResolver) Email() string {
	email := r.property.property.EmailMap[r.rollup.EmailID]

	return email
}

// IsAdmin is true if the user is an admin
func (r *UserResolver) IsAdmin() bool {
	return r.rollup.IsAdmin
}

// IsMember is true if the user is a member, i.e. can make reservations
func (r *UserResolver) IsMember() bool {
	return r.rollup.IsMember
}

// IsSystem is true if the user is a system user
func (r *UserResolver) IsSystem() bool {
	return r.rollup.IsSystem
}

// State is the user state
func (r *UserResolver) State() models.UserState {
	return r.rollup.State
}

// Nickname is the user nickname
func (r *UserResolver) Nickname() string {
	return r.rollup.Nickname
}

func (r *UserResolver) emailID() string {
	return r.rollup.EmailID
}

// GetEventVersion returns the version of this user record
func (r *UserResolver) GetEventVersion() int {
	return int(r.rollup.EventVersion)
}
