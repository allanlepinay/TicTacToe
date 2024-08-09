package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	gamerules "github.com/allanlepinay/TicTacToe/backend/gameRules"
	"github.com/allanlepinay/TicTacToe/backend/types"
	"github.com/gorilla/mux"
)

func GetBoard(db *sql.DB, w http.ResponseWriter, r *http.Request) [3][3]string {
	vars := mux.Vars(r)
	gameId := vars["id"]
	var moves []types.Move
	res, err := db.Query("SELECT x, y, player FROM moves WHERE game_id = $1", gameId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Moves not found", http.StatusNotFound)
		return [3][3]string{}
	}
	for res.Next() {
		var move types.Move
		err = res.Scan(&move.X, &move.Y, &move.Player)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to scan move", http.StatusInternalServerError)
			return [3][3]string{}
		}
		moves = append(moves, move)
	}

	var board = [3][3]string{}
	for _, move := range moves {
		board[move.X][move.Y] = move.Player
	}

	return board
}

func MakeMove(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameId := vars["id"]
	gameIdInt, _ := strconv.ParseInt(gameId, 0, 64)

	move := types.Move{}
	err := json.NewDecoder(r.Body).Decode(&move)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	player, _ := GetPlayerByName(db, move.Username)

	gameDb, err := GetGame(db, gameIdInt)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get game", http.StatusInternalServerError)
		return
	}

	if (gameDb.Turn == "X" && gameDb.PlayerXId != player.ID) || (gameDb.Turn == "O" && gameDb.PlayerOId != player.ID) {
		http.Error(w, "It's not your turn", http.StatusForbidden)
		return
	}

	_, err = db.Exec("INSERT INTO moves (game_id, player, x, y) VALUES ($1, $2, $3, $4)", gameId, gameDb.Turn, move.X, move.Y)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to make move", http.StatusInternalServerError)
		return
	}

	board := GetBoard(db, w, r)
	victory, _ := gamerules.CheckVictory(board)

	var game types.Game
	if victory {
		game = types.Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   gameDb.Turn,
			Status: types.StatusTerminated,
		}
		UpdateGameStatus(db, gameIdInt, types.StatusTerminated)
	} else {
		var turn string

		if gameDb.Turn == "X" {
			turn = "O"
		} else {
			turn = "X"
		}

		game = types.Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   turn,
			Status: types.StatusInProgress,
		}
		UpdateGameTurn(db, gameIdInt)
		// TODO don't really want to update everytime
		UpdateGameStatus(db, gameIdInt, types.StatusInProgress)
	}

	json.NewEncoder(w).Encode(game)
}
