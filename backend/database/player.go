package database

import (
	"database/sql"

	"github.com/allanlepinay/TicTacToe/backend/types"
)

func GetPlayerByName(db *sql.DB, player_name string) (types.Player, error) {
	var player types.Player
	err := db.QueryRow("SELECT id, name, websocket_conn FROM players WHERE name = $1", player_name).Scan(&player.ID, &player.Name, &player.WebsocketConn)
	if err != nil {
		return types.Player{}, err
	}

	return player, nil
}

func GetPlayersByGameId(db *sql.DB, gameId int64) ([]types.Player, error) {
	var players []types.Player

	query := `
		SELECT players.id, players.name, players.websocket_conn
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
