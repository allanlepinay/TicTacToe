import React from 'react';
import Game from '../components/Game';
import LogoutButton from '../components/LogoutButton';

function GamePage() {
  return (
    <div>
      <h1>Game Page</h1>
        <Game />    
        <LogoutButton />
    </div>
  );
}

export default GamePage;
