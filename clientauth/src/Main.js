import React, { Component } from 'react';
import {
    Container
} from 'reactstrap';

import { Switch, Route } from "react-router-dom";

import HomeView from './Home';
import AboutView from './About';
import CreateReservation from './CreateReservation';
import ListProperties from './ListProperties';
import PropertySelect from './PropertySelect';
import PropertyHome from './PropertyHome';
import ExitProperty from './ExitProperty';
import Restrictions from './Restrictions';
import Users from './Users';
import Contents from './Contents';
import Membership from './Membership';
import Reservations from './Reservations';
import LedgerView from './LedgerView';
import NotificationsView from './NotificationsView';
import Settings from './Settings';
import AdminReservations from './AdminReservations';
import AdminAdvanced from './AdminAdvanced';

// import Logout from './Logout';

// loading example if we decide app is too big
//import ListProperties from './ListProperties';
// import Loadable from 'react-loadable';

// function Loading() {
//     return <h3>Loading...</h3>;
// }

// const ListPropertiesLoader = Loadable({
//     loader: () => import('./ListProperties'),
//     loading: Loading
// });

class Main extends Component {
    render() {
        return (
            <Container>
                <Switch>
                    <Route exact path='/' component={PropertySelect} />
                    <Route path='/home' component={HomeView} />
                    <Route path='/about' component={AboutView} />
                    <Route path='/listproperties' component={ListProperties} />
                    <Route path='/propertyselect' component={PropertySelect} />
                    <Route path='/createreservation' component={CreateReservation} />
                    <Route path='/propertyhome' component={PropertyHome} />
                    <Route path='/exitproperty' component={ExitProperty} />
                    <Route path='/restrictions' component={Restrictions} />
                    <Route path='/users' component={Users} />
                    <Route path='/contents' component={Contents} />
                    <Route path='/membership' component={Membership} />
                    <Route path='/reservations' component={Reservations} />
                    <Route path='/ledger' component={LedgerView} />
                    <Route path='/notifications' component={NotificationsView} />
                    <Route path='/settings' component={Settings} />
                    <Route path='/adminreservations' component={AdminReservations} />
                    <Route path='/adminadvanced' component={AdminAdvanced} />
                    {/* <Route path='/logout' component={Logout} /> */}
                </Switch>
            </Container>


        );
    }
}

export default Main;
