package database

import (
	"database/sql"

	"github.com/allanlepinay/TicTacToe/backend/types"
)

func GetPlayerByName(db *sql.DB, player_name string) (types.Player, error) {
	var player types.Player
	err := db.QueryRow("SELECT id, name FROM players WHERE name = $1", player_name).Scan(&player.ID, &player.Name)
	if err != nil {
		return types.Player{}, err
	}

	return player, nil
}
