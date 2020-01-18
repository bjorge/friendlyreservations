import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Modal, ModalBody, ModalHeader,
    Card,
    CardBody,
    Label, Input, FormFeedback,
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";

import DateRangePicker from "./DateRangePicker";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

import 'bootstrap/dist/css/bootstrap.css';

// CREATE_PROPERTY_GQL_QUERY, details: https://www.apollographql.com/docs/react/essentials/mutations.html
const CREATE_RESERVATION_GQL_QUERY = gql`
    mutation NewReservation(
    $propertyId: String!,
    $forVersion: Int!,
    $reservedForUserId: String!,
    $startDate: String!,
    $endDate: String!,
    $member: Boolean!,
    $nonMemberName: String,
    $nonMemberInfo: String,
    $adminRequest: Boolean!) {
        createReservation(propertyId: $propertyId, input: {
            forVersion: $forVersion,
            reservedForUserId: $reservedForUserId,
            startDate: $startDate,
            endDate: $endDate,
            member: $member,
            nonMemberName: $nonMemberName,
            nonMemberInfo: $nonMemberInfo,
            adminRequest: $adminRequest
        })  
        {
            propertyId
            settings {
                propertyName
            }
            me {
                state
                isAdmin
                isMember
                nickname
                userId
            }
        }
    }
`;

class CreateReservation extends Component {

    constructor(props) {
        super(props);

        this.inputValid = this.inputValid.bind(this);
        this.afterSubmit = this.afterSubmit.bind(this);
        this.toggleModal = this.toggleModal.bind(this);
        this.handleChange = this.handleChange.bind(this);

        this.state = {
            checkin: undefined,
            checkout: undefined,
            submitClicked: false,
            nonMemberName: '',
            nonMemberInfo: '',

            invalidDateText: null,
            invalidNonMemberName: null,
            invalidNonMemberInfo: null,
        };

    }

    afterSubmit() {
        this.setState({ submitClicked: false });
        this.setState({ checkin: undefined });
        this.setState({ checkout: undefined });
        this.setState({ nonMemberName: '' });
        this.setState({ nonMemberInfo: '' });
        this.setState({ invalidDateText: null });
        this.toggleModal();
    }

    onUpdateDatePickerRange = (range) => {
        console.log("range is:");
        console.log(range);
        console.log("state is:");
        console.log(this.state);

        var inDate = undefined;
        if (range.from) {
            inDate = range.from.toISOString().substring(0, 10);
        }

        var outDate = undefined;
        if (range.to) {
            outDate = range.to.toISOString().substring(0, 10);
        }

        this.setState({
            checkin: inDate,
            checkout: outDate,
        });

        var runningState = { ...this.state };
        runningState['checkin'] = inDate;
        runningState['checkout'] = outDate;
        this.inputValid(runningState);
    }

    toggleModal() {
        this.props.exitModal();
    }

    handleChange(event) {

        const target = event.target;
        const name = target.name;
        const value = target.value;

        console.log("handle change name: " + name + " value: " + value);

        this.setState({
            [name]: value
        });

        var runningState = { ...this.state };
        runningState[name] = value;

        this.inputValid(runningState);
    }

    render() {
        const apolloClient = this.props.appStateStore.apolloClient;

        const checkinDisabled = this.props.constraints.checkinDisabled;
        const checkoutDisabled = this.props.constraints.checkoutDisabled;
        const member = this.props.member;
        console.log("member: "+member);
        const adminRequest = this.props.adminRequest;
        console.log("adminRequest: "+adminRequest);
        var user = this.props.appStateStore.me;
        if (adminRequest) {
            user = this.props.reservedForUser;
        }

        return (
            <Modal isOpen={this.props.showModal} toggle={this.toggleModal}>
                <ModalHeader toggle={this.toggleModal}>Select dates!</ModalHeader>
                <ModalBody>
                    <Mutation client={apolloClient} mutation={CREATE_RESERVATION_GQL_QUERY} fetchPolicy='no-cache' onCompleted={(data) => {

                        if (data.createReservation !== undefined) {
                            this.props.appStateStore.setPropertyId(data.createReservation.propertyId);
                            this.props.appStateStore.setProperty(data.createReservation);
                            this.props.appStateStore.setPropertyEventVersion(data.createReservation.eventVersion)
                        }
                        this.afterSubmit();
                    }}>
                        {(newReservationSubmit, { loading, error }) => {
                            if (loading) return (<Spinner />);
                            return (
                                <div>
                                    {error && <ErrorModal error={error} />}
                                    <Card key="createReservation">
                                        <CardBody>
                                            <Form
                                                onSubmit={e => {
                                                    e.preventDefault();

                                                    this.setState({ submitClicked: true });
                                                    var runningState = { ...this.state };
                                                    runningState['submitClicked'] = true;

                                                    if (this.inputValid(runningState)) {
                                                        var info = {
                                                            forVersion: this.props.appStateStore.propertyEventVersion,
                                                            reservedForUserId: user.userId,
                                                            propertyId: this.props.appStateStore.propertyId,
                                                            startDate: this.state.checkin,
                                                            endDate: this.state.checkout,
                                                            member: member,
                                                            nonMemberInfo: this.state.nonMemberInfo,
                                                            nonMemberName: this.state.nonMemberName,
                                                            adminRequest: this.props.adminRequest,
                                                        }
                                                        console.log(info);
                                                        newReservationSubmit({
                                                            variables: info
                                                        });
                                                    }


                                                }}
                                            >
                                                <FormGroup>
                                                    {this.state.invalidDateText &&
                                                        <div className="invalid-feedback d-block">
                                                            {this.state.invalidDateText}
                                                        </div>
                                                    }
                                                    <DateRangePicker checkinDisabled={checkinDisabled} checkoutDisabled={checkoutDisabled} onUpdateRange={this.onUpdateDatePickerRange} />
                                                </FormGroup>

                                                {!member && <FormGroup>
                                                    <Label for="nonMemberName">Non Member Name</Label>
                                                    <Input type="text" name="nonMemberName" id="nonMemberName" value={this.state.nonMemberName}
                                                        placeholder="Friend's  name"
                                                        onChange={this.handleChange} invalid={this.state.invalidNonMemberName ? true : false} />
                                                    {this.state.invalidNonMemberName &&
                                                        <FormFeedback>{this.state.invalidNonMemberName}</FormFeedback>}
                                                </FormGroup>}

                                                {!member && <FormGroup>
                                                    <Label for="nonMemberInfo">Non Member Info</Label>
                                                    <Input type="text" name="nonMemberInfo" id="nonMemberInfo" value={this.state.nonMemberInfo}
                                                        placeholder="Contact info, arrival/departure time, etc."
                                                        onChange={this.handleChange} invalid={this.state.invalidNonMemberInfo ? true : false} />
                                                    {this.state.invalidNonMemberInfo &&
                                                        <FormFeedback>{this.state.invalidNonMemberInfo}</FormFeedback>}
                                                </FormGroup>}

                                                <div className="text-center">
                                                    <Button>Submit</Button>
                                                </div>
                                            </Form>
                                        </CardBody>
                                    </Card>
                                </div>
                            );
                        }}
                    </Mutation>
                </ModalBody>
            </Modal>);
    }

    inputValid(runningState) {
        console.log("inputValid start")
        if (!runningState.submitClicked) {
            return true;
        }
        console.log("inputValid runningState:");
        console.log(runningState);
        console.log("member: "+this.props.member);

        // submit has been clicked, so now validate the settings
        var valid = true;

        if (runningState.checkin && runningState.checkout) {
            this.setState({ invalidDateText: null });
            if (runningState.checkin === runningState.checkout) {
                this.setState({ invalidDateText: 'Check out date same as check in date' });
                valid = false;
            }
        } else if (runningState.checkin === undefined) {
            valid = false;
            this.setState({ invalidDateText: 'Please choose checkin date' });
        } else {
            valid = false;
            this.setState({ invalidDateText: 'Please choose checkout date' });
        }

        if (!this.props.member) {
            console.log(this.props.constraints);
            if (runningState.nonMemberName.length < this.props.constraints.nonMemberNameMin) {
                valid = false;
                this.setState({ invalidNonMemberName: 'Invalid name.' });
            } else if (runningState.nonMemberName.length > this.props.constraints.nonMemberNameMax) {
                valid = false;
                this.setState({ invalidNonMemberName: 'Name is too long.' });
            } else {
                this.setState({ invalidNonMemberName: null });
            }

            if (runningState.nonMemberInfo.length < this.props.constraints.nonMemberInfoMin) {
                valid = false;
                this.setState({ invalidNonMemberInfo: 'Invalid info.' });
            } else if (runningState.nonMemberInfo.length > this.props.constraints.nonMemberInfoMax) {
                valid = false;
                this.setState({ invalidNonMemberInfo: 'Info is too long.' });
            } else {
                this.setState({ invalidNonMemberInfo: null });
            }
        }

        return valid;
    }
}

export default inject('appStateStore')(observer(CreateReservation))


