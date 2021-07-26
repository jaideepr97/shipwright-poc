// import React from 'react';
// import { render } from 'react-dom';

// import HelloWorld from './HelloWorld';

// render(<HelloWorld />, document.getElementById('root'));

import React from 'react'
import ReactDOM from 'react-dom'
// import './index.scss'

import App from './components/App/App'
import { HashRouter } from 'react-router-dom'

const appJsx = (
  <HashRouter>
    <App />
  </HashRouter>
)

ReactDOM.render(appJsx, document.getElementById('root'))