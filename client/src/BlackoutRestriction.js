import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input, FormFeedback, Table
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

const CREATE_RESTRICTION_GQL_MUTATION = gql`
mutation NewRestriction(
    $propertyId: String!,
    $input: NewRestrictionInput!) {
        createRestriction(
            propertyId: $propertyId, 
            input: $input) {
                eventVersion
            }
    }
`;

class BlackoutDayRestriction extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.inputValid = this.inputValid.bind(this);

        this.state = {
            descriptionInValid: null,
            startDateInValid: null,
            endDateInValid: null,

            submitClicked: false,

            description: '',
            startDate: '',
            endDate: '',

        };
    }

    render() {
        const showView = this.props.viewinfo ? true : false;
        const startDate = showView ? this.props.viewinfo.restriction.startDate : 0;
        const endDate = showView ? this.props.viewinfo.restriction.endDate : "";
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (showView) {
            // Form to view/remove an existing restriction
            return (
                <Table bordered size="sm">
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Value</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>{"Blackout check in"}</td>
                            <td>{startDate}</td>
                        </tr>
                        <tr>
                            <td>{"Blackout check out"}</td>
                            <td>{endDate}</td>
                        </tr>
                    </tbody>
                </Table>
            );
        } else {

            const apolloClient = this.props.appStateStore.apolloClient;
            // Form to create a new restriction
            return (
                <Mutation client={apolloClient} mutation={CREATE_RESTRICTION_GQL_MUTATION} fetchPolicy='no-cache' onCompleted={(data) => {
                    if (data.createRestriction !== undefined) {
                        // update other components with latest version
                        this.props.appStateStore.setPropertyEventVersion(data.createRestriction.eventVersion);
                    }
                }}>
                    {(newRestrictionSubmit, { loading, error }) => {
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
                                                blackout: {
                                                    startDate: this.state.startDate,
                                                    endDate: this.state.endDate
                                                },
                                                description: this.state.description,
                                                forVersion: this.props.appStateStore.propertyEventVersion,
                                            }
                                        }
                                        newRestrictionSubmit({
                                            variables: info
                                        });
                                    }
                                }}
                                >
                                    <FormGroup>
                                        <Label for="description">Description</Label>
                                        <Input value={this.state.description} onChange={this.handleChange} invalid={this.state.descriptionInValid ? true : false} type="text" name="description" id="description" placeholder="Description" />
                                        {this.state.descriptionInValid &&
                                            <FormFeedback>{this.state.descriptionInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="startDate">Start Date</Label>
                                        <Input value={this.state.startDate} onChange={this.handleChange} invalid={this.state.startDateInValid ? true : false} type="date" name="startDate" id="startDate" placeholder="Blackout start date" />
                                        {this.state.startDateInValid &&
                                            <FormFeedback>{this.state.startDateInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="endDate">End Date</Label>
                                        <Input value={this.state.endDate} onChange={this.handleChange} invalid={this.state.endDateInValid ? true : false} type="date" name="endDate" id="endDate" placeholder="Blackout end date" />
                                        {this.state.endDateInValid &&
                                            <FormFeedback>{this.state.endDateInValid}</FormFeedback>}
                                    </FormGroup>

                                    <Button type="submit">Submit</Button>

                                </Form>

                            </div>);
                    }}
                </Mutation>);
        }
    }

    handleChange(event) {

        const target = event.target;
        const name = target.name;
        var tmpValue;
        switch (target.name) {
            default:
                tmpValue = target.value;
                break;
        }
        const value = tmpValue;

        this.setState({
            [name]: value
        });

        var runningState = { ...this.state };
        runningState[name] = value;

        this.inputValid(runningState);
    }

    inputValid(runningState) {

        if (!runningState.submitClicked) {
            return true;
        }

        // submit has been clicked, so now validate the settings
        var valid = true;

        if (runningState.description) {
            this.setState({ descriptionInValid: null })
        } else {
            valid = false;
            this.setState({ descriptionInValid: 'Please enter a description.' })
        }

        if (runningState.startDate) {
            this.setState({ startDateInValid: null })
        } else {
            valid = false;
            this.setState({ startDateInValid: 'Please enter a start date.' })
        }

        if (runningState.endDate) {
            this.setState({ endDateInValid: null })
        } else {
            valid = false;
            this.setState({ endDateInValid: 'Please enter an end date.' })
        }

        if (valid) {
            var startDateObj = new Date(runningState.startDate.replace(/-/g, '/'));
            var endDateObj = new Date(runningState.endDate.replace(/-/g, '/'));
            if (endDateObj <= startDateObj) {
                this.setState({ startDateInValid: 'End date must be after start date.' });
                this.setState({ endDateInValid: 'End date must be after start date.' });
                valid = false;
            }

            for (var i = 0; i < this.props.constraints.length; i++) {
                if (runningState.description.trim() === this.props.constraints[i]) {
                    this.setState({ descriptionInValid: 'Description already used.' })
                    return false;
                }
            }
        }

        return valid;
    }
}

export default inject('appStateStore')(observer(BlackoutDayRestriction))

