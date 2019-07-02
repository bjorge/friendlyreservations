import React, { Component } from 'react';
import {
    Collapse,
    Navbar,
    NavbarToggler,
    NavbarBrand,
    Nav,
    NavItem,
    Button
} from 'reactstrap';

import Signin from './Signin';

// use the NavLink from router 4 instead of from reactstrap
// since it will use components
import { NavLink } from "react-router-dom";
// import "./hamburger.css";

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faBars } from '@fortawesome/free-solid-svg-icons'
import { faTimes } from '@fortawesome/free-solid-svg-icons'

// make the button link line up with the navlinks
var buttonStyle = {
    margin: '0',
    padding: '0',
};


class Header extends Component {
    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.closeNavbar = this.closeNavbar.bind(this);
        this.displaySigninModal = this.displaySigninModal.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);
        this.state = {
            isOpen: false,
            showSigninModal: false
        };
    }
    toggle() {
        this.setState({
            isOpen: !this.state.isOpen
        });
    }

    closeNavbar() {
        if (this.state.isOpen === true) {
            this.toggle();
        }
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
            <Navbar className="bg-dark navbar-dark" light >
                <NavbarToggler onClick={this.toggle} >
                    {!this.state.isOpen && <span><FontAwesomeIcon icon={faBars} /></span>}
                    {this.state.isOpen && <span><FontAwesomeIcon icon={faTimes} /></span>}
                </NavbarToggler>
                <NavbarBrand href="/" className="mx-auto">Friendly Reservations</NavbarBrand>
                <Collapse isOpen={this.state.isOpen} navbar>
                    <Nav className="ml-auto" navbar>
                        <NavItem>
                            <NavLink exact to='/home' onClick={this.closeNavbar}>Home</NavLink>
                        </NavItem>
                        <NavItem>
                            <Signin showModal={this.state.showSigninModal} exitModal={this.turnOffModals} />
                            <Button color="link" style={buttonStyle} onClick={() => this.displaySigninModal()}>Signin</Button>
                        </NavItem>
                        <NavItem>
                            <NavLink exact to='/about' onClick={this.closeNavbar}>About</NavLink>
                        </NavItem>
                    </Nav>
                </Collapse>
            </Navbar>
        );
    }
}

export default Header;
