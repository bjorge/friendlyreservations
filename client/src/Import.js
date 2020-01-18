import React, { Component } from 'react';

import { Mutation } from 'react-apollo';
import gql from 'graphql-tag';

import { Modal, ModalHeader, ModalBody, Form, Button } from 'reactstrap';

import { inject, observer } from "mobx-react";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

const IMPORT_GQL_MUTATION = gql`
mutation ImportProperty {
        importProperty
  }
`

class Import extends Component {

    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
    }

    toggle() {
        this.props.exitModal();
    }

    render() {
        const apolloClient = this.props.appStateStore.apolloClient;
        const showModal = this.props.showModal;

        return (
            <Modal isOpen={showModal} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>Import Property Data</ModalHeader>
                <ModalBody>
                    To import click the import button and then check the list of properties.
                    Note that you must be a member or admin of the imported property.
                    <hr />
                    <Mutation client={apolloClient} mutation={IMPORT_GQL_MUTATION} fetchPolicy='no-cache'

                        onCompleted={(data) => {
                            this.toggle();
                        }}>
                        {(onSubmit, { loading, error }) => {
                            if (loading) return (<Spinner />);
                            return (
                                <div>
                                    {error && <ErrorModal error={error} />}
                                    <Form onSubmit={event => {
                                        event.preventDefault();

                                        onSubmit();

                                    }}
                                    >
                                        <div className="text-center">
                                            <Button color="primary" type="submit">Import</Button>
                                        </div>
                                    </Form>
                                </div>
                            );
                        }}
                    </Mutation>
                </ModalBody>
            </Modal>
        );
    }
}


export default inject('appStateStore')(observer(Import))

