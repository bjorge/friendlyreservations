package main

import (
	"fmt"
	"os"

	"github.com/bjorge/friendlyreservations/config"
	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/logger"
	"github.com/bjorge/friendlyreservations/persist"
	graphql "github.com/graph-gophers/graphql-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var log = logger.New()

var namespace string

var corsOriginURI string
var destinationURI string

var adminSchema *graphql.Schema
var memberSchema *graphql.Schema
var homeSchema *graphql.Schema

// variables from environment
// var hostedDomain string // example "mydomain.com"
var secure bool         // if true, use https, otherwise http (PLATFORM_SECURE)
var clientID string     // the oauth client id
var clientSecret string // the oauth client secret

// package variables
var googleOauthConfig *oauth2.Config

// BUG(bjorge): put into config 
var oauthStateString = "7ac66c0f148de9519b8bd264312c4d64"

func init() {
	corsOriginURI = config.GetConfig("PLATFORM_CORS_ORIGIN_URI")

	frapi.PersistedEmailStore = persist.NewPersistedEmailStore(false)
	frapi.PersistedVersionedEvents = persist.NewPersistedVersionedEvents(false)
	frapi.PersistedPropertyList = persist.NewPersistedPropertyList(false)

	adminSchema = graphql.MustParseSchema(frapi.AdminSchema, &frapi.Resolver{})
	memberSchema = graphql.MustParseSchema(frapi.MemberSchema, &frapi.Resolver{})
	homeSchema = graphql.MustParseSchema(frapi.HomeSchema, &frapi.Resolver{})

	namespace = os.Getenv("PLATFORM_NAMESPACE")
	if namespace == "" {
		panic(fmt.Errorf("must define PLATFORM_NAMESPACE in app.yaml"))
	}

	destinationURI = config.GetConfig("PLATFORM_DESTINATION_URI")
	if destinationURI == "" {
		panic(fmt.Errorf("PLATFORM_DESTINATION_URI is not set"))
	}

	// load settings from environment
	secure = os.Getenv("PLATFORM_SECURE") == "true"
	clientID = os.Getenv("PLATFORM_CLIENT_ID")
	clientSecret = os.Getenv("PLATFORM_CLIENT_SECRET")
	//hostedDomain = os.Getenv("PLATFORM_HOSTED_DOMAIN")

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  destinationURI + "/oauth2callback",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
