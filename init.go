package main

import (
	"github.com/bjorge/friendlyreservations/config"
	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/logger"
	"github.com/bjorge/friendlyreservations/persist"
	graphql "github.com/graph-gophers/graphql-go"
)

var log = logger.New()

var corsOriginURI string

var adminSchema *graphql.Schema
var memberSchema *graphql.Schema
var homeSchema *graphql.Schema

func init() {

	corsOriginURI = config.GetConfig("PLATFORM_CORS_ORIGIN_URI")

	frapi.PersistedEmailStore = persist.NewPersistedEmailStore(true)
	frapi.PersistedVersionedEvents = persist.NewPersistedVersionedEvents(true)
	frapi.PersistedPropertyList = persist.NewPersistedPropertyList(true)

	adminSchema = graphql.MustParseSchema(frapi.AdminSchema, &frapi.Resolver{})
	memberSchema = graphql.MustParseSchema(frapi.MemberSchema, &frapi.Resolver{})
	homeSchema = graphql.MustParseSchema(frapi.HomeSchema, &frapi.Resolver{})

}
