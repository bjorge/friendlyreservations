import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Modal, ModalHeader, ModalBody, Table
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import UpdateBalance from './UpdateBalance';
import CreateReservation from './CreateReservation';
import Membership from './Membership';
import LedgerView from './LedgerView';
import UpdateUser from './UpdateUser';

// make the button link looks like other links
var buttonStyle = {
    padding: '0',
    verticalAlign: 'baseline'
};

const CREATE_USER_GQL_MUTATION = gql`
mutation NewUser(
    $propertyId: String!,
    $input: NewUserInput!) {
        createUser(
            propertyId: $propertyId, 
            input: $input) {
                eventVersion
    }
}
`;


class User extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.inputValid = this.inputValid.bind(this);
        this.toggle = this.toggle.bind(this);
        this.toggleShowAdvanced = this.toggleShowAdvanced.bind(this);

        this.displayUpdateBalanceForm = this.displayUpdateBalanceForm.bind(this);
        this.displayUpdateUserForm = this.displayUpdateUserForm.bind(this);
        this.displayMemberReservationModal = this.displayMemberReservationModal.bind(this);
        this.displayNonMemberReservationModal = this.displayNonMemberReservationModal.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);

        this.state = {
            cachedProperty: null,
            showAdvanced: false,

            email: '',
            nickname: '',
            isAdmin: false,
            isMember: true,

            invalidEmailText: null,
            invalidNicknameText: null,
            invalidIsAdminText: null,
            invalidIsMemberText: null,

            submitClicked: false,

            showUpdateBalanceForm: false,
            showUpdateUserForm: false,
            showMemberReservationModal: false,
            showNonMemberReservationModal: false,
            showMembershipsModal: false,
            showLedgerModal: false,

        };
    }

    toggleShowAdvanced() {
        this.setState({
            showAdvanced: !this.state.showAdvanced
        });
    }

    displayUpdateBalanceForm() {
        this.setState({
            showUpdateBalanceForm: true
        });
    }

    displayUpdateUserForm() {
        this.setState({
            showUpdateUserForm: true
        });
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

    displayMembershipsModal() {
        this.setState({
            showMembershipsModal: true
        });
    }

    displayLedgerModal() {
        this.setState({
            showLedgerModal: true
        });
    }

    turnOffModals = () => {
        this.setState({ showUpdateBalanceForm: false });
        this.setState({ showUpdateUserForm: false });
        this.setState({ showMemberReservationModal: false });
        this.setState({ showNonMemberReservationModal: false });
        this.setState({ showMembershipsModal: false });
        this.setState({ showLedgerModal: false });
    }

    toggle() {
        this.setState({ submitClicked: false });
        this.setState({ invalidEmailText: null });
        this.setState({ invalidNicknameText: null });
        this.setState({ invalidIsAdminText: null });
        this.setState({ invalidIsMemberText: null });
        this.props.exitModal();
    }

    handleChange(event) {
        const target = event.target;
        const value = target.type === 'checkbox' ? target.checked : target.value;
        const name = target.name;

        this.setState({
            [name]: value
        });

        var runningState = { ...this.state };
        runningState[name] = value;

        this.inputValid(runningState);
    }

    render() {
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        const showView = this.props.viewinfo ? true : false;

        const apolloClient = this.props.appStateStore.apolloClient;

        if (showView) {

            // TODO: just use the prop below rather than this logic
            const state = showView ? this.props.viewinfo.state : "";
            const email = showView ? this.props.viewinfo.email : "";
            const nickname = showView ? this.props.viewinfo.nickname : "";
            const isAdmin = showView ? this.props.viewinfo.isAdmin : "";
            const isMember = showView ? this.props.viewinfo.isMember : "";
            const isSystem = showView ? this.props.viewinfo.isSystem : "";

            return (
                <div>
                    <Table bordered size="sm">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Value</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <th scope="row">{"Nickname"}</th>
                                <td>{nickname}</td>
                            </tr>
                            <tr>
                                <th scope="row">{"Email"}</th>
                                <td>{email}</td>
                            </tr>
                            <tr>
                                <th scope="row">{"State"}</th>
                                <td>{state}</td>
                            </tr>
                            <tr>
                                <th scope="row">{"Member"}</th>
                                <td>{isMember ? 'yes' : 'no'}</td>
                            </tr>
                            <tr>
                                <th scope="row">{"Administrator"}</th>
                                <td>{isAdmin ? 'yes' : 'no'}</td>
                            </tr>
                        </tbody>
                    </Table>

                    <Table bordered size="sm">
                        <tbody>
                            <tr>
                                <td>
                                    {/* Update Balance */}
                                    {isSystem === false && <UpdateBalance user={this.props.viewinfo} balanceConstraints={this.props.balanceConstraints} showform={this.state.showUpdateBalanceForm} exitModal={this.turnOffModals} />}
                                    {isSystem === false &&
                                        <Button style={buttonStyle} color="link" onClick={() => this.displayUpdateBalanceForm()}>Update Balance</Button>
                                    }
                                </td>
                            </tr>
                        </tbody>
                    </Table>

                    <FormGroup check inline>
                        {!this.state.showAdvanced && <Button style={buttonStyle} color="link" onClick={() => this.toggleShowAdvanced()}>Show Advanced</Button>}
                        {this.state.showAdvanced && <Button style={buttonStyle} color="link" onClick={() => this.toggleShowAdvanced()}>Hide Advanced</Button>}
                    </FormGroup>

                    <br />
                    {this.state.showAdvanced && <Table bordered size="sm">
                        <tbody>
                            <tr>
                                <td>
                                    {/* Update User */}
                                    {isSystem === false && <UpdateUser users={this.props.users} userConstraints={this.props.userConstraints} currentSettings={this.props.viewinfo} showform={this.state.showUpdateUserForm} exitModal={this.turnOffModals} />}
                                    {isSystem === false &&
                                        <Button style={buttonStyle} color="link" onClick={() => this.displayUpdateUserForm()}>Update User</Button>
                                    }
                                    <br />

                                    {/*Ledger Modal*/}
                                    <LedgerView isModal={true} isAdmin={true} user={this.props.viewinfo} showModal={this.state.showLedgerModal} exitModal={this.turnOffModals} />
                                    {this.state.showAdvanced && <Button style={buttonStyle} color="link" onClick={() => this.displayLedgerModal()}>Show Ledger</Button>}
                                    <br />

                                    {/* Make Reservation Modal */}
                                    <CreateReservation adminRequest={true} reservedForUser={this.props.viewinfo} member={true} constraints={this.props.reservationConstraints} showModal={this.state.showMemberReservationModal} exitModal={this.turnOffModals} />
                                    <CreateReservation adminRequest={true} reservedForUser={this.props.viewinfo} member={false} constraints={this.props.reservationConstraints} showModal={this.state.showNonMemberReservationModal} exitModal={this.turnOffModals} />
                                    {this.state.showAdvanced && this.props.reservationConstraints.newReservationAllowed &&
                                        <Button style={buttonStyle} color="link" onClick={() => this.displayMemberReservationModal()}>Make Reservation (Member)</Button>
                                    }
                                    <br />

                                    {this.state.showAdvanced && this.props.reservationConstraints.newReservationAllowed &&
                                        <Button style={buttonStyle} color="link" onClick={() => this.displayNonMemberReservationModal()}>Make Reservation (Non-Member)</Button>
                                    }
                                    <br />

                                    {/*Membership Modal*/}
                                    <Membership isModal={true} isAdmin={true} user={this.props.viewinfo} showModal={this.state.showMembershipsModal} exitModal={this.turnOffModals} />
                                    {this.state.showAdvanced && <Button style={buttonStyle} color="link" onClick={() => this.displayMembershipsModal()}>Membership</Button>}

                                </td>
                            </tr>
                        </tbody>
                    </Table>}

                </div>
            );
        } else {
            // Form to create a new user
            const showform = this.props.showform ? true : false;
            return (
                <Modal isOpen={showform} toggle={this.toggle}>
                    <ModalHeader toggle={this.toggle}>New User</ModalHeader>
                    <ModalBody>

                        <Mutation client={apolloClient} mutation={CREATE_USER_GQL_MUTATION} fetchPolicy='no-cache'
                            onCompleted={(data) => {
                                if (data.createUser !== undefined) {
                                    this.props.appStateStore.setPropertyEventVersion(data.createUser.eventVersion);
                                }
                                this.toggle();
                            }}>
                            {(newUserSubmit, { loading, error }) => {
                                if (loading) return (<Spinner />);
                                return (
                                    <div>
                                        {error && <ErrorModal error={error} />}
                                        <Form onSubmit={event => {
                                            event.preventDefault();

                                            this.setState({ submitClicked: true });
                                            var runningState = { ...this.state };
                                            runningState['submitClicked'] = true;

                                            if (this.inputValid(runningState)) {
                                                // ok, we can submit! let's setup a cool gql mutation
                                                var info = {
                                                    propertyId: propertyId,
                                                    input: {
                                                        forVersion: this.props.appStateStore.propertyEventVersion,
                                                        email: this.state.email,
                                                        isAdmin: this.state.isAdmin,
                                                        isMember: this.state.isMember,
                                                        nickname: this.state.nickname,
                                                    }
                                                }
                                                newUserSubmit({
                                                    variables: info
                                                });
                                            }
                                        }}
                                        >
                                            <FormGroup>
                                                <Label for="nickname">Nickname</Label>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidNicknameText ? true : false}
                                                    type="text" name="nickname" id="nickname" placeholder="Nickname" value={this.state.nickname} />
                                                {this.state.invalidNicknameText &&
                                                    <FormFeedback>{this.state.invalidNicknameText}</FormFeedback>}
                                            </FormGroup>

                                            <FormGroup>
                                                <Label for="email">Email</Label>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidEmailText ? true : false}
                                                    type="email" name="email" id="email" placeholder="Email" value={this.state.email} />
                                                {this.state.invalidEmailText &&
                                                    <FormFeedback>{this.state.invalidEmailText}</FormFeedback>}
                                            </FormGroup>

                                            <FormGroup check>
                                                <Label check>
                                                    <Input onChange={(e) => { this.handleChange(e) }} type="checkbox" name="isAdmin"
                                                        checked={this.state.isAdmin ? 'checked' : ''} invalid={this.state.invalidIsAdminText ? true : false} />{' '}
                                                    Administrator?
                                                </Label>
                                                {this.state.invalidIsAdminText &&
                                                    <div class="invalid-feedback d-block">
                                                        {this.state.invalidIsAdminText}
                                                    </div>}
                                            </FormGroup>
                                            <FormGroup check>
                                                <Label check>
                                                    <Input onChange={(e) => { this.handleChange(e) }} type="checkbox" name="isMember"
                                                        checked={this.state.isMember ? 'checked' : ''} />{' '}
                                                    Member?
                                                </Label>
                                                {/*add in order to display feedback until bug is fixed: https://github.com/twbs/bootstrap/pull/25210*/}
                                                {this.state.invalidIsMemberText &&
                                                    <div class="invalid-feedback d-block">
                                                        {this.state.invalidIsMemberText}
                                                    </div>
                                                }
                                            </FormGroup>

                                            <div className="text-center">
                                                <Button color="primary" type="submit">Submit</Button>
                                            </div>

                                        </Form>

                                    </div>);
                            }}
                        </Mutation>

                    </ModalBody>
                </Modal>
            );
        }
    }

    inputValid(runningState) {

        if (!runningState.submitClicked) {
            return true;
        }

        // submit has been clicked, so now validate the settings
        var valid = true;

        if (runningState.email.length < this.props.userConstraints.emailMin) {
            valid = false;
            this.setState({ invalidEmailText: 'Invalid name.' });
        } else if (runningState.email.length > this.props.userConstraints.emailMax) {
            valid = false;
            this.setState({ invalidEmailText: 'Name is too long.' });
        } else {
            this.setState({ invalidEmailText: null });
        }

        if (runningState.email.length > 0) {
            for (var i = 0; i < this.props.users.length; i++) {
                var invalidEmail = this.props.users[i].email;
                var email = runningState.email
                if (email.trim().toLowerCase() === invalidEmail.trim().toLowerCase()) {
                    valid = false;
                    this.setState({ invalidEmailText: 'Already exists' });
                }
            }
        }

        ///

        if (runningState.nickname.length < this.props.userConstraints.nicknameMin) {
            valid = false;
            this.setState({ invalidNicknameText: 'Invalid name.' });
        } else if (runningState.nickname.length > this.props.userConstraints.nicknameMax) {
            valid = false;
            this.setState({ invalidNicknameText: 'Name is too long.' });
        } else {
            this.setState({ invalidNicknameText: null });
        }

        if (runningState.nickname.length > 0) {
            for (i = 0; i < this.props.users.length; i++) {
                var invalidNickname = this.props.users[i].nickname;
                var nickname = runningState.nickname
                if (nickname.trim().toLowerCase() === invalidNickname.trim().toLowerCase() &&
                    this.state.userId !== this.props.users[i].userId) {
                    valid = false;
                    this.setState({ invalidNicknameText: 'Already exists.' });
                }
            }
        }

        if (runningState.isAdmin || runningState.isMember) {
            this.setState({ invalidIsAdminText: null });
            this.setState({ invalidIsMemberText: null });
        } else {
            valid = false;
            this.setState({ invalidIsAdminText: 'Either admin or member must be checked' });
            this.setState({ invalidIsMemberText: 'Either admin or member must be checked' });
        }


        return valid;
    }
}

export default inject('appStateStore')(observer(User))

