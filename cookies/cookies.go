package cookies

/////////////////////

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
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
func (r *AuthCookies) SetCookies(w http.ResponseWriter, numschi string, isAdmin bool) {
	log.LogDebugf("SetCookies")

	expiration := time.Now().UTC().Add(r.Duration)

	value := map[string]string{
		"isAdmin":    strconv.FormatBool(isAdmin),
		"numschi":    numschi,
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
	numschi, isAdmin, _ := r.GetCookiesValues(request)
	ctxWithValues := context.WithValue(ctx, contextKey("numschi"), numschi)
	ctxWithValues = context.WithValue(ctxWithValues, contextKey("isAdmin"), isAdmin)

	return ctxWithValues
}

// GetContextValues returns the values of the auth cookies
func (r *AuthCookies) GetContextValues(ctx context.Context) (string, bool, string) {
	numschi := ctx.Value(contextKey("numschi")).(string)
	isAdmin := ctx.Value(contextKey("isAdmin")).(bool)
	return numschi, isAdmin, ""
}

// GetCookiesValues returns the common value of the auth cookies
func (r *AuthCookies) GetCookiesValues(request *http.Request) (string, bool, string) {
	//LogDebugf("GetCookiesValue")
	httpCookieValues := make(map[string]string)
	if cookie, err := request.Cookie(r.HTTPOnlyName); err != nil {
		log.LogDebugf("Http cookie missing (user not logged in), error: %v", err.Error())
		return "", false, ""
	} else {
		var value []byte
		if err = r.SecureCookie.Decode(r.HTTPOnlyName, cookie.Value, &value); err == nil {
			//LogDebugf("read json http cookie: %+v\n", string(value))
			json.Unmarshal(value, &httpCookieValues)
		}
	}
	jsCookieValues := make(map[string]string)
	if cookie, err := request.Cookie(r.JSName); err != nil {
		log.LogDebugf("JS cookie missing (user not logged in), error: %v", err.Error())
		return "", false, ""
	} else {
		log.LogDebugf("found a js cookie")
		if value, err := base64.StdEncoding.DecodeString(cookie.Value); err != nil {
			log.LogDebugf("Unexpected JS cookie base64 decode error, error: %v", err.Error())
			return "", false, ""
		} else {
			//LogDebugf("read json js cookie: %+v\n", string(value))
			json.Unmarshal(value, &jsCookieValues)
		}
	}

	// BUG(bjorge): check expiration value itself to see if cookie has expired

	if httpCookieValues["isAdmin"] == jsCookieValues["isAdmin"] &&
		httpCookieValues["numschi"] == jsCookieValues["numschi"] &&
		httpCookieValues["expiration"] == jsCookieValues["expiration"] {
		//LogDebugf("Authorized, cookies match")
		isAdmin, err := strconv.ParseBool(httpCookieValues["isAdmin"])
		if err != nil {
			log.LogErrorf("Unexpected cookie parsing error: %v", err.Error())
			return "", false, ""
		}
		return httpCookieValues["numschi"], isAdmin, httpCookieValues["expiration"]
	}

	log.LogDebugf("Not authorized, cookies do not match")
	return "", false, ""
}

// IsContextAuthenticated returns true if authenticated, false otherwise
func (r *AuthCookies) IsContextAuthenticated(ctx context.Context) bool {
	log.LogDebugf("IsContextAuthenticated")
	numschi := ctx.Value(contextKey("numschi"))
	if numschi == "" {
		log.LogDebugf("IsContextAuthenticated: false")
		return false
	}
	return true
}

// IsAuthenticated returns true if authenticated, false otherwise
func (r *AuthCookies) IsAuthenticated(writer http.ResponseWriter, request *http.Request) bool {
	numschi, _, _ := r.GetCookiesValues(request)
	if numschi == "" {
		log.LogDebugf("auth false, no cookie")
		r.ClearCookies(writer)
		return false
	}
	log.LogDebugf("auth true, got cookie with numschi %+v", numschi)
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

/////////////////////

// import (
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"time"

// 	"github.com/bjorge/friendlyreservations/platform"

// 	"github.com/gorilla/securecookie"
// )

// type contextKey string

// // WriterKey is used for writing to cookies
// type WriterKey string

// // Logger is used for cookie logging
// var Logger platform.Logger

// // AuthCookies manages secure cookies used for authentication
// // see: https://medium.com/lightrail/getting-token-authentication-right-in-a-stateless-single-page-application-57d0c6474e3
// type AuthCookies struct {
// 	SecureCookie *securecookie.SecureCookie // signed secure cookie implementation
// 	HTTPOnlyName string                     // http only cookie
// 	JSName       string                     // cookie viewable from js (i.e. to know if the session is no longer valid)
// 	Secure       bool                       // https, or in local test mode http
// 	Duration     time.Duration              // duration for the auth cookies
// }

// // New creates a new AuthCookies
// func New() *AuthCookies {
// 	// load settings from environment
// 	secure := config.GetConfig("PLATFORM_SECURE") == "true"

// 	value := config.GetConfig("PLATFORM_AUTH_COOKIE_HASH")
// 	if value == "" {
// 		panic("cookie hash not set in environment variable PLATFORM_AUTH_COOKIE_HASH")
// 	}
// 	authCookieHash := value

// 	value = config.GetConfig("PLATFORM_SESSION_DURATION")
// 	if value == "" {
// 		panic("session duration not set in environment variable PLATFORM_SESSION_DURATION")
// 	}
// 	sessionDuration, err := time.ParseDuration(value)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Hash keys should be at least 32 bytes long
// 	jwtCookie := securecookie.New([]byte(authCookieHash), nil)

// 	return &AuthCookies{
// 		SecureCookie: jwtCookie,
// 		HTTPOnlyName: "httpauth",
// 		JSName:       "jsauth",
// 		Secure:       secure,
// 		Duration:     sessionDuration,
// 	}
// }

// // SetCookies sets the cookies for authentication
// func (r *AuthCookies) SetCookies(w http.ResponseWriter, numschi string, isAdmin bool) {
// 	log.LogDebugf("SetCookies")

// 	expiration := time.Now().UTC().Add(r.Duration)

// 	value := map[string]string{
// 		"isAdmin":    strconv.FormatBool(isAdmin),
// 		"numschi":    numschi,
// 		"expiration": expiration.Format(time.RFC3339),
// 	}

// 	jsonData, err := json.Marshal(value)
// 	log.LogDebugf("json session is: %+v", jsonData)

// 	secureEncoded, err := r.SecureCookie.Encode(r.HTTPOnlyName, jsonData)

// 	if err == nil {
// 		httpOnlyCookie := &http.Cookie{
// 			Name:     r.HTTPOnlyName,
// 			Value:    secureEncoded,
// 			Path:     "/",
// 			Secure:   r.Secure,
// 			HttpOnly: true,
// 		}
// 		http.SetCookie(w, httpOnlyCookie)

// 		b64Encoded := base64.StdEncoding.EncodeToString(jsonData)

// 		log.LogDebugf("jsAuth cookie value is; %s", b64Encoded)

// 		jsCookie := &http.Cookie{
// 			Name:     r.JSName,
// 			Value:    b64Encoded,
// 			Path:     "/",
// 			Secure:   r.Secure,
// 			HttpOnly: false,
// 			MaxAge:   int(r.Duration.Seconds()),
// 		}
// 		http.SetCookie(w, jsCookie)

// 	}
// }

// // ContextWithCookies puts the cookie values into the context
// func (r *AuthCookies) ContextWithCookies(ctx context.Context, request *http.Request) context.Context {
// 	log.LogDebugf("ContextWithCookies")
// 	numschi, isAdmin, _ := r.GetCookiesValues(request)
// 	ctxWithValues := context.WithValue(ctx, contextKey("numschi"), numschi)
// 	ctxWithValues = context.WithValue(ctxWithValues, contextKey("isAdmin"), isAdmin)

// 	return ctxWithValues
// }

// // GetContextValues returns the values of the auth cookies
// func (r *AuthCookies) GetContextValues(ctx context.Context) (string, bool, string) {
// 	numschi := ctx.Value(contextKey("numschi")).(string)
// 	isAdmin := ctx.Value(contextKey("isAdmin")).(bool)
// 	return numschi, isAdmin, ""
// }

// // GetCookiesValues returns the common value of the auth cookies
// func (r *AuthCookies) GetCookiesValues(request *http.Request) (string, bool, string) {
// 	//utilities.DebugLog(nil,"GetCookiesValue")
// 	httpCookieValues := make(map[string]string)
// 	cookie, err := request.Cookie(r.HTTPOnlyName)
// 	if err != nil {
// 		log.LogDebugf("Http cookie missing (user not logged in), error: %v", err.Error())
// 		return "", false, ""
// 	}
// 	var value []byte
// 	if err = r.SecureCookie.Decode(r.HTTPOnlyName, cookie.Value, &value); err == nil {
// 		//utilities.DebugLog(nil,"read json http cookie: %+v\n", string(value))
// 		json.Unmarshal(value, &httpCookieValues)
// 	}

// 	jsCookieValues := make(map[string]string)
// 	cookie, err = request.Cookie(r.JSName)
// 	if err != nil {
// 		log.LogDebugf("JS cookie missing (user not logged in), error: %v", err.Error())
// 		return "", false, ""
// 	}
// 	log.LogDebugf("found a js cookie")
// 	value, err = base64.StdEncoding.DecodeString(cookie.Value)
// 	if err != nil {
// 		log.LogDebugf("Unexpected JS cookie base64 decode error, error: %v", err.Error())
// 		return "", false, ""
// 	}
// 	//utilities.DebugLog(nil,"read json js cookie: %+v\n", string(value))
// 	json.Unmarshal(value, &jsCookieValues)

// 	// BUG(bjorge): check expiration value itself to see if cookie has expired

// 	if httpCookieValues["isAdmin"] == jsCookieValues["isAdmin"] &&
// 		httpCookieValues["numschi"] == jsCookieValues["numschi"] &&
// 		httpCookieValues["expiration"] == jsCookieValues["expiration"] {
// 		//utilities.DebugLog(nil,"Authorized, cookies match")
// 		isAdmin, err := strconv.ParseBool(httpCookieValues["isAdmin"])
// 		if err != nil {
// 			log.LogErrorf("Unexpected cookie parsing error: %v", err.Error())
// 			return "", false, ""
// 		}
// 		return httpCookieValues["numschi"], isAdmin, httpCookieValues["expiration"]
// 	}

// 	log.LogDebugf("Not authorized, cookies do not match")
// 	return "", false, ""
// }

// // IsContextAuthenticated returns true if authenticated, false otherwise
// func (r *AuthCookies) IsContextAuthenticated(ctx context.Context) bool {
// 	log.LogDebugf("IsContextAuthenticated")
// 	numschi := ctx.Value(contextKey("numschi"))
// 	if numschi == "" {
// 		log.LogDebugf("IsContextAuthenticated: false")
// 		return false
// 	}
// 	return true
// }

// // IsAuthenticated returns true if authenticated, false otherwise
// func (r *AuthCookies) IsAuthenticated(writer http.ResponseWriter, request *http.Request) bool {
// 	numschi, _, _ := r.GetCookiesValues(request)
// 	if numschi == "" {
// 		log.LogDebugf("auth false, no cookie")
// 		r.ClearCookies(writer)
// 		return false
// 	}
// 	log.LogDebugf("auth true, got cookie with numschi %+v", numschi)
// 	return true
// }

// // ClearCookies clears the auth cookies
// func (r *AuthCookies) ClearCookies(w http.ResponseWriter) {
// 	log.LogDebugf("ClearCookies")

// 	httpOnlyCookie := &http.Cookie{
// 		Name:   r.HTTPOnlyName,
// 		MaxAge: -1,
// 	}
// 	http.SetCookie(w, httpOnlyCookie)

// 	jsCookie := &http.Cookie{
// 		Name:   r.JSName,
// 		MaxAge: -1,
// 	}
// 	http.SetCookie(w, jsCookie)
// }
