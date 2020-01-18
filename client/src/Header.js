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

import Logout from './Logout';

// use the NavLink from router 4 instead of from reactstrap
// since it will use components
import { NavLink } from "react-router-dom";

import { inject, observer } from "mobx-react";

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faBars } from '@fortawesome/free-solid-svg-icons'
import { faTimes } from '@fortawesome/free-solid-svg-icons'

import Signin from './Signin';


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

        this.displayLogoutModal = this.displayLogoutModal.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);

        this.setAdminView = this.setAdminView.bind(this);
        this.setMemberView = this.setMemberView.bind(this);

        this.state = {
            isOpen: false,
            showLogoutModal: false,
            showSigninModal: false,
        };

        this.displaySigninModal = this.displaySigninModal.bind(this);

    }
    toggle() {
        this.setState({
            isOpen: !this.state.isOpen
        });
    }

    displaySigninModal() {
        this.setState({
            showSigninModal: true
        });
    }

    displayLogoutModal() {
        console.log("Header: logout called, close navbar");
        this.closeNavbar();
        this.props.appStateStore.clearAll();
        this.props.logoutCallback();
    }

    turnOffModals = () => {
        this.setState({ showLogoutModal: false });
        this.setState({ showSigninModal: false });
        this.closeNavbar();
    }

    closeNavbar() {
        // if (this.state.isOpen === true) {
        this.toggle();
        // }
    }

    setAdminView = () => {
        this.closeNavbar();
        this.props.appStateStore.setPropertyView('ADMIN');
    }

    setMemberView = () => {
        this.closeNavbar();
        this.props.appStateStore.setPropertyView('MEMBER');
    }

    render() {
        //const authenticated = this.props.appStateStore.authenticated;
        const memberView = this.props.appStateStore.propertyView === 'MEMBER';
        const adminView = this.props.appStateStore.propertyView === 'ADMIN';
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;
        const isAdmin = this.props.appStateStore.me ? this.props.appStateStore.me.isAdmin : false;
        const isMember = this.props.appStateStore.me ? this.props.appStateStore.me.isMember : false;
        const authenticated = this.props.authenticated;

        authenticated ? console.log("Header: authenticated true") : console.log("Header: authenticated false");
        memberView ? console.log("Header: member view true") : console.log("Header: member view false");
        adminView ? console.log("Header: admin view true") : console.log("Header: admin view false");
        propertyId ? console.log("Header:  id set") : console.log("Header: property id not set");
        return (
            <Navbar className="bg-dark navbar-dark" light >
                <NavbarToggler onClick={this.toggle} >
                    {!this.state.isOpen && <span><FontAwesomeIcon icon={faBars} /></span>}
                    {this.state.isOpen && <span><FontAwesomeIcon icon={faTimes} /></span>}
                </NavbarToggler>
                <NavbarBrand href="/" className="mx-auto">Friendly Reservations</NavbarBrand>
                <Collapse isOpen={this.state.isOpen} navbar>
                    <Nav className="ml-auto" navbar>
                        {!authenticated &&
                            <div>
                                <NavItem>
                                    <NavLink exact to='/splashhome' onClick={this.closeNavbar}>Home</NavLink>
                                </NavItem>
                                <NavItem>
                                    <Signin showModal={this.state.showSigninModal} exitModal={this.turnOffModals} />
                                    <Button color="link" style={buttonStyle} onClick={() => this.displaySigninModal()}>Signin</Button>
                                </NavItem>
                                <NavItem>
                                    <NavLink exact to='/about' onClick={this.closeNavbar}>About</NavLink>
                                </NavItem>
                            </div>
                        }
                        {authenticated && propertyId === null &&
                            <div>
                                <NavItem>
                                    <NavLink exact to='/propertyselect' onClick={this.closeNavbar}>Select Property</NavLink>
                                </NavItem>
                                {/* <NavItem>
                                    <NavLink exact to='/createproperty' onClick={this.closeNavbar}>Create Property</NavLink>
                                </NavItem> */}
                                <NavItem>
                                    <NavLink exact to='/about' onClick={this.closeNavbar}>About</NavLink>
                                </NavItem>
                                <NavItem>
                                    <Logout showModal={this.state.showLogoutModal} exitModal={this.turnOffModals} />
                                    <Button color="link" style={buttonStyle} onClick={() => this.displayLogoutModal()}>Logout</Button>
                                </NavItem>
                            </div>
                        }
                        {authenticated && propertyId !== null &&
                            <div>
                                <NavItem>
                                    <NavLink exact to='/propertyhome' onClick={this.closeNavbar}>Home</NavLink>
                                </NavItem>
                                {memberView &&
                                    <div>
                                        <NavItem>
                                            <NavLink exact to='/reservations' onClick={this.closeNavbar}>Reservations</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/ledger' onClick={this.closeNavbar}>Ledger</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/notifications' onClick={this.closeNavbar}>Notifications</NavLink>
                                        </NavItem>
                                        {/* <NavItem>
                                            <NotificationsView propertyId={propertyId} userId={userId} showModal={this.state.showNotificationsModal} exitModal={this.turnOffModals} />
                                            <Button color="link" style={buttonStyle} onClick={() => this.displayNotificationsModal()}>Notifications</Button>
                                        </NavItem> */}

                                        {isAdmin &&
                                            <NavItem>
                                                <NavLink exact to='/propertyhome' onClick={() => this.setAdminView()}>Admin View</NavLink>
                                            </NavItem>}
                                        <hr />
                                    </div>
                                }
                                {adminView &&
                                    <div>
                                        {/* <NavItem> 
                                            <MemberRestrictionsView admin={true} propertyId={propertyId} userId={userId} showModal={this.state.showRestrictionsModal} exitModal={this.turnOffModals} />
                                            <Button color="link" style={buttonStyle} onClick={() => this.displayRestrictionsModal()}>Restrictions</Button>
                                        </NavItem> */}
                                        <NavItem>
                                            <NavLink exact to='/restrictions' onClick={this.closeNavbar}>Restrictions</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/users' onClick={this.closeNavbar}>Users</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/adminreservations' onClick={this.closeNavbar}>Reservations</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/settings' onClick={this.closeNavbar}>Settings</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/contents' onClick={this.closeNavbar}>Homepages</NavLink>
                                        </NavItem>
                                        <NavItem>
                                            <NavLink exact to='/adminadvanced' onClick={this.closeNavbar}>Advanced</NavLink>
                                        </NavItem>
                                        {isMember &&
                                            <NavItem>
                                                <NavLink exact to='/propertyhome' onClick={() => this.setMemberView()}>Member View</NavLink>
                                            </NavItem>}
                                        <hr />
                                    </div>
                                }
                                <NavItem>
                                    <NavLink exact to='/exitproperty' onClick={this.closeNavbar}>Exit Property</NavLink>
                                </NavItem>
                            </div>
                        }
                    </Nav>
                </Collapse>
            </Navbar >
        );
    }
}

export default inject('appStateStore')(observer(Header))
