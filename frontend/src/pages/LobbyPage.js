import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import LogoutButton from '../components/LogoutButton';
import LeaveQueueButton from '../components/LeaveQueueButton';

const LobbyPage = () => {
  const [status, setStatus] = useState('');
  const [gameId, setGameId] = useState(null);
  const [messages, setMessages] = useState([]); // Added to store incoming messages
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
    </div>
  );
};

export default LobbyPage;