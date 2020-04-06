## friendlyreservations

Friendly Reservations is a reservation system framework and implementation using GraphQL, React and Go. 

The underlying architecture uses an Event Sourcing design pattern where only GraphQL mutations are persisted. GraphQL queries are generated in memory from cached replays of the events.

The project is a work in progress but is currenty in live use.

## try it out (requires gmail login)

- [Live demo site](https://trial.friendlyreservations.org/)
- Demo site GraphQL playgrounds (first login to demo site)
  - [home](https://trial.friendlyreservations.org/homeschema) (ex. `{ properties { propertyId settings { propertyName } } }`)
  - [member](https://trial.friendlyreservations.org/memberschema)
  - [admin](https://trial.friendlyreservations.org/adminschema)
- [Recorded demo](https://youtu.be/5C7mCkCO6qk)

## install and run it

    (install)
    mkdir $GOPATH/src/github.com/bjorge
    cd $GOPATH/src/github.com/bjorge; git clone https://github.com/bjorge/friendlyreservations.git
    cd friendlyreservations
    go get -u ./...
    cd client; npm install; npm run build; cd ..
    
    (run)
    go build && ./friendlyreservations

to install and run in appengine see [gae instructions](../master/gae_main/doc.go)

## architectural features

- event-sourced non-mutable persisted store of GraphQL mutations for each property
- emails (PII) stored in separate mutable persisted store to support EU GDPR
- replay (rollup) of events on the fly in memory for GraphQL queries
- events and replays are cached in memcache
- duplicate request suppression
- secure authentication cookies
- easy cors setup, for example for react client development (ex. npm start)
- platform (ex gae, aws) abstracted behind interface calls in platform package
- an appengine implentation is included, which uses google oauth for authentication

## implementation features

### member

- property selection
- accept/reject property membership
- reservation management (create, delete)
- purchase membership (optional)
- non-member (friend) reservation (optional)
- view ledger (history of reservations, payments, memberships, etc)
- view past notifications (history of email notifications)
- login/logout

### admin

- property management (create, delete)
- property settings management (reservation rates, property timezone, etc.)
- user management (add, modify, delete)
- member balance management (payment, expense)
- member reservation override (create, delete)
- restriction management (memberships, blackouts, etc.)
- membership override (payment, optout)
- home screen customization for members and admins
- export/import property database

## use cases

Friendly Reservations supports managing a property for a set of members, for example a family vacation home, a scout cabin, etc. The system allows members to (optionally) make a reservation on behalf of a non-member (ex. friend) at a separate rate. The system supports restrictions such as membership periods, blackout dates, etc. The system provides a configurable credit limit in order to promote use of the property while at the same time limiting monopolization of reservation dates by single members. Email reminders are sent for low balances ("nag" notifications). Administrative interaction with the system is minimal, primarily member balance updating for payments or expenses. However, administrators can also perform actions on behalf of members, manage settings and restrictions, and can also be members themselves.

## code organization

### GraphQL schema

There are 3 schemas in use:

- property selection (home) schema
- member schema
- administrative schema

### Go code

Each application feature, example reservations, users, restrictions, settings, etc., is generally divided into 4 go files:

- feature_mutation.go - handles the graphql mutation request, persisting the mutation
  - [example user_mutation.go](../master/frapi/user_mutation.go)
- feature_query.go - handles the graphql query request, first replaying the feature mutations to get the latest feature state
  - [example user_query.go](../master/frapi/user_query.go)
- feature_rollup.go - called by query code, creates by replaying mutations or retrieves from memcache a map of versions of each feature record
  - [example user_rollup.go](../master/frapi/user_rollup.go)
- feature_constraints.go - called by the client and the mutation code to validate the mutation
  - [example user_constraints.go](../master/frapi/user_constraints.go)
- feature_test.go - tests feature code
  - [example user_test.go](../master/frapi/user_test.go)

### React code

Generally each feature is a js file with the feature name.
  - [example Settings.js](../master/client/src/Settings.js)

## session

The core layer manages secure session cookies which follows [this pattern](https://medium.com/lightrail/getting-token-authentication-right-in-a-stateless-single-page-application-57d0c6474e3).

## authentication

Authentication is managed by the implementation layer. In the case of gae, authentication is handled using oauth and gmail accounts.

## other

[Interesting event sourcing talk](https://youtu.be/rUDN40rdly8)

[Interesting GraphQL/Go article](https://medium.com/safetycultureengineering/why-we-moved-our-graphql-server-from-node-js-to-golang-645b00571535)








