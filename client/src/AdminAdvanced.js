import React, { Component } from 'react';
import {
    Button,
} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import Spinner from './Spinner';
import Export from './Export';
import ExportCSV from './ExportCSV';
import DeleteProperty from './DeleteProperty';
import ErrorModal from './ErrorModal';

import {
    Redirect
} from "react-router-dom";

import { inject, observer } from "mobx-react";

const GET_SETTINGS_GQL = gql`
query PropertySettings(
  $propertyId: String!) {
    property(id: $propertyId) {
        eventVersion
        updateSettingsConstraints {
            allowPropertyDelete
        }
    }
}
`;

var buttonStyle = {
    margin: '0',
    padding: '0',
};

class AdminAdvanced extends Component {

    constructor(props) {
        super(props);
        this.turnOffModals = this.turnOffModals.bind(this);
        this.displayExportModal = this.displayExportModal.bind(this);
        this.displayExportCSVModal = this.displayExportCSVModal.bind(this);
        this.displayDeleteModal = this.displayDeleteModal.bind(this);

        this.state = {
            cachedProperty: null,
            showExportModal: false,
            showExportCSVModal: false,
            showDeleteModal: false,
        };
    }

    render() {
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (propertyId === null) {
            return (<Redirect to="/propertyselect" />)
        } else {

            var info = { propertyId: propertyId };
            const apolloClient = this.props.appStateStore.apolloClient;
            const queryKey = this.props.appStateStore.apolloQueryKey;

            return (

                <Query client={apolloClient} key={queryKey} query={GET_SETTINGS_GQL} fetchPolicy='no-cache'
                    variables={info}
                    onCompleted={(data) => {
                        if (data.property !== undefined) {
                            this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
                            this.setState({ cachedProperty: data.property });
                        }
                    }}>
                    {({ loading, error, data }) => {
                        if (loading) { return (<Spinner />); }

                        var property = this.state.cachedProperty;
                        if (data.property !== undefined) {
                            property = data.property;
                        }

                        if (property === undefined) {
                            return (<div>No data from service, please refresh and try again</div>)
                        }
                        return (

                            <div>
                                {error && <ErrorModal error={error} />}
                                <Export showModal={this.state.showExportModal} exitModal={this.turnOffModals} />
                                <Button color="link" style={buttonStyle} onClick={() => this.displayExportModal()}>Export Backup</Button>
                                <br />
                                <ExportCSV showModal={this.state.showExportCSVModal} exitModal={this.turnOffModals} />
                                <Button color="link" style={buttonStyle} onClick={() => this.displayExportCSVModal()}>Export CSV</Button>
                                <br />
                                <DeleteProperty showModal={this.state.showDeleteModal} exitModal={this.turnOffModals} />
                                {property.updateSettingsConstraints.allowPropertyDelete && <Button color="link" style={buttonStyle} onClick={() => this.displayDeleteModal()}>Delete Property</Button>}
                            </div>
                        )
                    }}
                </Query>)
        }
    }


    turnOffModals = () => {
        this.setState({ showExportModal: false });
        this.setState({ showExportCSVModal: false });
        this.setState({ showDeleteModal: false });
    }

    displayExportModal() {
        this.setState({
            showExportModal: true
        });
    }

    displayExportCSVModal() {
        this.setState({
            showExportCSVModal: true
        });
    }

    displayDeleteModal() {
        this.setState({
            showDeleteModal: true
        });
    }
}

export default inject('appStateStore')(observer(AdminAdvanced))
