import React, { useState, useEffect } from 'react';
import Board from './Board';
import axios from 'axios';

function Game() {
    const [board, setBoard] = useState([['', '', ''], ['', '', ''], ['', '', '']]);
    const [turn, setTurn] = useState('X');
    const [gameId, setGameId] = useState(null);
    const [gameOver, setGameOver] = useState(false);
    const [winner, setWinner] = useState('');

    useEffect(() => {
        // Fetch initial game state from the backend
        axios.post(`${process.env.REACT_APP_API_URL}/game`)
            .then(response => {
                setGameId(response.data.id);
                setBoard(response.data.board);
                setTurn(response.data.turn);
            })
            .catch(error => {
                console.error('Error fetching game state:', error);
            });
    }, []);

    const handleClick = (i, j) => {
        if (board[i][j] !== '' || gameOver) return; // Prevent clicking on filled cells or after game over

        axios.post(`${process.env.REACT_APP_API_URL}/game/${gameId}/move`, { x: i, y: j, player: turn })
            .then(response => {
                setBoard(response.data.board);
                setTurn(response.data.turn);
                if (response.data.status === 2) { // StatusTerminated
                    setGameOver(true);
                    setWinner(response.data.turn);
                }
            })
            .catch(error => {
                console.error('Error making move:', error);
            });        
    };

    return (
        <div>
            <Board board={board} onClick={handleClick} />
            <div>Current Turn: {turn}</div>
            {gameOver && <div>{winner} has won!</div>}
        </div>
    );
}

export default Game;