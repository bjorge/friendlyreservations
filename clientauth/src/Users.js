import React, { Component } from 'react';

import { Query } from 'react-apollo';
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import gql from 'graphql-tag';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faMinus } from '@fortawesome/free-solid-svg-icons'

import { inject, observer } from "mobx-react";

import {
    Collapse,
    Card,
    CardBody,
    Container,
    Button,
} from 'reactstrap';

import {
    Redirect
} from "react-router-dom";
import User from './User';

const GET_USERS_GQL = gql`
query Users(
  $propertyId: String!) {
  property(id: $propertyId) {
    eventVersion
    users {
        userId
        email
        isAdmin
        isMember
        isSystem
        state
        nickname
    }

    membershipStatusConstraints {
        user {
          nickname
        }
    }

    updateUserConstraints {
        nicknameMin
        nicknameMax
        emailMin
        emailMax
        invalidNicknames
        invalidEmails
    }

    updateBalanceConstraints {
        amountMin
        amountMax
        descriptionMin
        descriptionMax  
    }

    reservationConstraints: newReservationConstraints(userType: ADMIN) {
        newReservationAllowed
        nonMemberNameMin
        nonMemberNameMax
        nonMemberInfoMin
        nonMemberInfoMax
        checkinDisabled {
          before
          after
          from
          to
        }
        checkoutDisabled {
          before
          after
          from
          to
        }
    }
  }
}
`;

class Users extends Component {
    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.toggleNewUserForm = this.toggleNewUserForm.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);

        this.state = {
            collapse: null,
            showNewUserForm: false,
            cachedProperty: null,
        };

    }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }

    toggleNewUserForm() {
        this.setState({ showNewUserForm: !this.state.showNewUserForm });
    }

    turnOffModals = () => {
        this.setState({ showNewUserForm: false });
    }

    render() {
        const { collapse } = this.state;
        //const eventVersion = this.props.appStateStore.propertyEventVersion ? this.props.appStateStore.propertyEventVersion : 0;
        const me = this;
        //const apolloClient = this.props.appStateStore.apolloClient;
        //const queryKey = this.props.appStateStore.apolloQueryKey;

        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (propertyId === null) {
            return (<Redirect to="/propertyselect" />)
        } else {

            //const propertyView = this.props.appStateStore.propertyView;
            const apolloClient = this.props.appStateStore.apolloClient;
            const queryKey = this.props.appStateStore.apolloQueryKey;
            return (
                <Query client={apolloClient} key={queryKey} query={GET_USERS_GQL} fetchPolicy='no-cache'
                    variables={{
                        propertyId: this.props.appStateStore.propertyId
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
                        if (error) { return (<ErrorModal error={error} />); }
                        if (data) {

                            var property = this.state.cachedProperty;
                            if (data.property !== undefined) {
                                property = data.property;
                            }

                            if (property === undefined) {
                                return (<div>No data from service, please refresh and try again</div>)
                            }
                        }
                        return (
                            <Container>
                                {property.users && property.users.map(function (user) {
                                    return (
                                        <div key={user.userId}>
                                            {!user.isSystem && <Card key={user.userId}>
                                                <Button className="text-left" onClick={() => me.toggle(user.userId)}>
                                                    {collapse === user.userId ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                                    &nbsp;
                                            {user.nickname}
                                                </Button>
                                                <Collapse isOpen={collapse === user.userId}>
                                                    <CardBody>
                                                        <User users={property.users} userConstraints={property.updateUserConstraints}  balanceConstraints={property.updateBalanceConstraints} reservationConstraints={property.reservationConstraints} viewinfo={user} />
                                                    </CardBody>
                                                </Collapse>
                                            </Card>}
                                        </div>
                                    )
                                })}

                                <hr />
                                <User users={property.users} userConstraints={property.updateUserConstraints} showform={this.state.showNewUserForm} exitModal={this.turnOffModals} />
                                <div className="text-center">
                                    <Button color="primary" onClick={() => this.toggleNewUserForm()}>New User</Button>
                                </div>
                            </Container>);
                    }}
                </Query>
            )
        }


    }
}

export default inject('appStateStore')(observer(Users))

