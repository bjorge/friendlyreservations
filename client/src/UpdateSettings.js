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

const UPDATE_SETTINGS_GQL_MUTATION = gql`
mutation UpdateSettings(
    $propertyId: String!,
    $input: UpdateSettingsInput!) {
        updateSettings(propertyId: $propertyId,
        input: $input) {
        eventVersion
        }
  }
`

class UpdateSettings extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.inputValid = this.inputValid.bind(this);
        this.toggle = this.toggle.bind(this);
        this.handleCurrencyChange = this.handleCurrencyChange.bind(this);
        this.updateCurrentSettings = this.updateCurrentSettings.bind(this);
        this.toggleAllowNonMembers = this.toggleAllowNonMembers.bind(this);

        this.state = {
            memberRate: 0.00,
            propertyName: '',
            currency: "USD",
            allowNonMembers: true,
            nonMemberRate: 8500,
            timezone: "Pacific",
            minBalance: -100000,
            maxOutDays: 365,
            minInDays: 0,
            reservationReminderDaysBefore: 3,
            balanceReminderIntervalDays: 14,

            invalidPropertyName: null,
            invalidMemberRate: null,
            invalidNonMemberRate: null,

            invalidMinBalance: null,

            submitClicked: false,
            currentSettings: {},
        };
    }

    toggleAllowNonMembers() {
        this.setState({
            allowNonMembers: !this.state.allowNonMembers
        });
    }

    toggle() {
        this.setState({ submitClicked: false });
        this.setState({ invalidPropertyName: null });
        this.setState({ invalidMemberRate: null });
        this.setState({ invalidNonMemberRate: null });
        this.updateCurrentSettings(this.state.currentSettings);
        this.props.exitModal();
    }

    handleCurrencyChange(event, maskedvalue, floatvalue) {
        var name = event.target.name;
        this.setState({
            [name]: floatvalue
        });

        var runningState = { ...this.state };
        runningState[name] = floatvalue;

        this.inputValid(runningState);
    }

    handleRadioChange(event) {
        var name = 'allowNonMembers';
        var value = this.state.allowNonMembers ? false : true;
        this.setState({
            [name]: value
        });

        var runningState = { ...this.state };
        runningState[name] = value;

        this.inputValid(runningState);
    }

    handleChange(event) {

        const target = event.target;
        const name = target.name;
        var tmpValue;
        switch (target.name) {
            case 'allowNonMembers':
                tmpValue = (target.checked ? true : false);
                break;
            case 'currency':
                tmpValue = target.options[target.selectedIndex].value;
                break;
            case 'propertyName':
            case 'timezone':
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

    updateCurrentSettings(currentSettings) {
        this.setState({ currentSettings: currentSettings });

        this.setState({ allowNonMembers: currentSettings.allowNonMembers });
        this.setState({ memberRate: currentSettings.memberRate });
        this.setState({ propertyName: currentSettings.propertyName });
        this.setState({ currency: currentSettings.currencyAcronym });
        this.setState({ nonMemberRate: currentSettings.nonMemberRate });
        this.setState({ timezone: currentSettings.timezone });
        this.setState({ minBalance: currentSettings.minBalance });
        this.setState({ maxOutDays: currentSettings.maxOutDays });
        this.setState({ minInDays: currentSettings.minInDays });
        this.setState({ balanceReminderIntervalDays: currentSettings.balanceReminderIntervalDays });
        this.setState({ reservationReminderDaysBefore: currentSettings.reservationReminderDaysBefore });
    }

    componentDidMount() {
        this.updateCurrentSettings(this.props.currentSettings)
    }

    render() {
        const apolloClient = this.props.appStateStore.apolloClient;

        //this.updateCurrentSettings(this.props.currentSettings)

        const showform = this.props.showform ? true : false;

        // Form to create a new content template
        return (
            <Modal isOpen={showform} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>Update Settings</ModalHeader>
                <ModalBody>

                    <Mutation client={apolloClient} mutation={UPDATE_SETTINGS_GQL_MUTATION} fetchPolicy='no-cache'

                        onCompleted={(data) => {
                            if (data.updateSettings !== undefined) {
                                this.props.appStateStore.setPropertyEventVersion(data.updateSettings.eventVersion);
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
                                                    propertyName: this.state.propertyName,
                                                    currency: this.state.currency,
                                                    memberRate: Math.trunc(this.state.memberRate * 100),
                                                    allowNonMembers: this.state.allowNonMembers,
                                                    nonMemberRate: Math.trunc(this.state.nonMemberRate * 100),
                                                    timezone: this.state.timezone,
                                                    minBalance: Math.trunc(this.state.minBalance * 100),
                                                    maxOutDays: Math.trunc(this.state.maxOutDays),
                                                    minInDays: Math.trunc(this.state.minInDays),
                                                    reservationReminderDaysBefore: Math.trunc(this.state.reservationReminderDaysBefore),
                                                    balanceReminderIntervalDays: Math.trunc(this.state.balanceReminderIntervalDays)
                                                }
                                            }

                                            newContentSubmit({
                                                variables: info
                                            });
                                        }
                                    }}
                                    >

                                        <FormGroup row>
                                            <Label for="propertyName" sm={2}>Property Name</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidPropertyName ? true : false}
                                                    type="text" name="propertyName" id="propertyName" placeholder="Property Name" value={this.state.propertyName} />
                                                {this.state.invalidPropertyName &&
                                                    <FormFeedback>{this.state.invalidPropertyName}</FormFeedback>}
                                            </Col>
                                        </FormGroup>
                                        <FormGroup row>
                                            <Label for="currency" sm={2}>Currency</Label>
                                            <Col sm={10}>
                                                <Input type="select" name="currency" id="currency" onChange={this.handleChange} value={this.state.currency}>
                                                    <option value='USD'>$</option>
                                                    <option value='EUR'>€</option>
                                                </Input>
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="memberRate" sm={2}>Member Rate</Label>
                                            <Col sm={10}>
                                                <div id="memberRate">
                                                    <CurrencyInput name="memberRate" className="form-control" value={this.state.memberRate} onChangeEvent={this.handleCurrencyChange} />
                                                </div>
                                                {this.state.invalidMemberRate &&
                                                    <div className="invalid-feedback d-block">
                                                        {this.state.invalidMemberRate}
                                                    </div>
                                                }
                                            </Col>
                                        </FormGroup>

                                        <FormGroup check row>
                                            <Col sm={10}>
                                                <Label check sm={2}>
                                                    <Input onClick={this.toggleAllowNonMembers} type="checkbox" name="allowNonMembers" checked={this.state.allowNonMembers ? 'checked' : ''} onChange={this.handleChange} />
                                                    Allow Friends of Members?
                                                </Label>
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="nonMemberRate" sm={2}>Non Member Rate</Label>
                                            <Col sm={10}>
                                                <div id="nonMemberRate">
                                                    <CurrencyInput name="nonMemberRate" className="form-control" value={this.state.nonMemberRate} onChangeEvent={this.handleCurrencyChange} />
                                                </div>
                                                {this.state.invalidNonMemberRate &&
                                                    <div className="invalid-feedback d-block">
                                                        {this.state.invalidNonMemberRate}
                                                    </div>
                                                }
                                            </Col>
                                        </FormGroup>


                                        <FormGroup row>
                                            <Label for="timezone" sm={2}>Time Zone</Label>
                                            <Col sm={10}>

                                                <Input type="select" name="timezone" id="timezone" onChange={this.handleChange}
                                                    value={this.state.timezone}>
                                                    {/* <option hidden >Select a timezone</option> */}
                                                    <option value="Pacific/Pago_Pago">(GMT-11:00) Pago Pago</option>
                                                    <option value="Pacific/Honolulu">(GMT-10:00) Hawaii Time</option>
                                                    <option value="America/Los_Angeles">(GMT-08:00) Pacific Time</option>
                                                    <option value="America/Tijuana">(GMT-08:00) Pacific Time - Tijuana</option>
                                                    <option value="America/Denver">(GMT-07:00) Mountain Time</option>
                                                    <option value="America/Phoenix">(GMT-07:00) Mountain Time - Arizona</option>
                                                    <option value="America/Mazatlan">(GMT-07:00) Mountain Time - Chihuahua, Mazatlan</option>
                                                    <option value="America/Chicago">(GMT-06:00) Central Time</option>
                                                    <option value="America/Mexico_City">(GMT-06:00) Central Time - Mexico City</option>
                                                    <option value="America/Regina">(GMT-06:00) Central Time - Regina</option>
                                                    <option value="America/Guatemala">(GMT-06:00) Guatemala</option>
                                                    <option value="America/Bogota">(GMT-05:00) Bogota</option>
                                                    <option value="America/New_York">(GMT-05:00) Eastern Time</option>
                                                    <option value="America/Lima">(GMT-05:00) Lima</option>
                                                    <option value="America/Caracas">(GMT-04:30) Caracas</option>
                                                    <option value="America/Halifax">(GMT-04:00) Atlantic Time - Halifax</option>
                                                    <option value="America/Guyana">(GMT-04:00) Guyana</option>
                                                    <option value="America/La_Paz">(GMT-04:00) La Paz</option>
                                                    <option value="America/Argentina/Buenos_Aires">(GMT-03:00) Buenos Aires</option>
                                                    <option value="America/Godthab">(GMT-03:00) Godthab</option>
                                                    <option value="America/Montevideo">(GMT-03:00) Montevideo</option>
                                                    <option value="America/St_Johns">(GMT-03:30) Newfoundland Time - St. Johns</option>
                                                    <option value="America/Santiago">(GMT-03:00) Santiago</option>
                                                    <option value="America/Sao_Paulo">(GMT-02:00) Sao Paulo</option>
                                                    <option value="Atlantic/South_Georgia">(GMT-02:00) South Georgia</option>
                                                    <option value="Atlantic/Azores">(GMT-01:00) Azores</option>
                                                    <option value="Atlantic/Cape_Verde">(GMT-01:00) Cape Verde</option>
                                                    <option value="Africa/Casablanca">(GMT+00:00) Casablanca</option>
                                                    <option value="Europe/Dublin">(GMT+00:00) Dublin</option>
                                                    <option value="Europe/Lisbon">(GMT+00:00) Lisbon</option>
                                                    <option value="Europe/London">(GMT+00:00) London</option>
                                                    <option value="Africa/Monrovia">(GMT+00:00) Monrovia</option>
                                                    <option value="Africa/Algiers">(GMT+01:00) Algiers</option>
                                                    <option value="Europe/Amsterdam">(GMT+01:00) Amsterdam</option>
                                                    <option value="Europe/Berlin">(GMT+01:00) Berlin</option>
                                                    <option value="Europe/Brussels">(GMT+01:00) Brussels</option>
                                                    <option value="Europe/Budapest">(GMT+01:00) Budapest</option>
                                                    <option value="Europe/Belgrade">(GMT+01:00) Central European Time - Belgrade</option>
                                                    <option value="Europe/Prague">(GMT+01:00) Central European Time - Prague</option>
                                                    <option value="Europe/Copenhagen">(GMT+01:00) Copenhagen</option>
                                                    <option value="Europe/Madrid">(GMT+01:00) Madrid</option>
                                                    <option value="Europe/Paris">(GMT+01:00) Paris</option>
                                                    <option value="Europe/Rome">(GMT+01:00) Rome</option>
                                                    <option value="Europe/Stockholm">(GMT+01:00) Stockholm</option>
                                                    <option value="Europe/Vienna">(GMT+01:00) Vienna</option>
                                                    <option value="Europe/Warsaw">(GMT+01:00) Warsaw</option>
                                                    <option value="Europe/Athens">(GMT+02:00) Athens</option>
                                                    <option value="Europe/Bucharest">(GMT+02:00) Bucharest</option>
                                                    <option value="Africa/Cairo">(GMT+02:00) Cairo</option>
                                                    <option value="Asia/Jerusalem">(GMT+02:00) Jerusalem</option>
                                                    <option value="Africa/Johannesburg">(GMT+02:00) Johannesburg</option>
                                                    <option value="Europe/Helsinki">(GMT+02:00) Helsinki</option>
                                                    <option value="Europe/Kiev">(GMT+02:00) Kiev</option>
                                                    <option value="Europe/Kaliningrad">(GMT+02:00) Moscow-01 - Kaliningrad</option>
                                                    <option value="Europe/Riga">(GMT+02:00) Riga</option>
                                                    <option value="Europe/Sofia">(GMT+02:00) Sofia</option>
                                                    <option value="Europe/Tallinn">(GMT+02:00) Tallinn</option>
                                                    <option value="Europe/Vilnius">(GMT+02:00) Vilnius</option>
                                                    <option value="Europe/Istanbul">(GMT+03:00) Istanbul</option>
                                                    <option value="Asia/Baghdad">(GMT+03:00) Baghdad</option>
                                                    <option value="Africa/Nairobi">(GMT+03:00) Nairobi</option>
                                                    <option value="Europe/Minsk">(GMT+03:00) Minsk</option>
                                                    <option value="Asia/Riyadh">(GMT+03:00) Riyadh</option>
                                                    <option value="Europe/Moscow">(GMT+03:00) Moscow+00 - Moscow</option>
                                                    <option value="Asia/Tehran">(GMT+03:30) Tehran</option>
                                                    <option value="Asia/Baku">(GMT+04:00) Baku</option>
                                                    <option value="Europe/Samara">(GMT+04:00) Moscow+01 - Samara</option>
                                                    <option value="Asia/Tbilisi">(GMT+04:00) Tbilisi</option>
                                                    <option value="Asia/Yerevan">(GMT+04:00) Yerevan</option>
                                                    <option value="Asia/Kabul">(GMT+04:30) Kabul</option>
                                                    <option value="Asia/Karachi">(GMT+05:00) Karachi</option>
                                                    <option value="Asia/Yekaterinburg">(GMT+05:00) Moscow+02 - Yekaterinburg</option>
                                                    <option value="Asia/Tashkent">(GMT+05:00) Tashkent</option>
                                                    <option value="Asia/Colombo">(GMT+05:30) Colombo</option>
                                                    <option value="Asia/Almaty">(GMT+06:00) Almaty</option>
                                                    <option value="Asia/Dhaka">(GMT+06:00) Dhaka</option>
                                                    <option value="Asia/Rangoon">(GMT+06:30) Rangoon</option>
                                                    <option value="Asia/Bangkok">(GMT+07:00) Bangkok</option>
                                                    <option value="Asia/Jakarta">(GMT+07:00) Jakarta</option>
                                                    <option value="Asia/Krasnoyarsk">(GMT+07:00) Moscow+04 - Krasnoyarsk</option>
                                                    <option value="Asia/Shanghai">(GMT+08:00) China Time - Beijing</option>
                                                    <option value="Asia/Hong_Kong">(GMT+08:00) Hong Kong</option>
                                                    <option value="Asia/Kuala_Lumpur">(GMT+08:00) Kuala Lumpur</option>
                                                    <option value="Asia/Irkutsk">(GMT+08:00) Moscow+05 - Irkutsk</option>
                                                    <option value="Asia/Singapore">(GMT+08:00) Singapore</option>
                                                    <option value="Asia/Taipei">(GMT+08:00) Taipei</option>
                                                    <option value="Asia/Ulaanbaatar">(GMT+08:00) Ulaanbaatar</option>
                                                    <option value="Australia/Perth">(GMT+08:00) Western Time - Perth</option>
                                                    <option value="Asia/Yakutsk">(GMT+09:00) Moscow+06 - Yakutsk</option>
                                                    <option value="Asia/Seoul">(GMT+09:00) Seoul</option>
                                                    <option value="Asia/Tokyo">(GMT+09:00) Tokyo</option>
                                                    <option value="Australia/Darwin">(GMT+09:30) Central Time - Darwin</option>
                                                    <option value="Australia/Brisbane">(GMT+10:00) Eastern Time - Brisbane</option>
                                                    <option value="Pacific/Guam">(GMT+10:00) Guam</option>
                                                    <option value="Asia/Magadan">(GMT+10:00) Moscow+07 - Magadan</option>
                                                    <option value="Asia/Vladivostok">(GMT+10:00) Moscow+07 - Yuzhno-Sakhalinsk</option>
                                                    <option value="Pacific/Port_Moresby">(GMT+10:00) Port Moresby</option>
                                                    <option value="Australia/Adelaide">(GMT+10:30) Central Time - Adelaide</option>
                                                    <option value="Australia/Hobart">(GMT+11:00) Eastern Time - Hobart</option>
                                                    <option value="Australia/Sydney">(GMT+11:00) Eastern Time - Melbourne, Sydney</option>
                                                    <option value="Pacific/Guadalcanal">(GMT+11:00) Guadalcanal</option>
                                                    <option value="Pacific/Noumea">(GMT+11:00) Noumea</option>
                                                    <option value="Pacific/Majuro">(GMT+12:00) Majuro</option>
                                                    <option value="Asia/Kamchatka">(GMT+12:00) Moscow+09 - Petropavlovsk-Kamchatskiy</option>
                                                    <option value="Pacific/Auckland">(GMT+13:00) Auckland</option>
                                                    <option value="Pacific/Fakaofo">(GMT+13:00) Fakaofo</option>
                                                    <option value="Pacific/Fiji">(GMT+13:00) Fiji</option>
                                                    <option value="Pacific/Tongatapu">(GMT+13:00) Tongatapu</option>
                                                    <option value="Pacific/Apia">(GMT+14:00) Apia</option>
                                                </Input>
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="maxOutDays" sm={2}>Last Allowed Checkout</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidMaxOutDays ? true : false}
                                                    value={this.state.maxOutDays} type="number" name="maxOutDays" id="maxOutDays" placeholder="Last checkout day" />
                                                {this.state.invalidMaxOutDays &&
                                                    <FormFeedback>{this.state.invalidMaxOutDays}</FormFeedback>}
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="minInDays" sm={2}>First Allowed Checkin</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidMinInDays ? true : false}
                                                    value={this.state.minInDays} type="number" name="minInDays" id="minInDays" placeholder="Last checkout day" />
                                                {this.state.invalidMinInDays &&
                                                    <FormFeedback>{this.state.invalidMinInDays}</FormFeedback>}
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="reservationReminderDaysBefore" sm={2}>Reservation Reminder</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidReservationReminderDaysBefore ? true : false}
                                                    value={this.state.reservationReminderDaysBefore} type="number" name="reservationReminderDaysBefore" id="reservationReminderDaysBefore" placeholder="Last checkout day" />
                                                {this.state.invalidReservationReminderDaysBefore &&
                                                    <FormFeedback>{this.state.invalidReservationReminderDaysBefore}</FormFeedback>}
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="balanceReminderIntervalDays" sm={2}>Balance Reminder</Label>
                                            <Col sm={10}>
                                                <Input onChange={(e) => { this.handleChange(e) }} invalid={this.state.invalidBalanceReminderIntervalDays ? true : false}
                                                    value={this.state.balanceReminderIntervalDays} type="number" name="balanceReminderIntervalDays" id="balanceReminderIntervalDays" placeholder="Last checkout day" />
                                                {this.state.invalidBalanceReminderIntervalDays &&
                                                    <FormFeedback>{this.state.invalidBalanceReminderIntervalDays}</FormFeedback>}
                                            </Col>
                                        </FormGroup>

                                        <FormGroup row>
                                            <Label for="minBalance" sm={2}>Minimum Balance</Label>
                                            <Col sm={10}>
                                                <div id="minBalance">
                                                    <CurrencyInput allowNegative={true} name="minBalance" className="form-control" value={this.state.minBalance} onChangeEvent={this.handleCurrencyChange} />
                                                </div>
                                                {this.state.invalidMinBalance &&
                                                    <div className="invalid-feedback d-block">
                                                        {this.state.invalidMinBalance}
                                                    </div>
                                                }
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

        if (runningState.propertyName) {
            this.setState({ invalidPropertyName: null });
        } else {
            valid = false;
            this.setState({ invalidPropertyName: 'Please enter a property name.' });
        }

        if (runningState.memberRate >= this.props.constraints.memberRateMin / 100 &&
            runningState.memberRate <= this.props.constraints.memberRateMax / 100) {
            this.setState({ invalidMemberRate: null });
        } else {
            valid = false;
            this.setState({ invalidMemberRate: 'Please enter a valid rate.' });
        }

        if (runningState.nonMemberRate > 0.0) {
            this.setState({ invalidNonMemberRate: null });
        } else {
            valid = false;
            this.setState({ invalidNonMemberRate: 'Please enter a non-zero rate.' });
        }

        if (runningState.minBalance < this.props.constraints.minBalanceMin / 100) {
            valid = false;
            this.setState({ invalidMinBalance: 'Please enter a higher valid minimum balance.' });
        } else if (runningState.minBalance > this.props.constraints.minBalanceMax / 100) {
            valid = false;
            this.setState({ invalidMinBalance: 'Please enter a lower valid minimum balance.' });
        } else {
            this.setState({ invalidMinBalance: null });
        }

        return valid;
    }
}

export default inject('appStateStore')(observer(UpdateSettings))

