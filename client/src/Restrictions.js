import React, { Component } from 'react';

import 'bootstrap/dist/css/bootstrap.css';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { faMinus } from '@fortawesome/free-solid-svg-icons'

import { inject, observer } from "mobx-react";

import { Query } from 'react-apollo';
import Spinner from './Spinner';
import ErrorModal from './ErrorModal';
import gql from 'graphql-tag';


import {
    Collapse,
    Card,
    CardBody,
    Container,
    Button,
    Label,
    Input,

} from 'reactstrap';

import {
    Redirect
} from "react-router-dom";
import BlackoutRestriction from './BlackoutRestriction';
import MembershipRestriction from './MembershipRestriction';

const GET_RESTRICTIONS_GQL = gql`
query Restrictions(
  $propertyId: String!) {
  property(id: $propertyId) {
    eventVersion
    restrictions {
        description
        restrictionId
        restriction {
          __typename
              ... on BlackoutRestriction {
            startDate
            endDate
          }
          ... on MembershipRestriction {
            inDate
            outDate
            gracePeriodOutDate
            prePayStartDate
            amount
          }
        }
      }
  }
}
`;

class Restrictions extends Component {
    constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.handleChange = this.handleChange.bind(this);


        this.state = {
            collapse: null,
            restrictionselect: "BlackoutRestriction",
            cachedProperty: null,
        };
    }

    toggle(event) {
        this.setState({ collapse: this.state.collapse === event ? null : event });
    }
    render() {

        const { collapse } = this.state;
        const propertyId = this.props.appStateStore.propertyId ? this.props.appStateStore.propertyId : null;

        if (propertyId === null) {
            return (<Redirect to="/propertyselect" />)
        } else {
            const apolloClient = this.props.appStateStore.apolloClient;
            const queryKey = this.props.appStateStore.apolloQueryKey;
            return (
                <Query query={GET_RESTRICTIONS_GQL} fetchPolicy='no-cache'
                    client={apolloClient} key={queryKey}
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

                            if (property === null) {
                                return (<div>No data from service, please refresh and try again</div>)
                            }

                            // make restrictions into a map
                            var restrictions = property.restrictions.reduce(function (map, obj) {
                                map[obj.restrictionId] = { restrictionId: obj.restrictionId, restriction: obj.restriction, description: obj.description };
                                return map;
                            }, {});

                            var nameConstraints = [];
                            for (var i=0; i<property.restrictions.length; i++) {
                                nameConstraints.push(property.restrictions[i].description);
                            }

                        }
                        return (
                            <Container>
                                {/* {error && <ErrorModal error={error} />} */}
                                {Object.keys(restrictions).map(key => {
                                    return (
                                        <Card key={key}>
                                            <Button className="text-left" onClick={() => this.toggle(key)}>
                                                {collapse === key ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                                &nbsp;
                                            {restrictions[key].restriction.__typename}{': '}{restrictions[key].description}
                                            </Button>
                                            <Collapse isOpen={collapse === key}>
                                                <CardBody>
                                                    {{
                                                        'BlackoutRestriction': (
                                                            <BlackoutRestriction viewinfo={restrictions[key]} />
                                                        )
                                                    }[restrictions[key].restriction.__typename]}
                                                </CardBody>
                                            </Collapse>
                                            <Collapse isOpen={collapse === key}>
                                                <CardBody>
                                                    {{
                                                        'MembershipRestriction': (
                                                            <MembershipRestriction viewinfo={restrictions[key]} />
                                                        )
                                                    }[restrictions[key].restriction.__typename]}
                                                </CardBody>
                                            </Collapse>
                                        </Card>
                                    )
                                })}
                                <Card key='abc'>
                                    <Button className="text-left" onClick={() => this.toggle('abc')}>
                                        {collapse === 'abc' ? <FontAwesomeIcon icon={faMinus} pull="left" /> : <FontAwesomeIcon icon={faPlus} pull="left" />}
                                        &nbsp;
                                        'New Restriction'
                                            </Button>
                                    <Collapse isOpen={collapse === 'abc'}>
                                        <CardBody>

                                            <Label for="restrictionselect">Select Restriction Type</Label>
                                            <Input type="select" value={this.state.restrictionselect} name="restrictionselect" id="restrictionselect" onChange={this.handleChange}>
                                                {/* <option hidden >Select a restriction type</option> */}
                                                <option value="BlackoutRestriction">BlackoutRestriction</option>
                                                <option value="MembershipRestriction">MembershipRestriction</option>
                                            </Input>
                                            {this.state.restrictionselect && <div>
                                                {{
                                                    BlackoutRestriction: (
                                                        <BlackoutRestriction constraints={nameConstraints}/>
                                                    ),
                                                    MembershipRestriction: (
                                                        <MembershipRestriction constraints={nameConstraints}/>
                                                    )
                                                }[this.state.restrictionselect]}
                                            </div>}
                                        </CardBody>
                                    </Collapse>
                                </Card>
                            </Container>);
                    }}
                </Query>)

        }
    }

    handleChange(event) {
        const target = event.target;
        switch (target.name) {
            case 'restrictionselect':
                this.setState({ restrictionselect: target.value });
                break;
            default:
                console.log("unexpected setting: " + target.name)
                break;
        }
    }
}

export default inject('appStateStore')(observer(Restrictions))

