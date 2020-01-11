package frapi

import (
	"context"
	"strings"

	"github.com/bjorge/friendlyreservations/utilities"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// GetUser retrieves the email address of the logged in user
func GetUser(ctx context.Context) *user.User {
	if utilities.TestUserEmail != "" {
		lowerCaseEmail := strings.ToLower(strings.TrimSpace(utilities.TestUserEmail))
		testUser := &user.User{Email: lowerCaseEmail}
		Logger.LogDebugf("Found env user: %+v", testUser.Email)
		return testUser
	}

	if realUser := user.Current(ctx); realUser != nil {
		copy := *realUser
		copy.Email = strings.ToLower(strings.TrimSpace(realUser.Email))
		Logger.LogDebugf("Found real user: %+v", copy.Email)
		return &copy
	}

	return nil

}

// LogoutURL returns a URL the client can use to logout
func LogoutURL(ctx context.Context, dest string) (string, error) {
	currentUser := GetUser(ctx)
	if currentUser != nil {
		log.Debugf(ctx, "Logout request for user: %+v", currentUser.Email)
	} else {
		log.Debugf(ctx, "Logout request with not current user")
	}
	// propably hard to have no current user since protected yaml setting requires
	// a user for this method...
	url, err := user.LogoutURL(ctx, dest)
	return url, err
}
