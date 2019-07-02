import React, { Component } from 'react';
import {
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Modal, ModalHeader, ModalBody, ModalFooter 
} from 'reactstrap';
import gql from "graphql-tag";
import { Mutation } from "react-apollo";
import { inject, observer } from "mobx-react";
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

import AppStateStore from './AppStateStore';

import 'bootstrap/dist/css/bootstrap.css';

const CREATE_USER_GQL_MUTATION = gql`
mutation NewUser(
    $propertyId: String!,
    $input: NewUserInput!) {
        createUser(
            propertyId: $propertyId, 
            input: $input) {
            ${AppStateStore.propertyGqlResultsString()}
    }
}
`;

class User extends Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.toggle = this.toggle.bind(this);

        this.state = {
            emailInValid: null,
            nicknameInValid: null,
            modal: false,
        };
    }

    toggle() {
        this.setState({
          modal: false
        });
      }

    render() {
        const showView = this.props.viewinfo ? true : false;
        const nickname = showView ? this.props.viewinfo.nickname : 0;
        const isAdmin = showView ? this.props.viewinfo.isAdmin : "";
        console.log("User component");
        if (showView) {
            console.log(this.props.viewinfo);
        }

        if (showView) {
            // Form to view/remove an existing restriction
            return (
                <div>
                    nickname: {nickname}<br />
                    isAdmin: {isAdmin ? 'true' : 'false'}
                </div>
            );
        } else {
            // Form to create a new user
            return (
                <div>done</div>
            );
        }
    }

    handleChange(event) {
        const target = event.target;
        switch (target.name) {
            case 'email':
                this.setState({ emailInValid: null })
                break;
            case 'nickname':
                this.setState({ nicknameInValid: null })
                break;
            default:
                break;
        }
    }
}

export default inject('appStateStore')(observer(User))

