import React, { Component } from 'react';

import 'bootstrap/dist/css/bootstrap.css';

import { inject, observer } from "mobx-react";

import {
  Redirect
} from "react-router-dom";



class ExitProperty extends Component {

  componentDidMount() {
    if (this.props.appStateStore) {
      this.props.appStateStore.clearAll()
    }
  }

  render() {
    return (
      <Redirect to="/propertyselect" />
    )
  }
}

export default inject('appStateStore')(observer(ExitProperty))

