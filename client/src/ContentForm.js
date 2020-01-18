import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Modal, ModalHeader, ModalBody, Col
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

const CREATE_CONTENT_GQL_MUTATION = gql`
mutation NewContent(
    $propertyId: String!,
    $input: NewContentInput!) {
        createContent(propertyId: $propertyId, input: $input) {
            eventVersion
        } 
}
`;

class ContentForm extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.inputValid = this.inputValid.bind(this);
        this.toggle = this.toggle.bind(this);

        this.state = {
            template: '',
            comment: '',

            invalidTemplateText: null,
            invalidCommentText: null,

            submitClicked: false,
        };
    }

    toggle() {
        this.setState({ submitClicked: false });
        this.setState({ invalidTemplateText: null });
        this.setState({ invalidCommentText: null });
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
        const apolloClient = this.props.appStateStore.apolloClient;

        const content = this.props.content ? this.props.content : null;

        const showform = this.props.showform ? true : false;

        // console.log("ContentForm render, content is: ");
        // console.log(content);


        // Form to create a new content template
        return (
            <Modal isOpen={showform} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>New Template for {content.name}</ModalHeader>
                <ModalBody>

                    <Mutation client={apolloClient} mutation={CREATE_CONTENT_GQL_MUTATION} fetchPolicy='no-cache'

                        onCompleted={(data) => {
                            console.log("ContentForm onCompleted data: ");
                            console.log(data);
                            if (data.createContent !== undefined) {
                                console.log("set event version to: " + data.createContent.eventVersion);
                                this.props.appStateStore.setPropertyEventVersion(data.createContent.eventVersion);
                            }
                            this.toggle();
                        }}>
                        {(newContentSubmit, { loading, error }) => {
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
                                                input: {
                                                    forVersion: this.props.appStateStore.propertyEventVersion,
                                                    name: content.name,
                                                    template: this.state.template,
                                                    comment: this.state.comment,
                                                }
                                            }
                                            newContentSubmit({
                                                variables: info
                                            });
                                        }
                                    }}
                                    >
                                        <FormGroup row>
                                            <Label for="template" sm={2}>Custom Template</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidTemplateText ? true : false}
                                                    type="textarea" name="template" id="template" placeholder="New template" value={this.state.template} />
                                                {this.state.invalidTemplateText &&
                                                    <FormFeedback>{this.state.invalidTemplateText}</FormFeedback>}
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="comment" sm={2}>Comment</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidCommentText ? true : false}
                                                    type="text" name="comment" id="comment" placeholder="Reason for change" value={this.state.comment} />
                                                {this.state.invalidCommentText &&
                                                    <FormFeedback>{this.state.invalidCommentText}</FormFeedback>}
                                            </Col>
                                        </FormGroup>

                                        <hr />
                                        <FormGroup row>
                                            <Label for="id1" sm={2}>Default Template</Label>
                                            <Col sm={10}>
                                                <Input readOnly={true} type="textarea" name="name2" id="id3" value={content.defaultTemplate} />
                                            </Col>
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

        if (runningState.template) {
            this.setState({ invalidTemplateText: null });
        } else {
            valid = false;
            this.setState({ invalidTemplateText: 'Please enter a template.' });
        }

        if (runningState.comment) {
            this.setState({ invalidCommentText: null });
        } else {
            valid = false;
            this.setState({ invalidCommentText: 'Please enter a comment.' });
        }

        return valid;
    }
}

export default inject('appStateStore')(observer(ContentForm))

