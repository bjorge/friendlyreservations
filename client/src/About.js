import React, { Component } from 'react';

import {
    Card,
    Container,
    CardBody,
    CardTitle,
    CardText
} from 'reactstrap';

export default class About extends Component {

    render() {
        return (
            <Container>
                <Card>
                    <CardBody>
                        <CardTitle>About!</CardTitle>
                        <CardText>For more information visit <a href="https://github.com/bjorge/friendlyreservations/blob/master/README.md">friendlyreservations</a>.</CardText>
                        <CardText>1.6</CardText>
                   </CardBody>
                </Card>
            </Container>
        );
    }
}
