import React, { Component } from 'react';

import {
    Collapse,
    Card,
    CardBody,
    Container,
    Button,
    Label,
    Input,
    Table,

} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

import {
    Redirect
} from "react-router-dom";

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faMinus } from '@fortawesome/free-solid-svg-icons'

import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";


const CANCEL_RESERVATION_GQL = gql`
mutation CancelReservation(
    $propertyId: String!
    $forVersion: Int!
    $reservationId: String!
) {
    cancelReservation(propertyId: $propertyId, forVersion: $forVersion, adminRequest: true, reservationId: $reservationId) {
      eventVersion
    }
  }
`;

const GET_ALL_RESERVATIONS_GQL = gql`
query MemberReservations(
    $propertyId: String!) {
        property(id: $propertyId) {
            propertyId
            eventVersion

            me {
                state
                isAdmin
                isMember
                nickname
                email
                userId
            }

            settings {
                minBalance
              }

            reservations(order: DESCENDING) {
                reservationId
                startDate
                endDate
                member
                nonMemberName
                nonMemberInfo
                reservedFor {
                    nickname
                    userId
                }
                author {
                    nickname
                }
                amount
                canceled
            }

            cancelReservationConstraints(userType: ADMIN) {
                cancelReservationAllowed
            }


        }
    }
`

class AdminReservations extends Component {

    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.toggleIncludeCanceledReservations = this.toggleIncludeCanceledReservations.bind(this);

        this.state = {
            collapse: null,
            showCanceled: false,
            cachedProperty: null,
        };
    }

    toggleIncludeCanceledReservations() {
        this.setState({ showCanceled: !this.state.showCanceled });
    }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }


    static formatDate(dateTime) {
        var date = new Date(dateTime);
        return date.toLocaleString();
    }

    render() {

        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        const { collapse } = this.state;

        var property = this.state.cachedProperty

        if (propertyId === null) {
            return (<Redirect to="/propertyselect" />)
        } else {
            const apolloClient = this.props.appStateStore.apolloClient;
            const queryKey = this.props.appStateStore.apolloQueryKey;

            var info = { propertyId: propertyId };

            return (
                <Query query={GET_ALL_RESERVATIONS_GQL} fetchPolicy='no-cache'
                    client={apolloClient} key={queryKey}
                    variables={info}
                    onCompleted={(data) => {
                        if (data.property !== undefined) {
                            this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
                            this.setState({ cachedProperty: data.property });
                        }
                    }}>
                    {({ loading, error, data }) => {
                        if (loading) { return (<Spinner />); }
                        if (error) { return (<ErrorModal error={error} />); }

                        if (data) {

                            if (data.property !== undefined) {
                                property = data.property
                            }
                            if (property === null) {
                                // cache property is also null
                                return (<div>No data from service, please refresh and try again</div>)
                            }

                            var reservations = property.reservations.reduce(function (map, obj) {
                                map[obj.reservationId] = obj;
                                return map;
                            }, {});

                            var numReservations = Object.keys(reservations).length;

                            console.log("cancel constraints:");
                            console.log(property.cancelReservationConstraints);

                            var allowCancel = {}
                            for (var i=0; i<property.cancelReservationConstraints.cancelReservationAllowed.length; i++) {
                                allowCancel[property.cancelReservationConstraints.cancelReservationAllowed[i]] = true;
                            }
                            console.log(allowCancel);
                        }
                        return (
                            <Container>

                                {/* Reservations List */}
                                {numReservations > 0 && <div><Label check>
                                    <Input type="checkbox" onClick={this.toggleIncludeCanceledReservations} checked={this.state.showCanceled ? 'checked' : ''} onChange={() => { }} />{' '}
                                    Include canceled reservations
                                    </Label></div>}
                                {numReservations > 0 && Object.keys(reservations).map(key => {
                                    var showCanceled = !reservations[key].canceled || this.state.showCanceled
                                    var show = showCanceled
                                    return (
                                        <div key={key}>
                                            {show && <Card key={key}>
                                                <Button className="text-left" onClick={() => this.toggle(key)}>
                                                    {collapse === key ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                                    &nbsp;
                                                    {reservations[key].startDate} {reservations[key].reservedFor.nickname}
                                                </Button>
                                                <Collapse isOpen={collapse === key}>
                                                    <CardBody>
                                                        <Table bordered size="sm">
                                                            <thead>
                                                                <tr>
                                                                    <th>Name</th>
                                                                    <th>Value</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                <tr>
                                                                    <th scope="row">{"Check in date"}</th>
                                                                    <td>{reservations[key].startDate}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Check out date"}</th>
                                                                    <td>{reservations[key].endDate}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Reservation is for"}</th>
                                                                    <td>{reservations[key].reservedFor.nickname}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Reservation made by"}</th>
                                                                    <td>{reservations[key].author.nickname}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Reserved for a non-member?"}</th>
                                                                    <td>{reservations[key].member ? "no" : "yes"}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Non-member name"}</th>
                                                                    <td>{reservations[key].nonMemberName}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Non-member info"}</th>
                                                                    <td>{reservations[key].nonMemberInfo}</td>
                                                                </tr>
                                                                <tr>
                                                                    <th scope="row">{"Canceled"}</th>
                                                                    <td>{reservations[key].canceled ? "yes" : "no"}</td>
                                                                </tr>
                                                            </tbody>
                                                        </Table>

                                                        {allowCancel[key] &&

                                                            <Mutation client={apolloClient} mutation={CANCEL_RESERVATION_GQL} fetchPolicy='no-cache' onCompleted={(data) => {
                                                                if (data.cancelReservation !== undefined) {
                                                                    this.props.appStateStore.setPropertyEventVersion(data.cancelReservation.eventVersion);
                                                                }
                                                            }}>
                                                                {(cancelReservation, { data }) => (
                                                                    <div>
                                                                        <form
                                                                            onSubmit={e => {
                                                                                e.preventDefault();
                                                                                console.log("cancel reservation for id: " + reservations[key].reservationId);
                                                                                cancelReservation({
                                                                                    variables: {
                                                                                        propertyId: info.propertyId,
                                                                                        forVersion: this.props.appStateStore.propertyEventVersion,
                                                                                        reservationId: reservations[key].reservationId
                                                                                    }
                                                                                });
                                                                                this.toggle(key);
                                                                            }}
                                                                        >
                                                                            <button type="submit">Cancel Reservation</button>
                                                                        </form>
                                                                    </div>
                                                                )}
                                                            </Mutation>
                                                        }
                                                    </CardBody>
                                                </Collapse>
                                            </Card>}
                                        </div>
                                    )
                                })}
                            </Container>
                        )
                    }}
                </Query>
            )
        }
    }
}

export default inject('appStateStore')(observer(AdminReservations))
