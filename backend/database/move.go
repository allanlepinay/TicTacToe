package database

import (
	"database/sql"
	"fmt"

	gamerules "github.com/allanlepinay/TicTacToe/backend/gameRules"
	"github.com/allanlepinay/TicTacToe/backend/types"
)

func GetBoard(db *sql.DB, move types.Move) [3][3]string {
	var moves []types.Move
	res, err := db.Query("SELECT x, y, player FROM moves WHERE game_id = $1", move.GameId)
	if err != nil {
		fmt.Println(err)
		return [3][3]string{}
	}
	for res.Next() {
		var move types.Move
		err = res.Scan(&move.X, &move.Y, &move.Turn)
		if err != nil {
			fmt.Println(err)
			return [3][3]string{}
		}
		moves = append(moves, move)
	}

	var board = [3][3]string{}
	for _, move := range moves {
		board[move.X][move.Y] = move.Turn
	}

	return board
}

func MakeMove(db *sql.DB, move types.Move) types.Game {
	player, _ := GetPlayerByName(db, move.Username)
	board := GetBoard(db, move)

	gameDb, err := GetGame(db, int64(move.GameId))
	if err != nil {
		fmt.Println("move.go-MakeMove-GetGame-", err)
		return types.Game{}
	}

	if (gameDb.Turn == "X" && gameDb.PlayerXId != player.ID) || (gameDb.Turn == "O" && gameDb.PlayerOId != player.ID) {
		return types.Game{
			ID:     int64(move.GameId),
			Board:  board,
			Turn:   gameDb.Turn,
			Status: types.StatusInProgress,
		}

	}

	_, err = db.Exec("INSERT INTO moves (game_id, player, x, y) VALUES ($1, $2, $3, $4)", move.GameId, gameDb.Turn, move.X, move.Y)
	if err != nil {
		fmt.Println("move.go-MakeMove-INSERT", err)
		return types.Game{
			ID:     int64(move.GameId),
			Board:  board,
			Turn:   gameDb.Turn,
			Status: types.StatusInProgress,
		}
	}

	board = GetBoard(db, move)
	victory, _ := gamerules.CheckVictory(board)

	var game types.Game
	if victory {
		game = types.Game{
			ID:     int64(move.GameId),
			Board:  board,
			Turn:   gameDb.Turn,
			Status: types.StatusTerminated,
		}
		UpdateGameStatus(db, int64(move.GameId), types.StatusTerminated)
	} else {
		var turn string

		if gameDb.Turn == "X" {
			turn = "O"
		} else {
			turn = "X"
		}

		game = types.Game{
			ID:     int64(move.GameId),
			Board:  board,
			Turn:   turn,
			Status: types.StatusInProgress,
		}
		UpdateGameTurn(db, int64(move.GameId))
		// TODO don't really want to update everytime
		UpdateGameStatus(db, int64(move.GameId), types.StatusInProgress)
	}

	return game
}
