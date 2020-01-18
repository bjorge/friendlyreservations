import React, { Component } from 'react';

import { Query } from 'react-apollo';
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import ContentForm from './ContentForm';
import gql from 'graphql-tag';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faMinus } from '@fortawesome/free-solid-svg-icons'

import { inject, observer } from "mobx-react";

import {
    Collapse,
    Card,
    CardBody,
    Container,
    Button,
    Form,
    FormGroup,
    Label,
    Col,
    Input,
} from 'reactstrap';

import {
    Redirect
} from "react-router-dom";

const GET_CONTENTS_GQL = gql`
query Contents(
  $propertyId: String!) {
  property(id: $propertyId) {
    eventVersion
    contents {
        name
        rendered
        template
        comment
        createDateTime
        author {
          nickname
        }
        defaultTemplate
        default
    }
  }
}
`;

class Contents extends Component {
    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.toggleNewTemplateForm = this.toggleNewTemplateForm.bind(this);
        this.turnOffModals = this.turnOffModals.bind(this);

        this.state = {
            collapse: null,
            showNewTemplateForm: false,
            cachedProperty: null,
        };

    }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }

    toggleNewTemplateForm(key) {
        this.setState({ showNewTemplateForm: key });
    }

    turnOffModals = () => {
        this.setState({ showNewTemplateForm: false });
    }

    render() {
        const { collapse } = this.state;
        const me = this;
        const apolloClient = this.props.appStateStore.apolloClient;
        const queryKey = this.props.appStateStore.apolloQueryKey;

        return (
            <div>
                {{
                    'LIST_PROPERTIES': (
                        <Redirect to="/propertyselect" />
                    ),
                    'MANAGE_PROPERTY': (
                        <Query client={apolloClient} key={queryKey} query={GET_CONTENTS_GQL} fetchPolicy='no-cache'
                            variables={{
                                propertyId: this.props.appStateStore.propertyId
                            }}
                            onCompleted={(data) => {
                                if (data.property !== undefined) {
                                    this.props.appStateStore.setPropertyEventVersion(data.property.eventVersion)
                                    this.setState({ cachedProperty: data.property });
                                }
                            }}
                        >
                            {({ loading, error, data }) => {
                                if (loading) { return (<Spinner />); }
                                if (error) { return (<ErrorModal error={error} />); }
                                if (data) {

                                    var property = this.state.cachedProperty;
                                    if (data.property !== undefined) {
                                        property = data.property;
                                    }

                                    if (property === undefined) {
                                        return (<div>No data from service, please refresh and try again</div>)

                                    }
                                    var contents = property.contents.reduce(function (map, obj) {
                                        map[obj.name] = obj;
                                        return map;
                                    }, {});

                                }
                                return (
                                    <Container>

                                        {contents && Object.keys(contents).map(key => {

                                            return (
                                                <Card key={key}>
                                                    <Button className="text-left" onClick={() => me.toggle(key)}>
                                                        {collapse === key ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                                        &nbsp;
                                            {key}
                                                    </Button>
                                                    <Collapse isOpen={collapse === key}>
                                                        <CardBody>
                                                            <Form>
                                                                <FormGroup row>
                                                                    <Label for="id1" sm={2}>Current Template</Label>
                                                                    <Col sm={10}>
                                                                        <Input readOnly={true} type="textarea" name="name1" id="id1" value={contents[key].template} />
                                                                    </Col>
                                                                </FormGroup>
                                                                <FormGroup row>
                                                                    <Label for="id1" sm={2}>Displayed Template</Label>
                                                                    <Col sm={10}>
                                                                        <Input readOnly={true} type="textarea" name="name1" id="id2" value={contents[key].rendered} />
                                                                    </Col>
                                                                </FormGroup>
                                                                <FormGroup check>
                                                                    <Label check>
                                                                        <Input readOnly={true} type="checkbox" id="checkbox2" checked={contents[key].default} />{' '}
                                                                        Using default template
                                                                    </Label>
                                                                </FormGroup>
                                                                <hr />
                                                                <FormGroup row>
                                                                    <Label for="id1" sm={2}>Default Template</Label>
                                                                    <Col sm={10}>
                                                                        <Input readOnly={true} type="textarea" name="name2" id="id3" value={contents[key].defaultTemplate} />
                                                                    </Col>
                                                                </FormGroup>
                                                            </Form>
                                                            <ContentForm content={contents[key]} showform={key === me.state.showNewTemplateForm} exitModal={me.turnOffModals} />
                                                            <div className="text-center">
                                                                <Button color="primary" onClick={() => me.toggleNewTemplateForm(key)}>Edit Template</Button>
                                                            </div>
                                                        </CardBody>
                                                    </Collapse>
                                                </Card>
                                            )
                                        })}

                                    </Container>);
                            }}
                        </Query>),

                }[this.props.appStateStore.headerPageState]}
            </div>

        );
    }
}

export default inject('appStateStore')(observer(Contents))

