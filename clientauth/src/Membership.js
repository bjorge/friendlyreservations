import React, { Component } from 'react';

import 'bootstrap/dist/css/bootstrap.css';

import { inject, observer } from "mobx-react";
import {
  Table,
  Button,
  Form,
  Modal, ModalHeader, ModalBody,
} from 'reactstrap';

import {
  Redirect
} from "react-router-dom";

import { Query, Mutation } from 'react-apollo';
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import gql from 'graphql-tag';

const GET_MEMBERSHIP_STATUS_GQL = gql`
query PropertyHome(
  $propertyId: String!,
    $userId: String!) {
      property(id: $propertyId) {
    propertyId
    eventVersion
    me {
      userId
    }

    membershipStatusConstraints(userId: $userId) {
      user {
        nickname
      }
      memberships {
        status
        optOutAllowed
        purchaseAllowed
        reservationCount
        info {
          restrictionId
          restriction {
            ... on MembershipRestriction {
              inDate
              outDate
              gracePeriodOutDate
              amount
            }
          }
        }
      }
    }
  }
}
`;

const UPDATE_MEMBERSHIP_STATUS_GQL = gql`
mutation MemberUpdateMembership(
  $propertyId: String!,
  $userId: String!,
    $input: UpdateMembershipInput!) {
  updateMembershipStatus(propertyId: $propertyId, 
    input: $input) {
      propertyId
      eventVersion
      me {
        userId
      }
  
      membershipStatusConstraints(userId: $userId) {
        user {
          nickname
        }
        memberships {
          status
          optOutAllowed
          purchaseAllowed
          info {
            restrictionId
            restriction {
              ... on MembershipRestriction {
                inDate
                outDate
                gracePeriodOutDate
                amount
              }
            }
          }
        }
      }
    }
}
`;

class Membership extends Component {
  constructor(props) {
    super(props);
    this.renderAction = this.renderAction.bind(this);
    this.renderTable = this.renderTable.bind(this);
    this.renderModalTable = this.renderModalTable.bind(this);
    this.toggleModal = this.toggleModal.bind(this);

    this.state = {
      cachedProperty: null,
    };
  }

  toggleModal() {
    this.props.exitModal();
  }

  renderAction(record, purchase, isAdmin, user) {
    return (

      <Mutation client={this.props.appStateStore.apolloClient} mutation={UPDATE_MEMBERSHIP_STATUS_GQL} fetchPolicy='no-cache'
        onCompleted={(data) => {
          if (data.updateMembershipStatus !== undefined) {
            this.props.appStateStore.setPropertyEventVersion(data.updateMembershipStatus.eventVersion)
            this.setState({ cachedProperty: data.updateMembershipStatus });
          }
        }}>
        {(newUserSubmit, { loading, error }) => {
          if (loading) return (<Spinner />);
          return (
            <div>
              {error && <ErrorModal error={error} />}
              <Form onSubmit={event => {
                event.preventDefault();

                var comment = null;
                if (isAdmin) {
                  comment = "update by admin";
                }

                // ok, we can submit! let's setup a cool gql mutation
                var info = {
                  propertyId: this.props.appStateStore.propertyId,
                  userId: user.userId,
                  input: {
                    forVersion: this.state.cachedProperty.eventVersion,
                    updateForUserId: user.userId,
                    restrictionId: record.info.restrictionId,
                    purchase: purchase,
                    adminUpdate: isAdmin,
                    comment: comment,
                  }
                }
                newUserSubmit({
                  variables: info
                });

              }}
              >
                <div className="text-center">
                  <Button color="primary" type="submit">{purchase ? "Buy" : "OptOut"}</Button>
                </div>
              </Form>

            </div>);
        }}
      </Mutation>
    );
  }

  renderModalTable(propertyId, isAdmin, user) {
    return (
      <Modal isOpen={this.props.showModal} toggle={this.toggleModal}>
        <ModalHeader toggle={this.toggleModal}>Memberships for {user.nickname}</ModalHeader>
        <ModalBody>
          {this.renderTable(propertyId, isAdmin, user)}
        </ModalBody>
      </Modal>
    );
  }

  renderTable(propertyId, isAdmin, user) {
    const buttonThis = this;

    return (
      <Query client={this.props.appStateStore.apolloClient} key={this.props.appStateStore.apolloQueryKey} query={GET_MEMBERSHIP_STATUS_GQL} fetchPolicy='no-cache'
        variables={{
          propertyId: propertyId,
          userId: user.userId
        }}
        onCompleted={(data) => {
          if (data.property !== undefined) {
            this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
            this.setState({ cachedProperty: data.property });
          }
        }}
      >
        {({ loading, error, data }) => {
          if (loading) { return (<Spinner />); }
          if (error) { console.log("Membership Access Error"); return (<ErrorModal error={error} />); }
          if (data) {

            var property = this.state.cachedProperty;
            if (data.property !== undefined) {
              property = data.property;
            }

            if (property === undefined) {
              return (<div>No data from service, please refresh and try again</div>)
            }

            var numStatusRecords = Object.keys(property.membershipStatusConstraints).length;

            if (numStatusRecords === 0) {
              return (<div />)
            }

            return (
              <Table bordered size="sm">
                <thead>
                  <tr>
                    <th>Membership</th>
                    <th>Status</th>
                    <th>Buy</th>
                    {isAdmin && <th>Opt Out</th>}
                    {isAdmin && <th>Res.</th>}
                  </tr>
                </thead>
                <tbody>
                  {property.membershipStatusConstraints[0].memberships && property.membershipStatusConstraints[0].memberships.map(function (record) {
                    return (
                      <tr key={record.info.restrictionId}>
                        <th scope="row">{record.info.restriction.inDate}</th>
                        <td>{record.status}</td>
                        {record.purchaseAllowed && <td>{buttonThis.renderAction(record, true, isAdmin, user)}</td>}
                        {!record.purchaseAllowed && <td>{' '}</td>}
                        {isAdmin && record.optOutAllowed && <td>{buttonThis.renderAction(record, false, isAdmin, user)}</td>}
                        {isAdmin && !record.optOutAllowed && <td>{' '}</td>}
                        {isAdmin && <td>{record.reservationCount}</td>}

                      </tr>
                    )
                  })}
                </tbody>
              </Table>
            );
          }
        }}
      </Query>);
  }



  render() {
    const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;


    if (propertyId === null) {
      return (<Redirect to="/propertyselect" />);
    } else {
      var user = this.props.appStateStore.me;
      if (this.props.isAdmin) {
        user = this.props.user;
      }
      return (
        <div>
          {!this.props.isModal && this.renderTable(propertyId, this.props.isAdmin, user)}
          {this.props.isModal && this.renderModalTable(propertyId, this.props.isAdmin, user)}
        </div>);
    }
  }
}

export default inject('appStateStore')(observer(Membership))

