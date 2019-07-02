import React, { Component } from 'react';
// import logo from './logo.svg';
// import './App.css';

import { HashRouter } from "react-router-dom";

import Header from './Header'
import Main from './Main'

class App extends Component {
  render() {
    return (
    <HashRouter>
      <div>
        <Header />
        <Main />
      </div>
    </HashRouter>
      // <div className="App">
      //   <header className="App-header">
      //     <img src={logo} className="App-logo" alt="logo" />
      //     <h1 className="App-title">Welcome to React</h1>
      //   </header>
      //   <p className="App-intro">
      //     To get started, edit <code>src/App.js</code> and save to reload.
      //   </p>
      //   <p><a href="/fr/">FR Anchor</a></p><br/>
      // </div>
    );
  }
}

export default App;
