package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bjorge/friendlyreservations/cookies"
	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/utilities"
	"github.com/rs/cors"
	graphqlupload "github.com/smithaitufe/go-graphql-upload"
	"google.golang.org/appengine"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

type googleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail string `json:"verified_email"`
	Picture       string `json:"picture"`
	HD            string `json:"hd"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func main() {
	if utilities.SystemEmail == "" {
		panic("DEFAULT_SYSTEM_EMAIL environment variable must be set, example set to noreply@mydomain.com in app.yaml")
	}

	// handle the gql playground pages
	for uri, query := range map[string]string{
		"/adminschema":  "adminquery",
		"/memberschema": "memberquery",
		"/homeschema":   "homequery"} {
		gqlSchemaHTML := mustGetSchemaHTML("gqlschema.html", query)
		http.Handle(uri, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(gqlSchemaHTML)
		}))
	}

	// handle the daily cron job
	http.Handle("/dailycron", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cronHeaders := r.Header["X-Appengine-Cron"]
		if len(cronHeaders) == 0 {
			log.LogDebugf("/dailycron called but not by appengine, so just return")
			return
		}

		log.LogDebugf("X-Appengine-Cron values %v", cronHeaders)

		ctx := appengine.NewContext(r)
		ctx, err := appengine.Namespace(ctx, namespace)
		if err != nil {
			panic(err)
		}
		log.LogInfof("Run daily cron in namespace %v", namespace)
		err = frapi.DailyCron(ctx)
		if err != nil {
			log.LogErrorf("Daily cron error: %+v", err)
		}
	}))

	// handle the graphql requests
	for uri, schema := range map[string]*graphql.Schema{
		"/homequery":   homeSchema,
		"/adminquery":  adminSchema,
		"/memberquery": memberSchema} {

		// the gql handler
		gqlRelayHandler := &relay.Handler{Schema: schema}

		// chain in the upload handler
		gqlHandler := graphqlupload.Handler(gqlRelayHandler)

		// chain in the cors handler
		if corsOriginURI != "" {
			log.LogInfof("cors handler added for graphql for uri: %v with origin: %v", uri, corsOriginURI)

			corsHandler := cors.New(cors.Options{
				AllowedOrigins: []string{corsOriginURI},
				AllowedMethods: []string{
					http.MethodPost,
				},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
			})

			gqlHandler = corsHandler.Handler(gqlHandler)
		}
		// chain in the gql context handler
		http.Handle(uri, gqlMiddleware(gqlHandler))
	}

	// the local test auth handler, override in production
	// note, use openid rather than google signin, more simple and secure according
	// to google: https://developers.google.com/identity/protocols/OpenIDConnect
	// see: https://github.com/plutov/packagemain/tree/master/11-oauth2
	http.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.LogDebugf("login handler (oauth)")

		// uncomment if domain users are restricted (ex. g suite account members only)
		// hdOption := oauth2.SetAuthURLParam("hd", hostedDomain)
		// url := googleOauthConfig.AuthCodeURL(oauthStateString, hdOption)
		url := googleOauthConfig.AuthCodeURL(oauthStateString)
		log.LogDebugf("url for oauth is: %s", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}))

	// handle the oauth callback
	http.Handle("/oauth2callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.LogDebugf("oauth2callback handler")
		originURI := destinationURI
		if corsOriginURI != "" {
			originURI = corsOriginURI
		}

		content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
		if err != nil {
			log.LogErrorf("Get user info error: %+v", err)
			http.Redirect(w, r, originURI+"/", http.StatusTemporaryRedirect)
			return
		}

		log.LogDebugf("oauth2 callback json: %s", content)

		var u googleUser
		err = json.Unmarshal(content, &u)
		if err != nil {
			// Extract the user data
			log.LogDebugf("email is: %+v", u.Email)
			// BUG(bjorge): remove admin from the cookies!
			frapi.FrapiCookies.SetCookies(w, u.Email, true)
		} else {
			log.LogErrorf("Unmarshal error: %+v", err)
			http.Redirect(w, r, originURI+"/", http.StatusTemporaryRedirect)
			return
		}

		// redirect here
		http.Redirect(w, r, originURI+"/", http.StatusTemporaryRedirect)

	}))

	/* uncomment /auth and /login for testing if oauth not setup or having trouble
	// handle the test auth
	http.Handle("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.LogDebugf("auth handler for localhost testing")
		if r.Method != "POST" {
			fmt.Fprintf(w, "hmm... not a post, try again")
			return
		}

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm err: %v", err)
			return
		}

		email := r.FormValue("email")
		if email == "" {
			fmt.Fprintf(w, "hmm... empty email, try again")
			return
		}

		// ok, this is just for testing, so assume a valid email
		// (although any identifier ok for testing...)

		// save auth credentials into cookies
		frapi.FrapiCookies.SetCookies(w, email, false)

		// go back to home
		redirectURL := "/"
		if corsOriginURI != "" {
			redirectURL = corsOriginURI + redirectURL
			log.LogDebugf("redirect to cors origin")
		}

		log.LogDebugf("redirect to: %v", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}))

	// the login handler for testing
	http.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.LogDebugf("login handler for localhost testing")
		frapi.FrapiCookies.ClearCookies(w)

		noCache(w)
		var htmlContent = `
			<html>
				Local test login<br/>
				<form action="/auth" method="post">
					Email:<br/>
					<input type="text" name="email" value=""><br/>
					<input type="submit" value="Submit">
				</form>
			</html>`

		fmt.Fprintf(w, htmlContent)
	}))
	*/

	// for production the spa is built and deployed to the spa directory
	spa := SpaHandler{StaticPath: "spa", IndexPath: "index.html"}
	http.Handle("/", spa)

	appengine.Main() // Starts the server to receive requests

}

func gqlMiddleware(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		ctx, err := appengine.Namespace(ctx, namespace)
		if err != nil {
			panic(err)
		}
		ctxWithValues := context.WithValue(ctx, cookies.WriterKey("writer"), w)
		ctxWithValues = frapi.FrapiCookies.ContextWithCookies(ctxWithValues, r)
		next.ServeHTTP(w, r.WithContext(ctxWithValues))
	}

	return http.HandlerFunc(fn)
}

func mustGetSchemaHTML(fileName string, gqlPath string) []byte {
	t := template.New(fileName)
	t, err := t.ParseFiles(fileName)
	if err != nil {
		panic(err)
	}

	// refer to the gql handler above
	var buffer bytes.Buffer
	err = t.Execute(&buffer, struct{ Path string }{Path: gqlPath})
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SpaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type SpaHandler struct {
	StaticPath string
	IndexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.StaticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.StaticPath, h.IndexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.StaticPath)).ServeHTTP(w, r)
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}
