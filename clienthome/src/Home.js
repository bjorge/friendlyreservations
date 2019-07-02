import React, { Component } from 'react';

import {
    Card,
    CardTitle,
    CardText,
    Container,
    Button
  } from 'reactstrap';

import Signin from './Signin';

// make the button link looks like other links
var buttonStyle = {
    padding: '0',
    verticalAlign: 'baseline'
  };

export default class Home extends Component {
    constructor(props) {
        super(props);
        this.displaySigninModal = this.displaySigninModal.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);
        this.state = {
            showSigninModal: false
        };
    }
    displaySigninModal() {
        this.setState({
            showSigninModal: true
        });
    }

    turnOffModals = () => {
        this.setState({ showSigninModal: false });
    }

    render() {
        return (
            <Container>
            <Signin showModal={this.state.showSigninModal} exitModal={this.turnOffModals} />
            <Card key="createProperty">
              <CardTitle>
              Welcome to Friendly Reservations!
            </CardTitle>
            <CardText>
              Click <Button style={buttonStyle} color="link" onClick={() => this.displaySigninModal()}>here</Button> to signin
                with your Gmail account.
            </CardText>
            </Card>
          </Container>        
        );
    }
}