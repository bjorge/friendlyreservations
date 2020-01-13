package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/bjorge/friendlyreservations/cookies"
	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/utilities"
	"github.com/rs/cors"
	graphqlupload "github.com/smithaitufe/go-graphql-upload"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

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
		ctx := context.Background()
		log.LogInfof("Run daily cron")
		err := frapi.DailyCron(ctx)
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
			log.LogInfof("cors handler added for graphql")
			log.LogInfof("cors allowed origin is %v\n", corsOriginURI)

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

	// handle the google test auth
	http.Handle("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.LogDebugf("auth handler")

	}))

	// the login handler
	http.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.LogDebugf("login handler")
		frapi.FrapiCookies.ClearCookies(w)

		noCache(w)
		var htmlIndex = `<html>
		<body>
			<a href="/auth">Login</a>
		</body>
		</html>`

		fmt.Fprintf(w, htmlIndex)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.LogDebugf("Defaulting to port %s", port)
	}

	log.LogDebugf("Listening on port %s", port)
	if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
		log.LogErrorf("error listening on port %v", err)
	}
}

func gqlMiddleware(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		ctxWithValues := context.WithValue(r.Context(), cookies.WriterKey("writer"), w)
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
