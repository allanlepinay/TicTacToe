import React from 'react';
import { Link } from 'react-router-dom';
import LogoutButton from '../components/LogoutButton';


function HomePage() {
  return (
    <div>
        <h1>Welcome to Tic Tac Toe</h1>
        <div>
            <div>
              <Link to="/login">Login</Link>
            </div>
            <div>
              <Link to="/register">Register</Link>
            </div>
        </div>
        <LogoutButton />
    </div>
  );
}

export default HomePage;
