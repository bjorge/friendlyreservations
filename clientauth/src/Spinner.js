import React, { Component } from 'react';

import { Card, CardHeader, CardBody, CardFooter } from 'reactstrap';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSpinner } from '@fortawesome/free-solid-svg-icons'

import 'bootstrap/dist/css/bootstrap.css';

export default class Spinner extends Component {

  render() {
    return (
      <Card style={{ backgroundColor: 'white', borderColor: 'white' }}>
        <CardHeader style={{ backgroundColor: 'white', borderColor: 'white' }}>&nbsp;</CardHeader>
        <CardBody style={{ backgroundColor: 'white', borderColor: 'white' }}><div className="text-center"><FontAwesomeIcon icon={faSpinner} spin /></div></CardBody>
        <CardFooter style={{ backgroundColor: 'white', borderColor: 'white' }}>&nbsp;</CardFooter>
      </Card>
    )
  }
}


