import React, { Component } from 'react';

import 'bootstrap/dist/css/bootstrap.css';

// import { graphql } from 'react-apollo';
import gql from 'graphql-tag';
import { Query } from 'react-apollo';
import { inject, observer } from "mobx-react";

import {
  Collapse,
  Card,
  CardBody,
  Container,
  Button,
  Row,
  Col
} from 'reactstrap';

// import CurrencyInput from 'react-currency-input';
import * as NumberFormat from 'react-number-format';


const GET_PROPERTIES = gql`
  {
    system {
      properties {
        id
        settings {
          name
          memberrate
          nonmemberrate
          currency
          allownonmembers
        }
      }
    }
  }
`;

class ListProperties extends Component {
  constructor(props) {
    super(props);
    this.toggle = this.toggle.bind(this);
    this.state = {
      collapse: null
    };
  }

  toggle(e) {
    console.log("toggle");
    let event = e.target.dataset.event;
    this.setState({ collapse: this.state.collapse === event ? null : event });
  }
  render() {
    const { collapse } = this.state;
    const apolloClient = this.props.appStateStore.apolloClient;

    return (
      <Query client={apolloClient} query={GET_PROPERTIES}
        fetchPolicy='no-cache'>
        {({ loading, error, data }) => {
          if (loading) return <div>Loading...</div>;
          // TODO: add button to refresh this page (ex. Redirect or Link)
          // example: https://medium.com/@anneeb/redirecting-in-react-4de5e517354a
          if (error) return <div>ERROR: +{error.graphQLErrors[0].message}</div>;

          return (
            <Container>
              {data.system.properties && data.system.properties.map(property => {
                return (
                  <Card key={property.id}>
                    <Button className="text-left" onClick={this.toggle} data-event={property.id}>{property.settings.name}</Button>
                    <Collapse isOpen={collapse === property.id}>
                      <CardBody>
                        <Row>
                          <Col>Currency</Col>
                          <Col>{property.settings.currency}</Col>
                        </Row>
                        <Row>
                          <Col>Member rate</Col>
                          <Col><NumberFormat value={property.settings.memberrate} displayType={'text'} decimalScale={2} fixedDecimalScale={true} /></Col>
                        </Row>
                        <Row>
                          <Col>Allow non-members</Col>
                          <Col>{property.settings.allownonmembers === true ? 'true' : 'false'}</Col>
                        </Row>
                        <Row>
                          <Col>Non Member rate</Col>
                          <Col><NumberFormat value={property.settings.nonmemberrate} displayType={'text'} decimalScale={2} fixedDecimalScale={true} /></Col>
                        </Row>
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

export default inject('appStateStore')(observer(ListProperties))

