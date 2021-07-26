import React, { Component, Fragment } from 'react'
import { Route } from 'react-router-dom'
// import './App.scss'

import { LandingPage } from '../LandingPage/LandingPage'


class App extends Component {
  constructor () {
    super()

    this.state = {}
  }

  render () {

    return (
      <Fragment>
        <Route exact path='/' render={() => (
            <LandingPage />
          )} />
      </Fragment>
    )
  }
}

export default App