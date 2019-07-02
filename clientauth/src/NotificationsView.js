import React, { Component } from 'react';
import {
    Collapse,
    Card,
    CardBody,
    Container,
    Button,
    CardText,
} from 'reactstrap';
import gql from "graphql-tag";
import { Query } from "react-apollo";

import Spinner from './Spinner';
import ErrorModal from './ErrorModal';

import {
    Redirect
} from "react-router-dom";

import { inject, observer } from "mobx-react";

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faMinus } from '@fortawesome/free-solid-svg-icons'

const GET_NOTIFICATIONS_GQL = gql`
query DisabledRanges(
    $propertyId: String!,
    $userId: String!) {
    property(id: $propertyId) {
        eventVersion
        notifications(userId: $userId, reverse: true) {
            to {
                nickname
                email
            }
            cc {
                nickname
                email
            }
            subject
            body
            notificationId
            createDateTime
            author {
                nickname
                email
            }
        }
    }
}
`
class NotificationsView extends Component {
    constructor(props) {
        super(props);
        // this.toggleModal = this.toggleModal.bind(this);
        this.toggle = this.toggle.bind(this);

        this.state = {
            collapse: null,
        };
    }

    // toggleModal() {
    //     this.props.exitModal();
    // }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }

    static formatDate(dateTime) {
        var date = new Date(dateTime);
        return date.toLocaleString();
    }

    static formatBody(body) {
        return body.replace(/\n/g, "<br />");
    }

    static formatTitle(subject) {
        var title = subject;
        if (subject.length > 30) {
            title = subject.substring(0, 30) + "...";
        }
        return title;
    }

    render() {

        const { collapse } = this.state;
        var me = this;

        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (propertyId === null) {
            return (<Redirect to="/propertyselect" />)
        } else {
            const apolloClient = this.props.appStateStore.apolloClient;
            const queryKey = this.props.appStateStore.apolloQueryKey;
            const info = { propertyId: propertyId, userId: this.props.appStateStore.me.userId };

            return (

                <Query key={queryKey} query={GET_NOTIFICATIONS_GQL} fetchPolicy='no-cache'
                    client={apolloClient}
                    variables={info}
                    onCompleted={(data) => {
                        // console.log("Notifications query completed");
                        if (data.property !== undefined) {
                            this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
                            this.setState({ cachedProperty: data.property });
                        }
                    }}>
                    {({ loading, error, data }) => {
                        if (loading) { return (<Spinner />); }
                        if (data) {
                            //console.log("notifications data:")
                            //console.log(data)
                            var notifications = {}

                            var property = this.state.cachedProperty;
                            if (data.property !== undefined) {
                                property = data.property;
                            }

                            if (property === undefined) {
                                return (<div>No data from service, please refresh and try again</div>)
                            }
                            notifications = property.notifications

                        }
                        return (
                            <Container>
                                {error && <ErrorModal error={error} />}
                                {notifications.map(function (notification) {
                                    return (
                                        <Card key={notification.notificationId}>
                                            <Button className="text-left" onClick={() => me.toggle(notification.notificationId)}>
                                                {collapse === notification.notificationId ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                                {' '}
                                                {NotificationsView.formatDate(notification.createDateTime)}
                                            </Button>
                                            <Collapse isOpen={collapse === notification.notificationId}>
                                                <CardBody>
                                                    <CardText>From: {notification.author.nickname}{' <'}{notification.author.email}{'>'}</CardText>
                                                    <CardText>Date: {NotificationsView.formatDate(notification.createDateTime)}</CardText>
                                                    <CardText>Subject: {notification.subject}</CardText>
                                                    <CardText>To:{' '}
                                                        {notification.to.map(function (item, index) {
                                                            var text = (index ? ', ' : '') + item.nickname + " <" + item.email + ">";
                                                            // console.log("item: " + item.nickname + " index: " + index + " text: " + text);
                                                            return (<span key={"to_" + index}>{text}</span>)
                                                        })}
                                                    </CardText>
                                                    <CardText>Cc:{' '}
                                                        {notification.cc.map(function (item, index) {
                                                            var text = (index ? ', ' : '') + item.nickname + " <" + item.email + ">";
                                                            // console.log("item: " + item.nickname + " index: " + index + " text: " + text);
                                                            return (<span key={"cc_" + index}>{text}</span>)
                                                        })}
                                                    </CardText>
                                                    <CardText><span dangerouslySetInnerHTML={{ __html: NotificationsView.formatBody(notification.body) }} /></CardText>
                                                </CardBody>
                                            </Collapse>
                                        </Card>
                                    )
                                })}
                            </Container>
                        )
                    }}
                </Query>)
        }
    }
}

export default inject('appStateStore')(observer(NotificationsView))
