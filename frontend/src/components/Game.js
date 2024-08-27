import React, { useState, useEffect, useRef } from 'react';
import Board from './Board';
import { useSelector } from 'react-redux';

function Game() {
    const [board, setBoard] = useState([['', '', ''], ['', '', ''], ['', '', '']]);
    const [turn, setTurn] = useState('X');
    const [gameOver, setGameOver] = useState(false);
    const [winner, setWinner] = useState('');
    const [wsStatus, setWsStatus] = useState('Disconnected');
    const socket = useSelector((state) => state.websocket.connection);
    const [gameId, setGameId] = useState(window.location.pathname.split('/').pop());

    useEffect(() => {
        if (socket) {
          setWsStatus('Connected');
          const username = localStorage.getItem('username');
          socket.send(JSON.stringify({
            type: "JoinGame",
            message: "JoinGame",
            gameId: parseInt(gameId),
            username: username
          }));
    
          socket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
          };
    
          socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            setWsStatus('Error');
          };
    
          socket.onclose = () => {
            console.log('WebSocket connection closed');
            setWsStatus('Disconnected');
          };
        }
      }, [socket, gameId]);

    const handleWebSocketMessage = (data) => {
        switch (data.type) {
            case 'move':
                var game = JSON.parse(data.message);
                setBoard(game['board']);
                if (game['status'] == 2) { // Status Terminated
                    setGameOver(true);
                    setWinner(game['turn'])
                } else {
                    setTurn(game['turn']);
                }
                break;
            default:
                console.log('Unknown message type:', data.type);
        }
    };

    const handleClick = (i, j) => {
        if (board[i][j] !== '' || gameOver) return;

        const move = { 
            type: 'move',
            x: i, 
            y: j, 
            gameId: parseInt(gameId),
            username: localStorage.getItem('username'),
            turn: turn
        };
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(move));
        } else {
            console.error('WebSocket is not connected');
        }
    };

    return (
        <div>
            <Board board={board} onClick={handleClick} />
            <div>Current Turn: {turn}</div>
            {gameOver && <div>{winner} has won!</div>}
            <div>WebSocket Status: {wsStatus}</div>
        </div>
    );
}

export default Game;