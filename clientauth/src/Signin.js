import React, { Component } from 'react';

import { Query } from 'react-apollo';
import gql from 'graphql-tag';

import { Modal, ModalHeader, ModalBody, ModalFooter, Container, Row, Col } from 'reactstrap';

import { inject, observer } from "mobx-react";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

class Signin extends Component {

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
            <Query client={apolloClient} query={Signin.loginGgqlRequest()} fetchPolicy='no-cache'>
                {({ loading, error, data }) => {
                    if (loading)
                        return (<Spinner />);
                    if (error) {
                        return (<ErrorModal error={error} />);
                    } else {
                        return (
                            <Modal isOpen={showModal} toggle={this.toggle}>
                                <ModalHeader toggle={this.toggle}>Signin to Friendly Reservations!</ModalHeader>
                                <ModalBody>
                                    Note that your Gmail account will be used to signin.
                </ModalBody>
                                <ModalFooter>
                                    <Container>
                                        <Row>
                                            <Col>&nbsp;</Col>
                                            <Col><a href={data.loginURL} className="btn btn-primary" role="button">Signin</a></Col>
                                            <Col>&nbsp;</Col>
                                        </Row>
                                    </Container>
                                </ModalFooter>
                            </Modal>
                        );
                    }
                }}
            </Query>

        )
    }

    static loginGgqlRequest() {
        return gql`
        query HomeLogin {
          loginURL(dest: "/")
          }
        `;
    }
}

export default inject('appStateStore')(observer(Signin))


