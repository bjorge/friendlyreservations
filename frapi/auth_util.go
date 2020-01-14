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

// LogoutURL returns a url used by the client to logout
func (r *Resolver) LogoutURL(ctx context.Context, args *struct {
	Dest string
}) (*string, error) {
	// url, err := LogoutURL(ctx, args.Dest)
	// return &url, err
	currentUser := GetUser(ctx)
	if currentUser != nil {
		Logger.LogDebugf("Logout request for user: %+v", currentUser.Email)
	} else {
		Logger.LogDebugf("Logout request with no current user")
	}

	Logger.LogWarningf("LogoutURL not implemented")

	url := "http://localhost:8080/needtoimplement"

	return &url, nil
}

// LoginURL returns a URL the client can use to login
func (r *Resolver) LoginURL(ctx context.Context, args *struct {
	Dest string
}) (*string, error) {
	currentUser := GetUser(ctx)
	if currentUser != nil {
		Logger.LogDebugf("Logout request for user: %+v", currentUser.Email)
	} else {
		Logger.LogDebugf("Logout request with no current user")
	}

	url := destinationURI + "/login"

	return &url, nil
}
