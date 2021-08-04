import React from 'react';
import background from '../../assets/images/shipit.jpg';
import { NavBar } from '../NavBar/NavBar.js';

export const LandingPage = () => (
  <div className="x">
    <NavBar />
    <img src={background} alt="a container ship at sea" />
  </div>
)