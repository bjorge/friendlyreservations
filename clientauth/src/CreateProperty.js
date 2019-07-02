import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input,
    Card,
    CardBody,
    CardHeader
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import CurrencyInput from 'react-currency-input';
import { inject, observer } from "mobx-react";
import {
    Redirect,
} from "react-router-dom";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';



// CREATE_PROPERTY_GQL_QUERY, details: https://www.apollographql.com/docs/react/essentials/mutations.html
const CREATE_PROPERTY_GQL_QUERY = gql`
    mutation NewProperty(
    $propertyName: String!,
    $currency: Currency!,
    $memberRate: Int!,
    $nonMemberRate: Int!,
    $allowNonMembers: Boolean!,
    $isMember: Boolean!,
    $nickname: String!,
    $timezone: String!) {
    createProperty(input: {
        propertyName: $propertyName,
        currency: $currency,
        memberRate: $memberRate,
        nonMemberRate: $nonMemberRate,
        allowNonMembers: $allowNonMembers,
        isMember: $isMember,
        nickname: $nickname,
        timezone: $timezone}) {
            propertyId
            eventVersion
            createDateTime
            settings {
              propertyName
            }
            me {
              state
              isAdmin
              isMember
              nickname
              email
              userId    
            }    
        }
}
`;

class CreateProperty extends Component {

    constructor(props) {
        super(props);

        this.toggleAllowNonMembers = this.toggleAllowNonMembers.bind(this);
        this.toggleIsMember = this.toggleIsMember.bind(this);
        this.state = {
            propertyName: '',
            currency: 'USD',
            memberRate: 40.00,
            nonMemberRate: 80.00,
            allowNonMembers: true,
            isMember: true,
            nickname: '',
            timezone: 'America/Los_Angeles'
        };

        this.handleChange = this.handleChange.bind(this);
        this.handleMemberRateCurrencyChange = this.handleMemberRateCurrencyChange.bind(this);
        this.handleNonMemberRateCurrencyChange = this.handleNonMemberRateCurrencyChange.bind(this);

    }

    componentDidMount() {
        if (this.props.item) {
            this.setState({ item: this.props.item })
        }
    }

    toggleAllowNonMembers() {
        this.setState({
            allowNonMembers: !this.state.allowNonMembers
        });
    }

    toggleIsMember() {
        this.setState({
            isMember: !this.state.isMember
        });
    }

    render() {

        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (propertyId !== null) {
            return (<Redirect to="/propertyhome" />)
        } else {
            const apolloClient = this.props.appStateStore.apolloClient;

            return (
                <Mutation client={apolloClient} fetchPolicy='no-cache' mutation={CREATE_PROPERTY_GQL_QUERY} onCompleted={(data) => {
                    console.log("data back from server:");
                    console.log(data);
                    this.props.appStateStore.setPropertyId(data.createProperty.propertyId);
                    this.props.appStateStore.setMe(data.createProperty.me);
                    this.props.appStateStore.setPropertyEventVersion(data.createProperty.eventVersion);
                    if (data.createProperty.me.isMember) {
                        this.props.appStateStore.setPropertyView('MEMBER');
                    } else if (data.createProperty.me.isAdmin) {
                        this.props.appStateStore.setPropertyView('ADMIN');
                    } else {
                        console.log("ERROR! USER IS NOT A MEMBER OR AND ADMIN! ");
                        console.log(data.createProperty.me)
                    }
                }}>
                    {(newPropertySubmit, { loading, error }) => {
                        if (loading) { return (<Spinner />); }
                        if (error) { return (<ErrorModal error={error} />); } 
                        return (
                            <Card key="createProperty">
                                <CardHeader>Fill out and submit the form below to create a new property!</CardHeader>
                                <CardBody>
                                    <Form
                                        onSubmit={e => {
                                            e.preventDefault();
                                            newPropertySubmit({
                                                variables: {
                                                    propertyName: this.state.propertyName,
                                                    currency: this.state.currency,
                                                    memberRate: Math.trunc(this.state.memberRate * 100),
                                                    nonMemberRate: Math.trunc(this.state.nonMemberRate * 100),
                                                    allowNonMembers: this.state.allowNonMembers,
                                                    isMember: this.state.isMember,
                                                    nickname: this.state.nickname,
                                                    timezone: this.state.timezone
                                                }
                                            });
                                        }}
                                    >
                                        <FormGroup>
                                            <Label for="propertyName">Property Name</Label>
                                            {/* <Input type="text" name="name" id="propertyName" placeholder="name of property" /> */}
                                            <Input type="text" name="propertyName" id="propertyName" value={this.state.propertyName}
                                                onChange={this.handleChange} />
                                        </FormGroup>

                                        <FormGroup>
                                            <Label for="timezone">Timezone</Label>
                                            <Input type="select" name="timezone" id="timezone" onChange={this.handleChange}>
                                                <option hidden >Select a timezone</option>
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
                                        </FormGroup>

                                        <FormGroup>
                                            <Label for="nickname">Your Name</Label>
                                            {/* <Input type="text" name="name" id="propertyName" placeholder="name of property" /> */}
                                            <Input type="text" name="nickname" id="nickname" value={this.state.nickname}
                                                onChange={this.handleChange} />
                                        </FormGroup>

                                        {/* <FormGroup check>
                                            <Label check>
                                                <Input onClick={this.toggleIsMember} type="checkbox" name="isMember" checked={this.state.isMember ? 'checked' : ''} onChange={this.handleChange} />
                                                Will you manage and make reservations?
                                                </Label>
                                        </FormGroup> */}

                                        <FormGroup>
                                            <Label for="currency">Currency</Label>
                                            <Input type="select" name="currency" id="currency" onChange={this.handleChange} >
                                                <option value='USD'>$</option>
                                                <option value='EUR'>€</option>
                                            </Input>
                                        </FormGroup>

                                        <FormGroup>
                                            <Label for="memberRate">Daily Member Rate</Label>
                                            <div id="memberRate">
                                                <CurrencyInput className="form-control" value={this.state.memberRate} onChangeEvent={this.handleMemberRateCurrencyChange} />
                                            </div>
                                        </FormGroup>

                                        <FormGroup check>
                                            <Label check>
                                                <Input onClick={this.toggleAllowNonMembers} type="checkbox" name="allowNonMembers" checked={this.state.allowNonMembers ? 'checked' : ''} onChange={this.handleChange} />
                                                Allow Friends of Members?
                                                </Label>
                                        </FormGroup>

                                        {this.state.allowNonMembers && <FormGroup >
                                            <Label for="nonMemberRate">Daily Non Member Rate</Label>
                                            <div id="nonMemberRate">
                                                <CurrencyInput className="form-control" value={this.state.nonMemberRate} onChangeEvent={this.handleNonMemberRateCurrencyChange} />
                                            </div>
                                        </FormGroup>}

                                        <div className="text-center">
                                            <Button>Submit</Button>
                                        </div>
                                    </Form>
                                </CardBody>
                            </Card>
                        )
                    }}
                </Mutation>

            )
        }


    }

    handleChange(event) {
        const target = event.target;
        var tmpValue;
        switch (target.name) {
            case 'allowNonMembers':
            case 'isMember':
                tmpValue = (target.checked ? true : false);
                break;
            case 'currency':
                tmpValue = target.options[target.selectedIndex].value;
                break;
            case 'propertyName':
            case 'nickname':
            case 'timezone':
            default:
                tmpValue = target.value;
                break;
        }
        const value = tmpValue;
        const name = target.name;
        this.setState({
            [name]: value
        });
    }

    handleMemberRateCurrencyChange(event, maskedvalue, floatvalue) {
        var name = 'memberRate';
        this.setState({
            [name]: floatvalue
        });    
    }

    handleNonMemberRateCurrencyChange(event, maskedvalue, floatvalue) {
        var name = 'nonMemberRate';
        this.setState({
            [name]: floatvalue
        }); 
    }

}

export default inject('appStateStore')(observer(CreateProperty))


