import React, { Component } from 'react';

import InvitationModal from './InvitationModal';
import { inject, observer } from "mobx-react";
import {
  Card,
  Container,
  CardBody,
  CardText,
} from 'reactstrap';

import {
  Redirect
} from "react-router-dom";

import { Query } from 'react-apollo';
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import gql from 'graphql-tag';

const GET_PROPERTY_HOME_GQL = gql`
query PropertyHome(
  $propertyId: String!) {
  property(id: $propertyId) {
    propertyId
    eventVersion
    contents {
      name
      rendered
    }
    me {
      nickname
      userId
      state
      isAdmin
      isMember
      email
    }
    settings {
      propertyName
    }
  }
}
`;

class PropertyHome extends Component {
  constructor(props) {
    super(props);
    this.onAcceptInvitation = this.onAcceptInvitation.bind(this);
    this.onDeclineInvitation = this.onDeclineInvitation.bind(this);
    this.exitInvitation = this.exitInvitation.bind(this);
    this.setAdminView = this.setAdminView.bind(this);
    this.setMemberView = this.setMemberView.bind(this);

    this.state = {
      cachedProperty: null,
    };
  }

  setAdminView = () => {
    this.props.appStateStore.setPropertyView('ADMIN');
  }

  setMemberView = () => {
    this.props.appStateStore.setPropertyView('MEMBER');
  }

  exitInvitation = () => {
    console.log("exitInvitation");
    this.props.appStateStore.clearAll()
  }

  onAcceptInvitation = () => {
    console.log("onAcceptInvitation");
  }

  onDeclineInvitation = () => {
    console.log("onDeclineInvitation");
    this.props.appStateStore.clearAll()
  }

  static formatRendered(body) {
    return body.replace(/\n/g, "<br />");
  }

  render() {
    const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

    if (propertyId === null) {
      return (<Redirect to="/propertyselect" />)
    } else {

      const propertyView = this.props.appStateStore.propertyView;
      const apolloClient = this.props.appStateStore.apolloClient;
      const queryKey = this.props.appStateStore.apolloQueryKey;

      return (
        <Query client={apolloClient} key={queryKey} query={GET_PROPERTY_HOME_GQL} fetchPolicy='no-cache'
          variables={{
            propertyId: propertyId
          }}
          onCompleted={(data) => {
            if (data.property !== undefined) {
              this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
              this.props.appStateStore.setMe(data.property.me)
              this.setState({ cachedProperty: data.property });
            }
          }}
        >
          {({ loading, error, data }) => {
            if (loading) { return (<Spinner />); }
            if (error) { console.log("Property Home Error"); return (<ErrorModal error={error} />); }
            if (data) {

              var property = this.state.cachedProperty;
              if (data.property !== undefined) {
                property = data.property;
              }

              if (property === undefined) {
                return (<div>No data from service, please refresh and try again</div>)
              }

              // get the property view state
              if (propertyView === undefined) {
                return (<div>View state unknown, please refresh and try again</div>)
              }

              if (property.me.state === 'DECLINED') {
                return (
                  <div>Error: You have previously declined to join.</div>
                );
              }

              if (property.me.state === 'WAITING_ACCEPT') {
                return (
                  <InvitationModal acceptCallback={this.onAcceptInvitation} declineCallback={this.onDeclineInvitation} property={property} showform={true} exitModal={this.exitInvitation} />
                );
              }

              var contentMap = property.contents.reduce(function (map, obj) {
                map[obj.name] = { name: obj.name, rendered: obj.rendered };
                return map;
              }, {});
              return (
                <Container>
                  <Card>
                    <CardBody>
                      {propertyView === 'ADMIN' && <CardText>
                        <span dangerouslySetInnerHTML={{ __html: PropertyHome.formatRendered(contentMap.ADMIN_HOME.rendered) }} />
                      </CardText>}
                      {propertyView === 'MEMBER' && <CardText>
                        <span dangerouslySetInnerHTML={{ __html: PropertyHome.formatRendered(contentMap.MEMBER_HOME.rendered) }} />
                      </CardText>}
                    </CardBody>
                  </Card>
                </Container>
              );

            }
          }}
        </Query>)
    }
  }
  // static homeGqlRequest() {
  //   return gql`
  //   query PropertyHome(
  //     $propertyId: String!) {
  //     property(id: $propertyId) {
  //       propertyId
  //       createDateTime
  //       settings {
  //         propertyName
  //       }
  //       me {
  //         state
  //         isAdmin
  //         isMember
  //         nickname
  //         email
  //         userId    
  //       }
  //     }
  //   }
  //   `;

  // }

}

export default inject('appStateStore')(observer(PropertyHome))

