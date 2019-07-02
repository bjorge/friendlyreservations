import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input,
    Modal, ModalHeader, ModalBody
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

const INVITATION_GQL_MUTATION = gql`
mutation Invitation(
    $propertyId: String!,
    $input: AcceptInvitationInput!) {
        acceptInvitation(
            propertyId: $propertyId, 
            input: $input) {
                propertyId
                eventVersion
                me {
                    nickname
                    userId
                    state
                    isAdmin
                    isMember
                  }
                  settings {
                    propertyName
                  }
            }
    }
`;

class InvitationModal extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.accept = this.accept.bind(this);
        this.decline = this.decline.bind(this);
        this.exit = this.exit.bind(this);

        this.state = {
            selectedOption: 'accept',
            submitClicked: false,
        };
        console.log("user component constructor:");
        console.log(this.state);
        console.log(this.props);
    }

    accept() {
        this.setState({ submitClicked: false });
        this.setState({ selectedOption: 'accept' });
        this.props.acceptCallback();
    }

    decline() {
        this.setState({ submitClicked: false });
        this.setState({ selectedOption: 'accept' });
        this.props.declineCallback();
    }

    exit() {
        this.setState({ submitClicked: false });
        this.setState({ selectedOption: 'accept' });
        this.props.exitModal();
    }

    handleChange(event) {
        const target = event.target;
        const value = target.name; // for radio buttons the name is the value...

        this.setState({
            selectedOption: value
        });

        var runningState = { ...this.state };
        runningState['selectedOption'] = value;

        console.log("handleChange running state: "+runningState)

    }

    render() {
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;
        const showform = this.props.showform ? true : false;
        console.log("Invitation component render");
        console.log(this.state);
        console.log(this.props);
        const apolloClient = this.props.appStateStore.apolloClient;

        // Form to respond to invitation
        return (
            <Modal isOpen={showform} toggle={this.exit}>
                <ModalHeader toggle={this.exit}>You have been invited to join a Friendly Reservations property!</ModalHeader>
                <ModalBody>
                    <Mutation client={apolloClient} mutation={INVITATION_GQL_MUTATION} fetchPolicy='no-cache' onCompleted={(data) => {
                        console.log("invitation result data back from server:");
                        console.log(data);

                        if (data.acceptInvitation !== undefined) {
                            this.props.appStateStore.setPropertyEventVersion(data.acceptInvitation.eventVersion);
                            this.props.appStateStore.setMe(data.acceptInvitation.me);

                            if (data.acceptInvitation.me.state === 'ACCEPTED') {
                                this.accept();
                            } else {
                                this.decline();
                            }
                        }
                    }}>
                        {(invitationSubmit, { loading, error }) => {
                            if (loading) return (<Spinner />);
                            return (
                                <div>
                                    {error && <ErrorModal error={error} />}
                                    <Form onSubmit={event => {
                                        event.preventDefault();

                                        this.setState({ submitClicked: true });
                                        var runningState = { ...this.state };
                                        runningState['submitClicked'] = true;

                                        console.log("submit new user form");
                                        // ok, we can submit! let's setup a cool gql mutation
                                        var info = {
                                            propertyId: propertyId,
                                            input: {
                                                forVersion: this.props.appStateStore.propertyEventVersion,
                                                accept: this.state.selectedOption === 'accept',
                                            }
                                        }

                                        console.log("info is:");
                                        console.log(info);
                                        invitationSubmit({
                                            variables: info
                                        });
                                    }}
                                    >
                                        <FormGroup tag="fieldset">
                                            <FormGroup check>
                                                <Label check>
                                                    <Input type="radio" name="accept"
                                                        checked={this.state.selectedOption === 'accept'}
                                                        onChange={this.handleChange} />{' '}
                                                    Accept the invitation.
                                                </Label>
                                            </FormGroup>
                                            <FormGroup check>
                                                <Label check>
                                                    <Input type="radio" name="decline"
                                                        checked={this.state.selectedOption === 'decline'}
                                                        onChange={this.handleChange} />{' '}
                                                    Decline the invitation.
                                                </Label>
                                            </FormGroup>
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

export default inject('appStateStore')(observer(InvitationModal))

