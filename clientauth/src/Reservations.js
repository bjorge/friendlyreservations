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
import CreateReservation from './CreateReservation';
import Membership from './Membership';

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
    cancelReservation(propertyId: $propertyId, forVersion: $forVersion, adminRequest: false, reservationId: $reservationId) {
      eventVersion
    }
  }
`;

const GET_MEMBER_RESERVATIONS_GQL = gql`
query MemberReservations(
    $propertyId: String!,
    $userId: String!) {
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

            ledgers(userId: $userId, reverse: false, last: 1) {
                records {
                  balance
                }
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

            cancelReservationConstraints(userType: MEMBER, userId: $userId) {
                cancelReservationAllowed
            }

            membershipStatusConstraints(userId: $userId) {
                user {
                  nickname
                }
            }

            memberConstraints: newReservationConstraints(userId: $userId, userType: MEMBER) {
                newReservationAllowed
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

            nonMemberConstraints: newReservationConstraints(userId: $userId, userType: NONMEMBER) {
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
`

class Reservations extends Component {

    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.toggleModal = this.toggleModal.bind(this);
        this.toggleShowMyReservations = this.toggleShowMyReservations.bind(this);
        this.toggleIncludeCanceledReservations = this.toggleIncludeCanceledReservations.bind(this);
        this.displayMemberReservationModal = this.displayMemberReservationModal.bind(this);
        this.displayNonMemberReservationModal = this.displayNonMemberReservationModal.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);

        this.state = {
            collapse: null,
            showMyReservations: false,
            showCanceled: false,
            showMemberReservationModal: false,
            showNonMemberReservationModal: false,
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

    displayMemberReservationModal() {
        this.setState({
            showMemberReservationModal: true
        });
    }

    displayNonMemberReservationModal() {
        this.setState({
            showNonMemberReservationModal: true
        });
    }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }

    turnOffModals = () => {
        this.setState({ showMemberReservationModal: false });
        this.setState({ showNonMemberReservationModal: false });
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

            var info = { propertyId: propertyId, userId: this.props.appStateStore.me.userId };

            return (
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

                            var numMembershipStatus = property.membershipStatusConstraints.length;

                            var me = property.me

                            var allowCancel = {}
                            for (var i=0; i<property.cancelReservationConstraints.cancelReservationAllowed.length; i++) {
                                allowCancel[property.cancelReservationConstraints.cancelReservationAllowed[i]] = true;
                            }
                            console.log(allowCancel);
                        }
                        return (
                            <Container>
                                {/* Ledger Balance */}
                                <Table bordered size="sm">
                                    <thead>
                                        <tr>
                                            <th>Current Balance</th>
                                            <th>Minimum Balance</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr key="abc">
                                            <td>{property.ledgers[0].records[0].balance}</td>
                                            <td>{property.settings.minBalance}</td>
                                        </tr>
                                    </tbody>
                                </Table>


                                {/* Membership status */}
                                {numMembershipStatus > 0 && <Membership isModal={false} isAdmin={false} />}

                                {/* Make Reservation Modal */}
                                <CreateReservation adminRequest={false} member={true} constraints={property.memberConstraints} showModal={this.state.showMemberReservationModal} exitModal={this.turnOffModals} />
                                <CreateReservation adminRequest={false} member={false} constraints={property.nonMemberConstraints} showModal={this.state.showNonMemberReservationModal} exitModal={this.turnOffModals} />
                                <div className="text-center">
                                    {property.memberConstraints.newReservationAllowed &&
                                        <Button color="primary" onClick={() => this.displayMemberReservationModal()}>Member<br />Reservation</Button>
                                    }
                                    {' '}
                                    {property.nonMemberConstraints.newReservationAllowed &&
                                        <Button color="primary" onClick={() => this.displayNonMemberReservationModal()}>Non-Member<br />Reservation</Button>
                                    }
                                </div>

                                {/* Reservations List */}
                                {numReservations > 0 && <div><Label check>
                                    <Input type="checkbox" onClick={this.toggleShowMyReservations} checked={this.state.showMyReservations ? 'checked' : ''} onChange={() => { }} />{' '}
                                    Show only my reservations
                                    </Label><br /></div>}
                                {numReservations > 0 && <div><Label check>
                                    <Input type="checkbox" onClick={this.toggleIncludeCanceledReservations} checked={this.state.showCanceled ? 'checked' : ''} onChange={() => { }} />{' '}
                                    Include canceled reservations
                                    </Label></div>}
                                {numReservations > 0 && Object.keys(reservations).map(key => {
                                    var showCanceled = !reservations[key].canceled || this.state.showCanceled
                                    var showMine = me.userId === reservations[key].reservedFor.userId || !this.state.showMyReservations
                                    var show = showCanceled && showMine
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

export default inject('appStateStore')(observer(Reservations))
