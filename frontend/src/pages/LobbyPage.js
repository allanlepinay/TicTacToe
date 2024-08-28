import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import LogoutButton from '../components/LogoutButton';
import LeaveQueueButton from '../components/LeaveQueueButton';

const LobbyPage = () => {
  const [status, setStatus] = useState('');
  const [gameId, setGameId] = useState(null);
  const [messages, setMessages] = useState([]); // Added to store incoming messages
  const [searchPlayerId, setSearchPlayerId] = useState(''); // Added to store the player ID to search
  const navigate = useNavigate();
  const socket = useSelector((state) => state.websocket.connection);

  useEffect(() => {
    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'gameCreated') {
        setGameId(data.gameId)
        navigate(`/game/${data.gameId}`);
      } else if (data.type === 'waiting') {
        setStatus('Waiting for an opponent...');
      } else if (data.type === 'playerProfile') {
        navigate(`/player/${data.id}`);
      } else if (data.type === 'message') {
        setMessages(prevMessages => [...prevMessages, data.message]);
      }
    };

    return () => {};
  }, []);

  const joinQueue = () => {
    if (socket) {
      socket.send(JSON.stringify({
        type: "JoinQueue",
        message: "JoinQueue",
        gameId: -1, 
        username: localStorage.getItem('username')
      }));
    }
  };

  const ping = () => {
    if (socket) {
      socket.send(JSON.stringify({
        type: "ping",
        message: "Ping",
        gameId: -1, 
        username: localStorage.getItem('username')
      }));
    }
  };

  const searchPlayer = () => {
    if (socket) {
      socket.send(JSON.stringify({
        type: "getPlayerProfile",
        message: JSON.stringify({"playerId": searchPlayerId})
      }));
    }
  };

  const handleSearchPlayerIdChange = (event) => {
    setSearchPlayerId(event.target.value);
  };

  return (
    <div>
      <h1>Lobby</h1>
      {status && <p>{status}</p>}
      {messages.length > 0 && (
        <div>
          <h2>Messages:</h2>
          <ul>
            {messages.map((message, index) => (
              <li key={index}>{message}</li>
            ))}
          </ul>
        </div>
      )}
      <button onClick={joinQueue}>
        Join queue
      </button>
      <button onClick={ping}>
        Ping
      </button>
      <LeaveQueueButton />
      <LogoutButton />
      <div>
        <input type="text" value={searchPlayerId} onChange={handleSearchPlayerIdChange} placeholder="Search player by ID" />
        <button onClick={searchPlayer}>Search</button>
      </div>
    </div>
  );
};

export default LobbyPage;