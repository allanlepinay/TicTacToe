import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useSelector } from 'react-redux';

const PlayerPage = () => {
    const { id } = useParams();
    const [player, setPlayer] = useState(null);
    const [games, setGames] = useState([]);
    const socket = useSelector((state) => state.websocket.connection);

    useEffect(() => {
        if (socket) {
            socket.send(JSON.stringify({
                type: "getPlayerProfile",
                message: JSON.stringify({"playerId": id})
            }));
        }
    }, [socket, id]);

    useEffect(() => {
        if (socket) {
            socket.onmessage = (event) => {
                const data = JSON.parse(event.data);
                if (data.type === 'playerProfile') {
                    setPlayer(data);
                    setGames(data.games);
                }
            };
        }
    }, [socket]);

    const getStatusName = (status) => {
        // There is probably a better way to do this
        switch(status) {
            case 0: return "Started";
            case 1: return "In-Progress";
            case 2: return "Terminated";
            default: return "Unknown";
        }
    };

    return (
        <div>
            {player ? (
                <div>
                    <h1>Player Profile: {player.name}</h1>
                    <p>ID: {player.id}</p>
                    <p>Wins: {player.wins}</p>
                    <p>Losses: {player.loses}</p>
                    <p>Draws: {player.draw}</p>
                    <h2>Games:</h2>
                    <ul>
                        {games.map(game => (
                            <li key={game.id}>
                                Game ID: {game.id}, Status: {getStatusName(game.status)}
                            </li>
                        ))}
                    </ul>
                </div>
            ) : (
                <p>Loading...</p>
            )}
        </div>
    );
};

export default PlayerPage;
