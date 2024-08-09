package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/allanlepinay/TicTacToe/backend/types"
)

func GetGame(db *sql.DB, gameId int64) (types.Game, error) {
	var game types.Game
	err := db.QueryRow("SELECT id, turn, player_x_id, player_o_id FROM games WHERE id = $1", gameId).Scan(&game.ID, &game.Turn, &game.PlayerXId, &game.PlayerOId)
	if err != nil {
		return types.Game{}, err
	}
	return game, nil
}

func UpdateGameStatus(db *sql.DB, gameId int64, status types.GameStatus) error {
	_, err := db.Exec("UPDATE games SET status = $1, updated_at = $2 WHERE id = $3", types.StatusName[status], time.Now(), gameId)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to update game status: %w", err)
	}
	return nil
}

func CreateNewGame(db *sql.DB, player_x_name string, player_o_name string) (types.Game, error) {
	playerO, _ := GetPlayerByName(db, player_o_name)
	playerX, _ := GetPlayerByName(db, player_x_name)

	res, err := db.Query("INSERT INTO games (status, player_x_id, player_o_id) VALUES ($1, $2, $3) RETURNING id", types.StatusName[types.StatusStarted], playerX.ID, playerO.ID)
	if err != nil {
		return types.Game{}, err
	}

	var id int64
	for res.Next() {
		err = res.Scan(&id)
		if err != nil {
			return types.Game{}, err
		}
	}

	game := types.Game{
		ID: id,
		Board: [3][3]string{
			{"", "", ""},
			{"", "", ""},
			{"", "", ""},
		},
		Turn:      "X",
		Status:    types.StatusStarted,
		PlayerXId: playerX.ID,
		PlayerOId: playerO.ID,
	}

	return game, nil
}

func UpdateGameTurn(db *sql.DB, gameId int64) error {
	_, err := db.Exec("UPDATE games SET turn = CASE WHEN turn = 'X' THEN 'O' ELSE 'X' END WHERE id = $1", gameId)
	if err != nil {
		return err
	}
	return nil
}

func GetGameTurn(db *sql.DB, gameId int64) (string, error) {
	var turn string
	err := db.QueryRow("SELECT turn FROM games WHERE id = $1", gameId).Scan(&turn)
	if err != nil {
		return "", err
	}
	return turn, nil
}
