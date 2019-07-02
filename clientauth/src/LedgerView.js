import React, { Component } from 'react';
import {
    Table,
    Modal, ModalHeader, ModalBody,
} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

import { inject, observer } from "mobx-react";

const GET_LEDGERS_GQL = gql`
query Ledgers(
  $propertyId: String!,
  $userId: String!) {
  property(id: $propertyId) {
    eventVersion

    ledgers(userId: $userId, reverse: true) {
      user {
        nickname
      }
      records {
        event
        eventDateTime
        amount
        balance
      }
    }
}
}
`;

const EventMap = {
    "PAYMENT": "Balance Payment",
    "EXPENSE": "Balance Expense",
    "RESERVATION": "Reservation Purchase",
    "CANCEL_RESERVATION": "Reservation Cancel",
    "MEMBERSHIP_PAYMENT": "Membership Purchase",
    "MEMBERSHIP_OPTOUT": "Membership Opt Out",
    "START": "New Account",
}
class LedgerView extends Component {
    constructor(props) {
        super(props);
        this.renderTable = this.renderTable.bind(this);
        this.renderModalTable = this.renderModalTable.bind(this);
        this.toggleModal = this.toggleModal.bind(this);

        this.state = {
            cachedProperty: null,
        };
    }

    static formatDate(dateTime) {
        var date = new Date(dateTime);
        return date.toLocaleString();
    }

    toggleModal() {
        this.props.exitModal();
    }

    renderModalTable(propertyId, isAdmin, user) {
        return (
            <Modal isOpen={this.props.showModal} toggle={this.toggleModal}>
                <ModalHeader toggle={this.toggleModal}>Ledger for {user.nickname}</ModalHeader>
                <ModalBody>
                    {this.renderTable(propertyId, isAdmin, user)}
                </ModalBody>
            </Modal>
        );
    }

    renderTable(propertyId, isAdmin, user) {

        var info = { propertyId: propertyId, userId: user.userId };
        const apolloClient = this.props.appStateStore.apolloClient;
        const queryKey = this.props.appStateStore.apolloQueryKey;

        return (
            <Query client={apolloClient} key={queryKey} query={GET_LEDGERS_GQL} fetchPolicy='no-cache'
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
                    if (error) { return (<ErrorModal error={error} />); }

                    if (data) {
                        var property = this.state.cachedProperty;
                        if (data.property !== undefined) {
                            property = data.property;
                        }

                        if (property === undefined) {
                            return (<div>No data from service, please refresh and try again</div>)
                        }

                        return (
                            <div>
                                <Table bordered size="sm">
                                    <thead>
                                        <tr>
                                            <th>Date</th>
                                            <th>Event</th>
                                            <th>Amount</th>
                                            <th>Balance</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {property.ledgers && property.ledgers[0].records.map(function (record) {
                                            return (
                                                <tr key={record.eventDateTime}>
                                                    <th scope="row">{LedgerView.formatDate(record.eventDateTime)}</th>
                                                    <td>{EventMap[record.event]}</td>
                                                    <td className="text-right">{record.amount}</td>
                                                    <td className="text-right">{record.balance}</td>
                                                </tr>
                                            )
                                        })}
                                    </tbody>
                                </Table>
                            </div>
                        );
                    }
                }}
            </Query>);
    }



    render() {
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (propertyId === null) {
            return (<div></div>);
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

export default inject('appStateStore')(observer(LedgerView))
