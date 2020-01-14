import { observable, action, computed, decorate } from "mobx";
//import gql from 'graphql-tag';

import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';

import { IntrospectionFragmentMatcher } from 'apollo-cache-inmemory';

// run getschema.js to generate the fragmentTypes.json file
import introspectionQueryResultData from './fragmentTypes.json';

// see https://blog.uncommon.is/a-simple-introduction-to-state-management-with-mobx-in-react-native-ed749aa2b5d7

// const PROPERTY_RESULTS_GQL_STRING = `
//   propertyId
//   eventVersion
//   settings {
//     propertyName
//   }
//   me {
//     state
//     isAdmin
//     isMember
//     nickname
//     email
//     userId    
//   }
//   users {
//     nickname
//     isAdmin
//     isMember
//     isSystem
//     userId
//     email
//   }

//   `

// the common gql for properties used by this store
// const GET_PROPERTY_GQL = gql`
// query PropertyHome(
//   $propertyId: String!) {
//   property(id: $propertyId) {
// ${PROPERTY_RESULTS_GQL_STRING}
//   }
// }
// `;

// get rid of warnings due to using unions in gql, see:
// https://www.apollographql.com/docs/react/advanced/fragments.html
const fragmentMatcher = new IntrospectionFragmentMatcher({
  introspectionQueryResultData
});

const cache = new InMemoryCache({ fragmentMatcher });

var baseGQLUrl = "";
if (process.env.NODE_ENV === "development") {
  baseGQLUrl = "http://localhost:8080";
}

const adminClient = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: baseGQLUrl + "/adminquery", credentials: 'same-origin', }),
  cache: cache,
  // defaultOptions: defaultOptions,
  connectToDevTools: true,
});

const memberClient = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: baseGQLUrl + "/memberquery", credentials: 'same-origin', }),
  cache: cache,
  // defaultOptions: defaultOptions,
  connectToDevTools: true,
});

const homeClient = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: baseGQLUrl + "/homequery", credentials: 'same-origin', }),
  cache: cache,
  // defaultOptions: defaultOptions,
  connectToDevTools: true,
});

export default class AppStateStore {
  // authenticated (logged in)
  authenticated = null;
  // propertyId is set by the property selector
  propertyId = null;
  // property is set by the property home page
  property = null;
  me = null;
  // view is null, ADMIN or MEMBER
  propertyView = null;
  // event version
  propertyEventVersion = null;

  // the apollo client
  apolloAdminClient = adminClient;
  apolloMemberClient = memberClient;
  apolloHomeClient = homeClient;

  setAuthenticated(isLoggedIn) {
    this.authenticated = isLoggedIn;
  }

  setPropertyView(view) {
    // set to 'MEMBER' or 'ADMIN'
    this.propertyView = view;
  }

  setPropertyEventVersion(version) {
    this.propertyEventVersion = version;
  }

  setProperty(property) {
    this.property = property;
  }

  setMe(me) {
    this.me = me;
  }

  setPropertyId(newId) {
    this.propertyId = newId;
  }

  clearAll() {
    this.propertyId = null;
    this.propertyView = null;
    this.property = null;
    this.propertyEventVersion = null;
  }

  get headerPageState() {
    if (this.propertyId == null) {
      return 'LIST_PROPERTIES'
    } else {
      return 'MANAGE_PROPERTY'
    }
  }

  // static propertyGql() {
  //   return GET_PROPERTY_GQL;
  // }

  // static propertyGqlResultsString() {
  //   return PROPERTY_RESULTS_GQL_STRING;
  // }

  get apolloClient() {
    if (this.propertyId == null) {
      return this.apolloHomeClient;
    } else if (this.propertyView === 'ADMIN') {
      return this.apolloAdminClient;
    } else {
      return this.apolloMemberClient;
    }
  }

  get apolloQueryKey() {
    const queryKey = this.propertyId + ":" + this.propertyEventVersion + ":" + this.me.userId;
    return queryKey;
  }

}

decorate(AppStateStore, {
  // the authenticated state (true or false)
  authenticated: observable,
  
  // the currently used property (can be set to null when at property selection page)
  propertyId: observable,
  // the current property
  property: observable,
  me: observable,
  // the current view of the property (member or admin)
  propertyView: observable,
  propertyEventVersion: observable,

  setAuthenticated: action,

  setPropertyId: action,
  setProperty: action,
  setMe: action,
  setPropertyView: action,
  setPropertyEventVersion: action,
  headerPageState: computed,
  apolloClient: computed,
  apolloQueryKey: computed,

  clearAll: action,
});
