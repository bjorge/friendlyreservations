import React, { Component } from 'react';

// import { graphql } from 'react-apollo';
import gql from 'graphql-tag';
import { Query } from 'react-apollo';
import ErrorModal from './ErrorModal';
import CreatePropertyModal from './CreatePropertyModal';
import Import from './Import';

import {
  Card,
  CardHeader,
  Container,
  Button,
  UncontrolledAlert
} from 'reactstrap';

import Logout from './Logout';

import { inject, observer } from "mobx-react";
import Spinner from './Spinner';

import {
  Redirect,
} from "react-router-dom";

// make the button link looks like other links
var buttonStyle = {
  padding: '0',
  verticalAlign: 'baseline'
};

const GET_PROPERTIES = gql`
{
  properties {
    propertyId
    eventVersion
    settings {
      propertyName
    }
    me {
      state
      isAdmin
      isMember
      nickname
      userId
      email
    }
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
    allowNewProperty
    allowPropertyImport
    trialOn
    trialDays
  }
  updateUserConstraints {
    nicknameMin
    nicknameMax
    invalidNicknames
    invalidEmails
  }
}
`;

class PropertySelect extends Component {
  constructor(props) {
    console.log("PropertySelect: process.env.NODE_ENV is: " + process.env.NODE_ENV);
    super(props);
    this.selectProperty = this.selectProperty.bind(this);
    this.displayLogoutModal = this.displayLogoutModal.bind(this);
    this.displayCreatePropertyModal = this.displayCreatePropertyModal.bind(this);
    this.displayImportModal = this.displayImportModal.bind(this);
    this.turnOffModals = this.turnOffModals.bind(this);
    this.state = {
      cachedProperties: null,
      cachedUpdateSettingsConstraints: null,
      cachedUpdateUserConstraints: null,
      showLogoutModal: false,
      showCreatePropertyModal: false,
      showImportModal: false,
      queryKey: 0,
    };
  }

  displayLogoutModal() {
    this.setState({
      showLogoutModal: true
    });
  }

  displayCreatePropertyModal() {
    this.setState({
      showCreatePropertyModal: true
    });
  }

  displayImportModal() {
    this.setState({
      showImportModal: true
    });
  }

  turnOffModals = () => {
    this.setState({ showLogoutModal: false });
    this.setState({ showCreatePropertyModal: false });
    this.setState({ queryKey: this.state.queryKey + 1 });
    this.setState({ showImportModal: false });
    // console.log("turn off modals");
  }

  selectProperty(property, e) {
    this.props.appStateStore.setPropertyId(property.propertyId);
    this.props.appStateStore.setPropertyEventVersion(property.eventVersion);
    this.props.appStateStore.setMe(property.me);
    //this.props.appStateStore.setProperty(property);
    if (property.me.isMember) {
      this.props.appStateStore.setPropertyView('MEMBER');
    } else if (property.me.isAdmin) {
      this.props.appStateStore.setPropertyView('ADMIN');
    } else {
      console.log("ERROR! USER IS NOT A MEMBER OR AND ADMIN! ");
      console.log(property.me)
    }
  }

  render() {
    const apolloClient = this.props.appStateStore.apolloHomeClient;
    const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

    if (propertyId !== null) {
      return (<Redirect to="/propertyhome" />)
    } else {
      return (<Query key={this.state.queryKey} client={apolloClient} query={GET_PROPERTIES} fetchPolicy='no-cache'
        onCompleted={(data) => {
          if (data.properties !== undefined) {
            this.setState({ cachedProperties: data.properties });
          }
          if (data.updateSettingsConstraints !== undefined) {
            this.setState({ cachedUpdateSettingsConstraints: data.updateSettingsConstraints });
          }
          if (data.updateUserConstraints !== undefined) {
            this.setState({ cachedUpdateUserConstraints: data.updateUserConstraints });
          }
        }}

      >
        {({ loading, error, data }) => {
          if (loading) return (<Spinner />);
          if (error) { return (<ErrorModal error={error} />); }
          if (data) {

            var properties = this.state.cachedProperties;
            if (data.properties !== undefined) {
              properties = data.properties;
            }

            var updateSettingsConstraints = this.state.cachedUpdateSettingsConstraints;
            if (data.updateSettingsConstraints !== undefined) {
              updateSettingsConstraints = data.updateSettingsConstraints;
            }

            var updateUserConstraints = this.state.cachedUpdateUserConstraints;
            if (data.updateUserConstraints !== undefined) {
              updateUserConstraints = data.updateUserConstraints;
            }

          }

          if (properties == null || Object.keys(properties).length === 0) {
            return (
              <Container>
                <Logout showModal={this.state.showLogoutModal} exitModal={this.turnOffModals} />
                <CreatePropertyModal userConstraints={updateUserConstraints} settingsConstraints={updateSettingsConstraints} showForm={this.state.showCreatePropertyModal} exitModal={this.turnOffModals} />
                {updateSettingsConstraints.allowNewProperty === true && <Card key="createProperty">
                  <CardHeader>It looks like you are not a member of any properties.
                        Click <Button style={buttonStyle} color="link" onClick={() => this.displayCreatePropertyModal()}>here</Button> to create a new property. 
                        Or if you logged in as the wrong user, then logout and login again as the correct user.
                  </CardHeader>
                </Card>}
                {updateSettingsConstraints.allowNewProperty === false && <Card key="importProperty">
                  <CardHeader>It looks like you are not a member of any properties.
                        If you logged in as the wrong user, then logout and login again as the correct user.
                  </CardHeader>
                </Card>}


                <Import showModal={this.state.showImportModal} exitModal={this.turnOffModals} />
                {updateSettingsConstraints.allowPropertyImport && <div className="text-center"><hr />
                  <Button color="primary" onClick={() => this.displayImportModal()}>Import</Button>
                </div>}

              </Container>
            );
          }
          return (
            <Container>
              {updateSettingsConstraints.trialOn &&
              <UncontrolledAlert color="info">
                Trial edition, properties are automatically deleted after {updateSettingsConstraints.trialDays} days
              </UncontrolledAlert>}
              <Card key="selectProperty">
                <CardHeader>Click on a property below to access the reservation system!</CardHeader>
              </Card>
              {properties && properties.map(property => {
                return (
                  <Card key={property.propertyId}>
                    <Button className="text-left" onClick={this.selectProperty.bind(this, property)}>{property.settings.propertyName}</Button>
                  </Card>
                )
              })}

              <CreatePropertyModal userConstraints={updateUserConstraints} settingsConstraints={updateSettingsConstraints} showForm={this.state.showCreatePropertyModal} exitModal={this.turnOffModals} />
              {updateSettingsConstraints.allowNewProperty && <div className="text-center"><hr />
                <Button color="primary" onClick={() => this.displayCreatePropertyModal()}>New Property</Button>
              </div>}

              <Import showModal={this.state.showImportModal} exitModal={this.turnOffModals} />
              {updateSettingsConstraints.allowPropertyImport && <div className="text-center"><hr />
                <Button color="primary" onClick={() => this.displayImportModal()}>Import</Button>
              </div>}

            </Container>
          )
        }}
      </Query>)
    }
  }
}

export default inject('appStateStore')(observer(PropertySelect))

