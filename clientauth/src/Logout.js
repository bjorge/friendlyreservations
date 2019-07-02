import React, { Component } from 'react';

import { Query } from 'react-apollo';
import gql from 'graphql-tag';

import { Modal, ModalHeader, ModalBody, ModalFooter, Container, Row, Col } from 'reactstrap';

import { inject, observer } from "mobx-react";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

class Logout extends Component {

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
      <Query client={apolloClient} query={Logout.logoutGgqlRequest()} fetchPolicy='no-cache'>
        {({ loading, error, data }) => {
          if (loading)
            return (<Spinner />);
          if (error) {
            return (<ErrorModal error={error} />);
          } else {
            return (
              <Modal isOpen={showModal} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>Logout from Friendly Reservations</ModalHeader>
                <ModalBody>
                  Note that logging out will log you out from your Gmail account.
                </ModalBody>
                <ModalFooter>
                  <Container>
                    <Row>
                      <Col>&nbsp;</Col>
                      <Col><a href={data.logoutURL} className="btn btn-primary" role="button">logout</a></Col>
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

  static logoutGgqlRequest() {
    return gql`
    query HomeLogout {
      logoutURL(dest: "/")
      }
    `;
  }


}

export default inject('appStateStore')(observer(Logout))

