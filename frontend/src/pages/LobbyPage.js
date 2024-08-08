import React, { useState, useEffect } from 'react';
import '../axiosConfig';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

function LobbyPage() {
  const [games, setGames] = useState([]);
  const [playerName, setPlayerName] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    // Récupère les parties disponibles
    const fetchAvailableGames = async () => {
      try {
        const response = await axios.get('/available-games');
        setGames(response.data);
      } catch (error) {
        console.error('Error fetching available games:', error);
      }
    };

    fetchAvailableGames();
    // Rafraîchit les parties disponibles toutes les 10 secondes
    const intervalId = setInterval(fetchAvailableGames, 10000);

    return () => clearInterval(intervalId);
  }, []);

  const handleCreateGame = async () => {
    try {
      const response = await axios.post('/create-game', { playerX: playerName });
      const gameId = response.data.gameId;
      localStorage.setItem('gameId', gameId);
      navigate('/game');
    } catch (error) {
      console.error('Error creating game:', error);
    }
  };

  const handleJoinGame = async (gameId) => {
    try {
      await axios.post('/join-game', { gameId, playerY: playerName });
      localStorage.setItem('gameId', gameId);
      navigate('/game');
    } catch (error) {
      console.error('Error joining game:', error);
    }
  };

  return (
    <div>
      <h1>Lobby</h1>
      <div>
        <label>
          Your Name:
          <input
            type="text"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
            required
          />
        </label>
      </div>
      <button onClick={handleCreateGame}>Create Game</button>
      <h2>Available Games</h2>
      <ul>
        {games.map(game => (
          <li key={game.id}>
            Game ID: {game.id}
            <button onClick={() => handleJoinGame(game.id)}>Join Game</button>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default LobbyPage;
