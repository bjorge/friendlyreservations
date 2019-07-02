import React from 'react'
import Header from './Header'
import Main from './Main'

import { HashRouter } from 'react-router-dom';

const App = () => (
  <HashRouter>
    <div>
      <Header />
      <Main />
    </div>
  </HashRouter>
)

export default App
