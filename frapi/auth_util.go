package frapi

import (
	"context"
	"strings"
)

// User represents a user of the application.
type User struct {
	Email string
	Admin bool
}

var testUserEmail string

// GetUser retrieves the email address of the logged in user
func GetUser(ctx context.Context) *User {
	if testUserEmail != "" {
		lowerCaseEmail := strings.ToLower(strings.TrimSpace(testUserEmail))
		Logger.LogDebugf("Found unit test user: %+v", lowerCaseEmail)
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
	Logger.LogWarningf("LogoutURL deprecated")

	url := "http://localhost:8080/deprecated"

	return &url, nil
}

// LoginURL returns a URL the client can use to login
func (r *Resolver) LoginURL(ctx context.Context, args *struct {
	Dest string
}) (*string, error) {
	url := destinationURI + "/login"
	return &url, nil
}
