package cookies

/////////////////////

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bjorge/friendlyreservations/config"
	"github.com/bjorge/friendlyreservations/logger"
	"github.com/gorilla/securecookie"
)

type contextKey string

// WriterKey is used for writing to cookies
type WriterKey string

var log = logger.New()

// AuthCookies manages secure cookies used for authentication
// see: https://medium.com/lightrail/getting-token-authentication-right-in-a-stateless-single-page-application-57d0c6474e3
type AuthCookies struct {
	SecureCookie *securecookie.SecureCookie // signed secure cookie implementation
	HTTPOnlyName string                     // http only cookie
	JSName       string                     // cookie viewable from js (i.e. to know if the session is no longer valid)
	Secure       bool                       // https, or in local test mode http
	Duration     time.Duration              // duration for the auth cookies
}

// NewCookies creates a new AuthCookies
func NewCookies() *AuthCookies {
	// load settings from environment
	secure := config.GetConfig("PLATFORM_SECURE") == "true"

	value := config.GetConfig("PLATFORM_AUTH_COOKIE_HASH")
	if value == "" {
		panic("cookie hash not set in environment variable PLATFORM_AUTH_COOKIE_HASH")
	}
	authCookieHash := value

	value = config.GetConfig("PLATFORM_SESSION_DURATION")
	if value == "" {
		panic("session duration not set in environment variable PLATFORM_SESSION_DURATION")
	}
	sessionDuration, err := time.ParseDuration(value)
	if err != nil {
		panic(err)
	}

	// Hash keys should be at least 32 bytes long
	jwtCookie := securecookie.New([]byte(authCookieHash), nil)

	return &AuthCookies{
		SecureCookie: jwtCookie,
		HTTPOnlyName: "httpauth",
		JSName:       "jsauth",
		Secure:       secure,
		Duration:     sessionDuration,
	}
}

// SetCookies sets the cookies for authentication
func (r *AuthCookies) SetCookies(w http.ResponseWriter, email string) {
	log.LogDebugf("SetCookies")

	expiration := time.Now().UTC().Add(r.Duration)

	value := map[string]string{
		"email":      email,
		"expiration": expiration.Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(value)

	//LogDebugf("json session is: %+v", jsonData)

	secureEncoded, err := r.SecureCookie.Encode(r.HTTPOnlyName, jsonData)

	if err == nil {
		httpOnlyCookie := &http.Cookie{
			Name:     r.HTTPOnlyName,
			Value:    secureEncoded,
			Path:     "/",
			Secure:   r.Secure,
			HttpOnly: true,
		}
		http.SetCookie(w, httpOnlyCookie)

		b64Encoded := base64.StdEncoding.EncodeToString(jsonData)

		//LogDebugf("jsAuth cookie value is; %s", b64Encoded)

		jsCookie := &http.Cookie{
			Name:     r.JSName,
			Value:    b64Encoded,
			Path:     "/",
			Secure:   r.Secure,
			HttpOnly: false,
			MaxAge:   int(r.Duration.Seconds()),
		}
		http.SetCookie(w, jsCookie)

	}
}

// ContextWithCookies puts the cookie values into the context
func (r *AuthCookies) ContextWithCookies(ctx context.Context, request *http.Request) context.Context {
	//LogDebugf("ContextWithCookies")
	email, _ := r.GetCookiesValues(request)
	ctxWithValues := context.WithValue(ctx, contextKey("email"), email)

	return ctxWithValues
}

// GetContextValues returns the values of the auth cookies
func (r *AuthCookies) GetContextValues(ctx context.Context) string {
	email := ctx.Value(contextKey("email")).(string)
	return email
}

func (r *AuthCookies) unmarshalCookieValues(request *http.Request, cookieName string, secure bool) (map[string]string, error) {
	//LogDebugf("GetCookiesValue")
	cookieValues := make(map[string]string)
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		log.LogDebugf("User not logged in, cookie %v missing (%v)", cookieName, err.Error())
		return nil, err
	}

	var value []byte
	if secure {
		err := r.SecureCookie.Decode(cookieName, cookie.Value, &value)
		if err != nil {
			log.LogErrorf("Could not decode secure cookie, error: %v", err.Error())
			//LogDebugf("read json http cookie: %+v\n", string(value))
			return nil, err
		}
		err = json.Unmarshal(value, &cookieValues)
		if err != nil {
			log.LogErrorf("Could not unmarshal cookie %v, error: %v", cookieName, err.Error())
		}

	} else {
		value, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err != nil {
			log.LogErrorf("Unexpected base64 decode error, error: %v", err.Error())
			return nil, err
		}
		//LogDebugf("read json js cookie: %+v\n", string(value))
		err = json.Unmarshal(value, &cookieValues)
		if err != nil {
			log.LogErrorf("Could not unmarshal cookie %v, error: %v", cookieName, err.Error())
		}
	}
	return cookieValues, nil
}

// GetCookiesValues returns the common value of the auth cookies
func (r *AuthCookies) GetCookiesValues(request *http.Request) (string, string) {
	//LogDebugf("GetCookiesValue")
	httpCookieValues, err := r.unmarshalCookieValues(request, r.HTTPOnlyName, true)
	if err != nil {
		return "", ""
	}

	jsCookieValues, err := r.unmarshalCookieValues(request, r.JSName, false)
	if err != nil {
		return "", ""
	}

	// BUG(bjorge): check expiration value itself to see if cookie has expired

	if httpCookieValues["email"] == jsCookieValues["email"] &&
		httpCookieValues["expiration"] == jsCookieValues["expiration"] {
		return httpCookieValues["email"], httpCookieValues["expiration"]
	}

	log.LogDebugf("Not authorized, cookies do not match")
	return "", ""
}

// IsContextAuthenticated returns true if authenticated, false otherwise
func (r *AuthCookies) IsContextAuthenticated(ctx context.Context) bool {
	log.LogDebugf("IsContextAuthenticated")
	email := ctx.Value(contextKey("email"))
	if email == "" {
		log.LogDebugf("IsContextAuthenticated: false")
		return false
	}
	return true
}

// IsAuthenticated returns true if authenticated, false otherwise
func (r *AuthCookies) IsAuthenticated(writer http.ResponseWriter, request *http.Request) bool {
	email, _ := r.GetCookiesValues(request)
	if email == "" {
		log.LogDebugf("auth false, no cookie")
		r.ClearCookies(writer)
		return false
	}
	log.LogDebugf("auth true, got cookie with email %+v", email)
	return true
}

// ClearCookies clears the auth cookies
func (r *AuthCookies) ClearCookies(w http.ResponseWriter) {
	log.LogDebugf("ClearCookies")

	httpOnlyCookie := &http.Cookie{
		Name:   r.HTTPOnlyName,
		MaxAge: -1,
	}
	http.SetCookie(w, httpOnlyCookie)

	jsCookie := &http.Cookie{
		Name:   r.JSName,
		MaxAge: -1,
	}
	http.SetCookie(w, jsCookie)
}
