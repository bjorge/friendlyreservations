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
import CurrencyInput from 'react-currency-input';


const CREATE_CONTENT_GQL_MUTATION = gql`
mutation AdminUpdateBalance(
    $propertyId: String!,
    $input: UpdateBalanceInput!) {
        updateBalance(propertyId: $propertyId, input: $input) {
            eventVersion
        } 
}
`;

class UpdateBalance extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.inputValid = this.inputValid.bind(this);
        this.toggle = this.toggle.bind(this);
        this.handleCurrencyChange = this.handleCurrencyChange.bind(this);

        this.state = {
            increase: true,
            amount: 0.00,
            description: '',

            invalidDescription: null,
            invalidAmount: null,

            submitClicked: false,
        };
    }

    toggle() {
        this.setState({ submitClicked: false });
        this.setState({ invalidDescription: null });
        this.setState({ invalidAmount: null });
        this.props.exitModal();
    }

    handleCurrencyChange(event, maskedvalue, floatvalue) {
        var name = 'amount'
        this.setState({
            [name]: floatvalue
        });

        var runningState = { ...this.state };
        runningState[name] = floatvalue;

        this.inputValid(runningState);
    }

    handleRadioChange(event) {
        var name = 'increase';
        var value = this.state.increase ? false : true;
        this.setState({
            [name]: value
        });

        var runningState = { ...this.state };
        runningState[name] = value;

        this.inputValid(runningState);
    }

    handleChange(event) {
        const target = event.target;
        const value = target.type === 'checkbox' ? target.checked : target.value;
        const name = target.name;

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

        const showform = this.props.showform ? true : false;

        const user = this.props.user;

        // Form to create a new content template
        return (
            <Modal isOpen={showform} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>Update Balance for {user.nickname}</ModalHeader>
                <ModalBody>

                    <Mutation client={apolloClient} mutation={CREATE_CONTENT_GQL_MUTATION} fetchPolicy='no-cache'

                        onCompleted={(data) => {
                            console.log("PaymentForm onCompleted data: ");
                            console.log(data);
                            if (data.updateBalance !== undefined) {
                                this.props.appStateStore.setPropertyEventVersion(data.updateBalance.eventVersion);
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
                                                    increase: this.state.increase,
                                                    amount: Math.trunc(this.state.amount * 100),
                                                    description: this.state.description,
                                                    updateForUserId: user.userId,
                                                }
                                            }
                                            console.log("to submit: ");
                                            console.log(info);
                                            newContentSubmit({
                                                variables: info
                                            });
                                        }
                                    }}
                                    >

                                        <FormGroup row>
                                            <Label for="amount">Amount</Label>
                                            <Col sm={10}>
                                                <div id="amount">
                                                    <CurrencyInput className="form-control" value={this.state.amount} onChangeEvent={this.handleCurrencyChange} />
                                                </div>
                                                {this.state.invalidAmount &&
                                                    <div className="invalid-feedback d-block">
                                                        {this.state.invalidAmount}
                                                    </div>
                                                }
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row tag="fieldset">
                                            {/* <legend>Radio Buttons</legend> */}
                                            <FormGroup check inline>
                                                <Label check>
                                                    <Input onChange={(e) => { this.handleRadioChange(e) }} type="radio" name="change"
                                                        checked={this.state.increase ? 'checked' : ''} />{' '}
                                                    Increase
                                                </Label>
                                            </FormGroup>
                                            <FormGroup check inline>
                                                <Label check>
                                                    <Input onChange={(e) => { this.handleRadioChange(e) }} type="radio" name="change"
                                                        checked={this.state.increase ? '' : 'checked'} />{' '}
                                                    Decrease
                                                </Label>
                                            </FormGroup>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="description">Description</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidDescription ? true : false}
                                                    type="text" name="description" id="description" 
                                                    placeholder={this.state.increase ? "Payment reference, check number, etc." : "Expense description, invoice number, etc."}
                                                    value={this.state.description} />
                                                {this.state.invalidDescription &&
                                                    <FormFeedback>{this.state.invalidDescription}</FormFeedback>}
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

        console.log("runnin state:");
        console.log(runningState);

        if (!runningState.submitClicked) {
            return true;
        }

        // submit has been clicked, so now validate the settings
        var valid = true;

        if (runningState.description.length < this.props.balanceConstraints.descriptionMin) {
            valid = false;
            this.setState({ invalidDescription: 'Invalid description.' });
        } else if (runningState.description.length > this.props.balanceConstraints.descriptionMax) {
            valid = false;
            this.setState({ invalidDescription: 'Description is too long.' });
        } else {
            this.setState({ invalidDescription: null });
        }

        if (runningState.amount >= this.props.balanceConstraints.amountMin/100 &&
            runningState.amount <= this.props.balanceConstraints.amountMax/100) {
            this.setState({ invalidAmount: null });
        } else {
            valid = false;
            this.setState({ invalidAmount: 'Please enter a valid amount.' });
        }

        return valid;
    }
}

export default inject('appStateStore')(observer(UpdateBalance))

