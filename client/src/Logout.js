import React, { Component } from 'react';

import { Modal, ModalHeader } from 'reactstrap';

import { inject, observer } from "mobx-react";

class Logout extends Component {

  constructor(props) {
    super(props);
    this.toggle = this.toggle.bind(this);
  }

  toggle() {
    this.props.exitModal();
  }


  render() {
    const showModal = this.props.showModal;

      
            return (
              <Modal isOpen={showModal} toggle={this.toggle}>
                <ModalHeader toggle={this.toggle}>Logged out from Friendly Reservations</ModalHeader>
              </Modal>
            );
          
    
  }



}

export default inject('appStateStore')(observer(Logout))

