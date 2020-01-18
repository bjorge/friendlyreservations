import React from 'react'

import AuthCookies from './AuthCookies'

import { BrowserRouter as Router } from 'react-router-dom';

const App = () => (
  <Router>
    <AuthCookies />
    {/* <div>
      <Header />
      <Main />
    </div> */}
  </Router>
)

export default App
