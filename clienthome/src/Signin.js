import React, { Component } from 'react';

import { Modal, ModalHeader, ModalBody, ModalFooter, Container, Row, Col } from 'reactstrap';

class Signin extends Component {

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
                <ModalHeader toggle={this.toggle}>Signin to Friendly Reservations!</ModalHeader>
                <ModalBody>
                    Note that your Gmail account will be used to signin.
                </ModalBody>
                <ModalFooter>
                    <Container>
                        <Row>
                            <Col>&nbsp;</Col>
                            <Col><a href="/fr/" className="btn btn-primary" role="button">Signin</a></Col>
                            <Col>&nbsp;</Col>
                        </Row>
                    </Container>
                </ModalFooter>
            </Modal>
        )
    }


}

export default Signin

