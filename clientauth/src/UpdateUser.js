import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Modal, ModalHeader, ModalBody
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

const UPDATE_USER_GQL_MUTATION = gql`
mutation UpdateUser(
    $propertyId: String!,
    $userId: String!,
    $input: UpdateUserInput!) {
        updateUser(propertyId: $propertyId,
        userId: $userId,
        input: $input) {
        eventVersion
        }
  }
`

class UpdateUser extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.inputValid = this.inputValid.bind(this);
        this.toggle = this.toggle.bind(this);

        this.state = {
            email: '',
            nickname: '',
            isAdmin: false,
            isMember: true,
            state: '',
            userId: '',

            invalidEmailText: null,
            invalidNicknameText: null,
            invalidIsAdminText: null,
            invalidIsMemberText: null,

            submitClicked: false,
            currentSettings: {},
        };
    }

    toggle() {
        this.setState({ submitClicked: false });
        this.setState({ invalidNickname: null });
        this.updateCurrentSettings(this.state.currentSettings);
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

    updateCurrentSettings(currentSettings) {
        this.setState({ currentSettings: currentSettings });

        this.setState({ nickname: currentSettings.nickname });
        this.setState({ email: currentSettings.email });
        this.setState({ isAdmin: currentSettings.isAdmin });
        this.setState({ isMember: currentSettings.isMember });
        this.setState({ state: currentSettings.state });
        this.setState({ userId: currentSettings.userId });
    }

    componentDidMount() {
        this.updateCurrentSettings(this.props.currentSettings)
    }

    render() {
        const apolloClient = this.props.appStateStore.apolloClient;

        const showform = this.props.showform ? true : false;

        console.log("update user render");
        console.log(this.props.currentSettings);

        console.log("user constraints:");
        console.log(this.props.userConstraints);

        // Form to create a new content template
        return (
            <Modal isOpen={showform} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>Update User</ModalHeader>
                <ModalBody>

                    <Mutation client={apolloClient} mutation={UPDATE_USER_GQL_MUTATION} fetchPolicy='no-cache'

                        onCompleted={(data) => {
                            if (data.updateUser !== undefined) {
                                this.props.appStateStore.setPropertyEventVersion(data.updateUser.eventVersion);
                            }
                            this.toggle();
                        }}>
                        {(updateUserSubmit, { loading, error }) => {
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
                                                propertyId: this.props.appStateStore.propertyId,
                                                userId: this.state.userId,
                                                input: {
                                                    forVersion: this.props.appStateStore.propertyEventVersion,
                                                    nickname: this.state.nickname,
                                                    email: this.state.email,
                                                    isAdmin: this.state.isAdmin,
                                                    isMember: this.state.isMember,
                                                    state: this.state.state,
                                                }
                                            }

                                            updateUserSubmit({
                                                variables: info
                                            });
                                        }
                                    }}
                                    >

                                        <FormGroup>
                                            <Label for="nickname">Nickname</Label>
                                            <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidNickname ? true : false}
                                                type="text" name="nickname" id="nickname" placeholder="Nickname" value={this.state.nickname} />
                                            {this.state.invalidNickname &&
                                                <FormFeedback>{this.state.invalidNickname}</FormFeedback>}
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
                console.log("invalid email: " + invalidEmail);
                var email = runningState.email
                if (email.trim().toLowerCase() === invalidEmail.trim().toLowerCase() &&
                    this.state.userId !== this.props.users[i].userId) {
                    valid = false;
                    this.setState({ invalidEmailText: 'Already exists.' });
                }
            }
        }

        if (runningState.nickname.length < this.props.userConstraints.nicknameMin) {
            valid = false;
            this.setState({ invalidNickname: 'Invalid name.' });
        } else if (runningState.nickname.length > this.props.userConstraints.nicknameMax) {
            valid = false;
            this.setState({ invalidNickname: 'Name is too long.' });
        } else {
            this.setState({ invalidNickname: null });
        }

        if (runningState.nickname.length > 0) {
            for (i = 0; i < this.props.users.length; i++) {
                var invalidNickname = this.props.users[i].nickname;
                var nickname = runningState.nickname
                if (nickname.trim().toLowerCase() === invalidNickname.trim().toLowerCase() &&
                    this.state.userId !== this.props.users[i].userId) {
                    valid = false;
                    this.setState({ invalidNickname: 'Already exists.' });
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

export default inject('appStateStore')(observer(UpdateUser))

