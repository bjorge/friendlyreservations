package frapi

import (
	"context"
	"strings"

	"github.com/bjorge/friendlyreservations/utilities"
)

// User represents a user of the application.
type User struct {
	Email string
	Admin bool
}

// GetUser retrieves the email address of the logged in user
func GetUser(ctx context.Context) *User {
	if utilities.TestUserEmail != "" {
		lowerCaseEmail := strings.ToLower(strings.TrimSpace(utilities.TestUserEmail))
		Logger.LogDebugf("Found env user: %+v", lowerCaseEmail)
		return &User{Email: lowerCaseEmail, Admin: true}
	}

	email, admin, _ := FrapiCookies.GetContextValues(ctx)
	if email != "" {
		lowerCaseEmail := strings.ToLower(strings.TrimSpace(email))
		Logger.LogDebugf("Found real user: %+v", lowerCaseEmail)
		return &User{Email: lowerCaseEmail, Admin: admin}
	}

	return nil
}

// LogoutURL returns a URL the client can use to logout
func LogoutURL(ctx context.Context, dest string) (string, error) {
	currentUser := GetUser(ctx)
	if currentUser != nil {
		Logger.LogDebugf("Logout request for user: %+v", currentUser.Email)
	} else {
		Logger.LogDebugf("Logout request with not current user")
	}

	Logger.LogWarningf("LogoutURL not implemented")

	return "http://localhost:8080/needtoimplement", nil
	// // propably hard to have no current user since protected yaml setting requires
	// // a user for this method...
	// url, err := user.LogoutURL(ctx, dest)
	// return url, err
}
