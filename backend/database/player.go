package database

import (
	"database/sql"
	"fmt"

	"github.com/allanlepinay/TicTacToe/backend/types"
)

func GetPlayerByName(db *sql.DB, username string) (types.Player, error) {
	var player types.Player
	err := db.QueryRow("SELECT id, name, websocket_conn FROM players WHERE name = $1", username).Scan(&player.ID, &player.Name, &player.WebsocketConn)
	if err != nil {
		return types.Player{}, err
	}

	return player, nil
}

func GetPlayersByGameId(db *sql.DB, gameId int64) ([]types.Player, error) {
	var players []types.Player

	query := `
		SELECT players.id, players.name
		FROM players
		JOIN games ON players.id = games.player_x_id OR players.id = games.player_o_id
		WHERE games.id = $1
	`
	rows, err := db.Query(query, gameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var player types.Player
		err = rows.Scan(&player.ID, &player.Name, &player.WebsocketConn)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil

}

func GetPlayerProfile(db *sql.DB, playerId string) (types.Player, []types.Game, error) {
	var player types.Player
	err := db.QueryRow("SELECT id, name FROM players WHERE id = $1", playerId).Scan(&player.ID, &player.Name)
	if err != nil {
		return types.Player{}, nil, err
	}

	// todo use a function in game.go
	rows, err := db.Query("SELECT id, status FROM games WHERE player_x_id = $1 OR player_o_id = $1", player.ID)
	if err != nil {
		return types.Player{}, nil, err
	}
	defer rows.Close()

	var games []types.Game
	for rows.Next() {
		var game types.Game
		var statusString string
		err := rows.Scan(&game.ID, &statusString)
		if err != nil {
			return types.Player{}, nil, err
		}
		for status, name := range types.StatusName {
			if name == statusString {
				game.Status = int64(status)
				break
			}
		}
		if game.Status == 0 && statusString != types.StatusName[types.StatusStarted] {
			return types.Player{}, nil, fmt.Errorf("unknown game status: %s", statusString)
		}
		games = append(games, game)
	}

	return player, games, nil
}
