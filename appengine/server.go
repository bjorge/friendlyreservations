package main

import (
	"encoding/json"
	"net/http"

	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/utilities"

	"github.com/graph-gophers/graphql-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var adminSchema *graphql.Schema
var memberSchema *graphql.Schema
var homeSchema *graphql.Schema

func init() {
	adminSchema = graphql.MustParseSchema(frapi.AdminSchema, &frapi.Resolver{})
	memberSchema = graphql.MustParseSchema(frapi.MemberSchema, &frapi.Resolver{})
	homeSchema = graphql.MustParseSchema(frapi.HomeSchema, &frapi.Resolver{})
}

func main() {
	if utilities.SystemEmail == "" {
		panic("DEFAULT_SYSTEM_EMAIL environment variable must be set, example set to noreply@mydomain.com in app.yaml")
	}
	http.Handle("/adminschema", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(adminPage)
	}))

	http.Handle("/memberschema", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(memberPage)
	}))

	http.Handle("/homeschema", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(homePage)
	}))

	http.Handle("/dailycron", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		log.Infof(ctx, "Run daily cron")
		err := frapi.DailyCron(ctx)
		if err != nil {
			log.Errorf(ctx, "Daily cron error: %+v", err)
		}
	}))

	// from example, does not handle OPTIONS request
	// http.Handle("/query", &relay.Handler{Schema: schema})

	// handle the OPTIONS request because app.yaml does not support CORS for non-static content
	http.Handle("/adminquery", NewGraphqlHandler(adminSchema))
	http.Handle("/memberquery", NewGraphqlHandler(memberSchema))
	http.Handle("/homequery", NewGraphqlHandler(homeSchema))

	//log.Fatal(http.ListenAndServe(":8080", nil))

	appengine.Main() // Starts the server to receive requests
}

// Handler is the http handler struct
type Handler struct {
	Schema *graphql.Schema
}

// NewGraphqlHandler creates an new GraphQL API HTTP handler with the specified graphql schema
func NewGraphqlHandler(s *graphql.Schema) *Handler {
	return &Handler{Schema: s}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// dump, err := httputil.DumpRequest(r, true)
	// if err != nil {
	// 	http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	// 	return
	// }
	// fmt.Printf("%s", dump)

	if utilities.AllowCrossDomainRequests {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, HEAD")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			return
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding")
	}

	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	ctx := appengine.NewContext(r)
	log.Debugf(ctx, "Server: start request %+v", r.RequestURI)

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := h.Schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	log.Debugf(ctx, "Server: end request")
	w.Write(responseJSON)
}

// BUG(bjorge): admin page and member page should be the same except for the uri, so template it so
var adminPage = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.css" rel="stylesheet" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/es6-promise/4.1.1/es6-promise.auto.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/2.0.3/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/16.2.0/umd/react.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react-dom/16.2.0/umd/react-dom.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/adminquery", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)

var memberPage = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.css" rel="stylesheet" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/es6-promise/4.1.1/es6-promise.auto.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/2.0.3/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/16.2.0/umd/react.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react-dom/16.2.0/umd/react-dom.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/memberquery", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)

var homePage = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.css" rel="stylesheet" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/es6-promise/4.1.1/es6-promise.auto.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/2.0.3/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/16.2.0/umd/react.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react-dom/16.2.0/umd/react-dom.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/homequery", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
