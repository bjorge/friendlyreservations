import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
//import registerServiceWorker from './registerServiceWorker';
import 'bootstrap/dist/css/bootstrap.css';

import { ApolloProvider } from 'react-apollo';

import AppStateStore from './AppStateStore';
import { Provider } from "mobx-react";

const appStateStore = new AppStateStore();

ReactDOM.render((
        <ApolloProvider client={appStateStore.apolloHomeClient}>
                <Provider appStateStore={appStateStore}>
                        <App />
                </Provider>
        </ApolloProvider>
), document.getElementById('root'));
//registerServiceWorker();
