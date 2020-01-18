import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input, FormFeedback, Table
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import CurrencyInput from 'react-currency-input';

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

class MembershipRestriction extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.handleCurrencyChange = this.handleCurrencyChange.bind(this);

        this.state = {
            amount: 300.00,
            descriptionInValid: null,
            inDateInValid: null,
            outDateInValid: null,
            gracePeriodOutDateInValid: null,
            prePayStartDateInValid: null,
            amountInvalid: null,
        };
    }

    render() {
        const showView = this.props.viewinfo ? true : false;
        const inDate = showView ? this.props.viewinfo.restriction.inDate : 0;
        const outDate = showView ? this.props.viewinfo.restriction.outDate : "";
        const gracePeriodOutDate = showView ? this.props.viewinfo.restriction.gracePeriodOutDate : "";
        const prePayStartDate = showView ? this.props.viewinfo.restriction.prePayStartDate : "";
        const amount = showView ? this.props.viewinfo.restriction.amount : "";

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
                        <th scope="row">{"First checkin date"}</th>
                        <td>{inDate}</td>
                    </tr>
                    <tr>
                        <th scope="row">{"Last checkout date"}</th>
                        <td>{outDate}</td>
                    </tr>
                    <tr>
                        <th scope="row">{"Grace period checkout date"}</th>
                        <td>{gracePeriodOutDate}</td>
                    </tr>
                    <tr>
                        <th scope="row">{"Pre-pay date"}</th>
                        <td>{prePayStartDate}</td>
                    </tr>
                    <tr>
                        <th scope="row">{"Amount"}</th>
                        <td>{amount}</td>
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
                                    const data = new FormData(event.target);

                                    var doSubmit = true;
                                    if (!data.get('description')) {
                                        doSubmit = false;
                                        this.setState({ descriptionInValid: 'Please enter a description.' })
                                    }
                                    if (!data.get('inDate')) {
                                        doSubmit = false;
                                        this.setState({ inDateInValid: 'Please enter a start date.' })
                                    }
                                    if (!data.get('outDate')) {
                                        doSubmit = false;
                                        this.setState({ outDateInValid: 'Please enter an end date.' })
                                    }

                                    if (!data.get('gracePeriodOutDate')) {
                                        doSubmit = false;
                                        this.setState({ gracePeriodOutDateInValid: 'Please enter a grace period date.' })
                                    }

                                    if (!data.get('prePayStartDate')) {
                                        doSubmit = false;
                                        this.setState({ prePayStartDateInValid: 'Please enter a pre payment start date.' })
                                    }

                                    if (doSubmit) {
                                        // ok, we can submit! let's setup a cool gql mutation
                                        var info = {
                                            propertyId: this.props.appStateStore.propertyId,
                                            input: {
                                                membership: {
                                                    inDate: data.get('inDate'),
                                                    outDate: data.get('outDate'),
                                                    gracePeriodOutDate: data.get('gracePeriodOutDate'),
                                                    prePayStartDate: data.get('prePayStartDate'),
                                                    amount: Math.trunc(this.state.amount * 100),
                                                },
                                                description: data.get('description'),
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
                                        <Input onChange={this.handleChange} invalid={this.state.descriptionInValid ? true : false} type="text" name="description" id="description" placeholder="Description" />
                                        {this.state.descriptionInValid &&
                                            <FormFeedback>{this.state.descriptionInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="inDate">First Checkin Date</Label>
                                        <Input onChange={this.handleChange} invalid={this.state.inDateInValid ? true : false} type="date" name="inDate" id="inDate" placeholder="First allowed checkin date" />
                                        {this.state.inDateInValid &&
                                            <FormFeedback>{this.state.inDateInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="outDate">Last Checkout Date</Label>
                                        <Input onChange={this.handleChange} invalid={this.state.outDateInValid ? true : false} type="date" name="outDate" id="outDate" placeholder="Last allowed checkout date" />
                                        {this.state.outDateInValid &&
                                            <FormFeedback>{this.state.outDateInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="gracePeriodOutDate">Last Grace Period Checkout Date</Label>
                                        <Input onChange={this.handleChange} invalid={this.state.gracePeriodOutDateInValid ? true : false} type="date" name="gracePeriodOutDate" id="gracePeriodOutDate" placeholder="Last allowed checkout date" />
                                        {this.state.gracePeriodOutDateInValid &&
                                            <FormFeedback>{this.state.gracePeriodOutDateInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="prePayStartDate">Pre Pay Start Date</Label>
                                        <Input onChange={this.handleChange} invalid={this.state.prePayStartDateInValid ? true : false} type="date" name="prePayStartDate" id="prePayStartDate" placeholder="First pre pay day" />
                                        {this.state.prePayStartDateInValid &&
                                            <FormFeedback>{this.state.prePayStartDateInValid}</FormFeedback>}
                                    </FormGroup>

                                    <FormGroup>
                                        <Label for="amount">Amount</Label>
                                            <div id="amount">
                                                <CurrencyInput name="amount" id="amount" className="form-control" value={this.state.amount} onChangeEvent={this.handleCurrencyChange} />
                                            </div>
                                            {this.state.amountInvalid &&
                                                <div className="invalid-feedback d-block">
                                                    {this.state.amountInvalid}
                                                </div>
                                            }
                                    </FormGroup>

                                    <Button type="submit">Submit</Button>

                                </Form>

                            </div>);
                    }}
                </Mutation>);
        }
    }

    handleCurrencyChange(event, maskedvalue, floatvalue) {
        var name = event.target.name;
        console.log("handle currency change name: " + name + " value: " + floatvalue);
        this.setState({
            [name]: floatvalue
        });
    }

    handleChange(event) {
        const target = event.target;
        switch (target.name) {
            case 'description':
                this.setState({ descriptionInValid: null })
                break;
            case 'inDate':
                this.setState({ inDateInValid: null })
                break;
            case 'outDate':
                this.setState({ outDateInValid: null })
                break;
            case 'gracePeriodOutDate':
                this.setState({ gracePeriodOutDateInValid: null })
                break;
            case 'prePayStartDate':
                this.setState({ prePayStartDateInValid: null })
                break;
            default:
                break;
        }
    }
}

export default inject('appStateStore')(observer(MembershipRestriction))

