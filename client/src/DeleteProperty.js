import React, { Component } from 'react';

import { Mutation } from 'react-apollo';
import gql from 'graphql-tag';

import { Modal, ModalHeader, ModalBody, Form, Button } from 'reactstrap';

import { inject, observer } from "mobx-react";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

const EXPORT_GQL_MUTATION = gql`
mutation DeleteProperty(
    $propertyId: String!) {
        deleteProperty(propertyId: $propertyId)
  }
`

class DeleteProperty extends Component {

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
                <ModalHeader toggle={this.toggle}>Delete Property</ModalHeader>
                <ModalBody>
                    To delete this property click the delete button.
                    <hr/>
                    <Mutation client={apolloClient} mutation={EXPORT_GQL_MUTATION} fetchPolicy='no-cache'

                        onCompleted={(data) => {
                            this.props.appStateStore.clearAll()
                            this.toggle();
                        }}>
                        {(onSubmit, { loading, error }) => {
                            if (loading) return (<Spinner />);
                            return (
                                <Form onSubmit={event => {

                                    event.preventDefault();

                                    // ok, we can submit! let's setup a cool gql mutation
                                    var info = {
                                        propertyId: this.props.appStateStore.propertyId
                                    }

                                    onSubmit({
                                        variables: info
                                    });

                                }}
                                >
                                {error && <ErrorModal error={error} />}
                                    <div className="text-center">
                                        <Button color="primary" type="submit">Delete</Button>
                                    </div>
                                </Form>
                            );
                        }}
                    </Mutation>
                </ModalBody>
            </Modal>
        );
    }
}


export default inject('appStateStore')(observer(DeleteProperty))

