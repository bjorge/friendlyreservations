import React, { Component } from 'react';

import {
    Modal, ModalBody, ModalHeader,
    Collapse,
    Card,
    CardBody,
    Container,
    Button,
    Label,
    Input,

} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faMinus } from '@fortawesome/free-solid-svg-icons'

import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";


const CANCEL_RESERVATION_GQL = gql`
mutation CancelReservation(
    $propertyId: String!
    $reservationId: String!
) {
    cancelReservation(propertyId: $propertyId, adminRequest: false, reservationId: $reservationId) {
      eventVersion
    }
  }
`;

const GET_MEMBER_RESERVATIONS_GQL = gql`
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

            reservations {
                reservationId
                startDate
                endDate
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
        }
    }
`

class MemberReservationsView extends Component {

    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.toggleModal = this.toggleModal.bind(this);
        this.toggleShowMyReservations = this.toggleShowMyReservations.bind(this);
        this.toggleIncludeCanceledReservations = this.toggleIncludeCanceledReservations.bind(this);

        this.state = {
            collapse: null,
            showMyReservations: true,
            showCanceled: false,
            cachedProperty: null,
        };
    }

    toggleShowMyReservations() {
        this.setState({ showMyReservations: !this.state.showMyReservations });
    }

    toggleIncludeCanceledReservations() {
        this.setState({ showCanceled: !this.state.showCanceled });
    }


    toggleModal() {
        this.setState({ collapse: null });
        this.props.exitModal();
    }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }

    static formatDate(dateTime) {
        var date = new Date(dateTime);
        return date.toLocaleString();
    }

    render() {
        const apolloClient = this.props.appStateStore.apolloClient;
        const queryKey = this.props.appStateStore.apolloQueryKey;

        const { collapse } = this.state;

        var info = { propertyId: this.props.propertyId, userId: this.props.userId };

        //const eventVersion = this.props.appStateStore.propertyEventVersion ? this.props.appStateStore.propertyEventVersion : 0;
        var property = this.state.cachedProperty

        return (
            <Modal isOpen={this.props.showModal} toggle={this.toggleModal}>
                <ModalHeader toggle={this.toggleModal}>Reservations</ModalHeader>
                <ModalBody>
                    <Query query={GET_MEMBER_RESERVATIONS_GQL} fetchPolicy='no-cache'
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

                                var me = property.me
                            }
                            return (
                                <Container>
                                    <Label check>
                                        <Input type="checkbox" onClick={this.toggleShowMyReservations} checked={this.state.showMyReservations ? 'checked' : ''} onChange={() => { }} />{' '}
                                        Show only my reservations
                                    </Label>
                                    <br />
                                    <Label check>
                                        <Input type="checkbox" onClick={this.toggleIncludeCanceledReservations} checked={this.state.showCanceled ? 'checked' : ''} onChange={() => { }} />{' '}
                                        Include canceled reservations
                                    </Label>
                                    {error && <ErrorModal error={error} />}
                                    {reservations && Object.keys(reservations).map(key => {
                                        var showCanceled = !reservations[key].canceled || this.state.showCanceled
                                        var showMine = me.userId === reservations[key].reservedFor.userId || !this.state.showMyReservations
                                        var show = showCanceled && showMine
                                        return (
                                            <div key={key}>
                                                {show && <Card key={key}>
                                                    <Button className="text-left" onClick={() => this.toggle(key)}>
                                                        {collapse === key ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                                        &nbsp;
                                                    checkin {reservations[key].startDate} for {reservations[key].reservedFor.nickname}
                                                    </Button>
                                                    <Collapse isOpen={collapse === key}>
                                                        <CardBody>
                                                            checkin: {reservations[key].startDate}<br />
                                                            checkout: {reservations[key].endDate}<br />
                                                            reserved for: {reservations[key].reservedFor.nickname}<br />
                                                            reserved by: {reservations[key].author.nickname}<br />
                                                            reserved for non member?: {reservations[key].member ? "false" : "true"}<br />
                                                            canceled: {reservations[key].canceled ? "true" : "false"}<br />
                                                            {reservations[key].canceled === false && me.userId === reservations[key].reservedFor.userId &&

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
                </ModalBody>
            </Modal>)
    }
}

export default inject('appStateStore')(observer(MemberReservationsView))
