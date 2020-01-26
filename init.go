package main

import (
	"github.com/bjorge/friendlyreservations/config"
	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/local_platform"
	"github.com/bjorge/friendlyreservations/logger"
	graphql "github.com/graph-gophers/graphql-go"
)

var log = logger.New()

var corsOriginURI string

var adminSchema *graphql.Schema
var memberSchema *graphql.Schema
var homeSchema *graphql.Schema

var redirectURL string
var redirectLabel string

func init() {

	corsOriginURI = config.GetConfig("PLATFORM_CORS_ORIGIN_URI")

	redirectURL = config.GetConfig("REDIRECT_URL")
	redirectLabel = config.GetConfig("REDIRECT_LABEL")

	frapi.PersistedEmailStore = localplatform.NewPersistedEmailStore()
	frapi.PersistedVersionedEvents = localplatform.NewPersistedVersionedEvents()
	frapi.PersistedPropertyList = localplatform.NewPersistedPropertyList()
	frapi.EmailSender = localplatform.NewEmailSender()

	adminSchema = graphql.MustParseSchema(frapi.AdminSchema, &frapi.Resolver{})
	memberSchema = graphql.MustParseSchema(frapi.MemberSchema, &frapi.Resolver{})
	homeSchema = graphql.MustParseSchema(frapi.HomeSchema, &frapi.Resolver{})

}
