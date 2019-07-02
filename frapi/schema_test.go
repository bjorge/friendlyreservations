package frapi

import (
	"context"
	"testing"

	"github.com/bjorge/friendlyreservations/utilities"
	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/graph-gophers/graphql-go"
)

// test that all the schemas compile
func TestSchemaParsing(t *testing.T) {
	graphql.MustParseSchema(MemberSchema, &Resolver{})
	graphql.MustParseSchema(AdminSchema, &Resolver{})
	graphql.MustParseSchema(HomeSchema, &Resolver{})
}

type helloWorldResolver1 struct{}

func (r *helloWorldResolver1) Hello() string {
	return "Hello world!"
}

type helloWorldResolver2 struct{}

func (r *helloWorldResolver2) Hello(ctx context.Context) (string, error) {
	return "Hello world!", nil
}

// TestHelloWorld just checks that the basic graphql engine is working
func TestHelloWorld(t *testing.T) {
	utilities.SetTestingNow()

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: graphql.MustParseSchema(`
				schema {
					query: Query
				}

				type Query {
					hello: String!
				}
			`, &helloWorldResolver1{}),
			Query: `
				{
					hello
				}
			`,
			ExpectedResult: `
				{
					"hello": "Hello world!"
				}
			`,
		},

		{
			Schema: graphql.MustParseSchema(`
				schema {
					query: Query
				}

				type Query {
					hello: String!
				}
			`, &helloWorldResolver2{}),
			Query: `
				{
					hello
				}
			`,
			ExpectedResult: `
				{
					"hello": "Hello world!"
				}
			`,
		},
	})
}
