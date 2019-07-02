import React, { Component } from 'react';
import {
    Container
} from 'reactstrap';

import { Switch, Route } from "react-router-dom";

import About from './About';
import Home from './Home';

class Main extends Component {
    render() {
        return (
            <Container>
                <Switch>
                    <Route exact path='/' component={Home} />
                    <Route path='/home' component={Home} />
                    <Route path='/about' component={About} />
                </Switch>
            </Container>


        );
    }
}

export default Main;
