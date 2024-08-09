import React, { useState, useEffect } from 'react';
import '../axiosConfig';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import LogoutButton from '../components/LogoutButton';
import LeaveQueueButton from '../components/LeaveQueueButton';

const LobbyPage = () => {
  const [status, setStatus] = useState('');
  const [game, setGame] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const searchGame = async () => {
      try {
        const response = await axios.get('/search-game', {
          params: {
            // todo this have to be changed when user check will be refractor
            username: localStorage.getItem('username')
          }
        });
        if (response.data.status === 'waiting') {
          setStatus('Waiting for an opponent...');
        } else {
          setGame(response.data);
        }
      } catch (error) {
        console.error('Error searching for game:', error);
      }
    };

    searchGame();
  }, []);

  useEffect(() => {
    if (game) {
      navigate(`/game/${game.id}`);
    }
  }, [game, navigate]);

  return (
    <div>
      <h1>Lobby</h1>
      {status && <p>{status}</p>}
      <LeaveQueueButton />
      <LogoutButton />
    </div>
  );
};

export default LobbyPage;
