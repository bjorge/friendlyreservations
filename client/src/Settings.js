import React, { Component } from 'react';
import {
    Table,
    Button,
} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import Spinner from './Spinner';
import UpdateSettings from './UpdateSettings';
import ErrorModal from './ErrorModal';

import {
    Redirect
} from "react-router-dom";

import { inject, observer } from "mobx-react";

const GET_SETTINGS_GQL = gql`
query Ledgers(
  $propertyId: String!) {
  property(id: $propertyId) {
    propertyId
    eventVersion

    settings {
        propertyName
        currencyAcronym: currency(format: ACRONYM)
        currencySymbol: currency(format: SYMBOL)
        memberRate
        allowNonMembers
        nonMemberRate
        timezone
        maxOutDays
        minInDays
        reservationReminderDaysBefore
        balanceReminderIntervalDays
        minBalance
      }

      updateSettingsConstraints {
        propertyNameMin
        propertyNameMax
        memberRateMin
        memberRateMax
        nonMemberRateMin
        nonMemberRateMax
        minBalanceMin
        minBalanceMax
        maxOutDaysMin
        maxOutDaysMax
        minInDaysMin
        minInDaysMax
        reservationReminderDaysBeforeMin
        reservationReminderDaysBeforeMax
        balanceReminderIntervalDaysMin
        balanceReminderIntervalDaysMax
        
      }
}
}
`;

class Settings extends Component {

    constructor(props) {
        super(props);
        this.turnOffModals = this.turnOffModals.bind(this);
        this.displaySettingsForm = this.displaySettingsForm.bind(this);

        this.state = {
            cachedProperty: null,
            showSettingsForm: false,
        };
    }

    render() {
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;
        var property = this.state.cachedProperty

        if (propertyId === null) {
            return (<Redirect to="/propertyselect" />)
        } else {

            var info = { propertyId: propertyId };
            const apolloClient = this.props.appStateStore.apolloClient;
            const queryKey = this.props.appStateStore.apolloQueryKey;

            return (

                <Query client={apolloClient} key={queryKey} query={GET_SETTINGS_GQL} fetchPolicy='no-cache'
                    variables={info}>
                    {({ loading, error, data }) => {
                        if (loading) { return (<Spinner />); }
                        if (error) { return (<ErrorModal error={error} />); }

                        if (data) {
                            if (data.property !== undefined) {
                                property = data.property
                            }
                            if (property === null) {
                                // cache property is also null
                                return (<div>No data from service, please refresh and try again</div>)
                            }

                        }
                        return (

                            <div>
                                {error && <ErrorModal error={error} />}
                                <Table bordered size="sm">
                                    <thead>
                                        <tr>
                                            <th>Name</th>
                                            <th>Value</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr>
                                            <th scope="row">{"Property Name"}</th>
                                            <td>{property.settings.propertyName}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Currency"}</th>
                                            <td>{property.settings.currencyAcronym} ({property.settings.currencySymbol})</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Member Rate"}</th>
                                            <td>{property.settings.memberRate}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Allow Non-Members"}</th>
                                            <td>{property.settings.allowNonMembers ? 'true' : 'false'}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Non-Member Rate"}</th>
                                            <td>{property.settings.nonMemberRate}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Timezone"}</th>
                                            <td>{property.settings.timezone}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Last Checkout"}</th>
                                            <td>{property.settings.maxOutDays}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"First Checkin"}</th>
                                            <td>{property.settings.minInDays}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Reservation Reminder"}</th>
                                            <td>{property.settings.reservationReminderDaysBefore}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Low Balance Reminder"}</th>
                                            <td>{property.settings.balanceReminderIntervalDays}</td>
                                        </tr>
                                        <tr>
                                            <th scope="row">{"Minimum Balance"}</th>
                                            <td>{property.settings.minBalance}</td>
                                        </tr>
                                    </tbody>
                                </Table>
                                <UpdateSettings currentSettings={property.settings} constraints={property.updateSettingsConstraints} showform={this.state.showSettingsForm} exitModal={this.turnOffModals} />
                                <div className="text-center">
                                    <Button color="primary" onClick={() => this.displaySettingsForm()}>Update Settings</Button>
                                </div>
                            </div>
                        )
                    }}
                </Query>)
        }
    }


    turnOffModals = () => {
        this.setState({ showSettingsForm: false });
    }

    displaySettingsForm() {
        this.setState({
            showSettingsForm: true
        });
    }
}

export default inject('appStateStore')(observer(Settings))
