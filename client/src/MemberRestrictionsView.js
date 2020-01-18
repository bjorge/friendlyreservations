import React, { Component } from 'react';
import {
    Modal, ModalBody, ModalHeader,
} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import { inject, observer } from "mobx-react";


import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import Restrictions from './Restrictions';

import 'bootstrap/dist/css/bootstrap.css';

const GET_MEMBER_RESTRICTIONS_GQL = gql`
query MemberRestrictions(
    $propertyId: String!) {
        property(id: $propertyId) {
            eventVersion
            settings {
                propertyName
                timezone
                maxOutDays
                minInDays
                minBalance
            }
            restrictions {
                description
                restrictionId
                restriction {
                    __typename
                    ... on BlackoutRestriction {
                        startDate
                        endDate
                    }
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
`;

class MemberRestrictionsView extends Component {
    constructor(props) {
        super(props);
        this.toggleModal = this.toggleModal.bind(this);
        this.state = {
            cachedProperty: null,
        };
    }

    toggleModal() {
        this.props.exitModal();
    }

    static formatDate(dateTime) {
        var date = new Date(dateTime);
        return date.toLocaleString();
    }

    render() {
        var property = this.state.cachedProperty;
        //const eventVersion = this.props.appStateStore.propertyEventVersion ? this.props.appStateStore.propertyEventVersion : 0;
        const apolloClient = this.props.appStateStore.apolloClient;
        const queryKey = this.props.appStateStore.apolloQueryKey;

        var info = { propertyId: this.props.propertyId, userId: this.props.userId };
        return (
            <Modal isOpen={this.props.showModal} toggle={this.toggleModal}>
                <ModalHeader toggle={this.toggleModal}>Reservation Restrictions</ModalHeader>
                <ModalBody>
                    <Query query={GET_MEMBER_RESTRICTIONS_GQL} fetchPolicy='no-cache'
                        client={apolloClient} key={queryKey} 
                        variables={info}
                        onCompleted={(data) => {
                            if (data.property !== undefined) {
                                this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
                                this.setState({ cachedProperty: data.property });
                            }
                        }}
                    >
                        {({ loading, error, data }) => {
                            if (loading) { return (<Spinner />); }
                            if (data) {

                                if (data.property !== undefined) {
                                    property = data.property;
                                }

                                if (property === null) {
                                    return (<div>No data from service, please refresh and try again</div>)
                                }
                            }
                            return (
                                <div>
                                    {error && <ErrorModal error={error} />}
                                    first day checkin allowed: {property.settings.minInDays}<br />
                                    last day checkout allowed: {property.settings.maxOutDays}<br />
                                    minimum balance for new reservation: {property.settings.minBalance}<br />

                                    Other restrictions:
                                    <Restrictions admin={this.props.admin} />
                                </div>
                            )
                        }}
                    </Query>
                </ModalBody>
            </Modal>)
    }
}

export default inject('appStateStore')(observer(MemberRestrictionsView))
