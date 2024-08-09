package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/allanlepinay/TicTacToe/backend/auth"
	"github.com/allanlepinay/TicTacToe/backend/database"
	gamerules "github.com/allanlepinay/TicTacToe/backend/gameRules"
	"github.com/allanlepinay/TicTacToe/backend/types"
	"github.com/allanlepinay/TicTacToe/backend/utils"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var leaveQueueMessage struct {
	Username string `json:"username"`
}

var waitingPlayers = make(chan string, 100)
var mutex = &sync.Mutex{}

func main() {
	// .env load
	viper.SetConfigFile("../.env")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	connStr := viper.GetString("DATABASE_CONN_STRING")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	r := mux.NewRouter()
	// Not protected route
	r.HandleFunc("/register", auth.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		Register(db, w, r)
	}))
	r.HandleFunc("/login", auth.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		Login(db, w, r)
	}))
	r.HandleFunc("/refresh-token", auth.WithCORS(auth.RefreshTokenHandler))
	// Protected route
	r.HandleFunc("/game/{id}", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		UpdateGameBoard(db, w, r)
	})))
	r.HandleFunc("/game/{id}/move", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		database.MakeMove(db, w, r)
	})))
	r.HandleFunc("/verify-token", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
	})))
	r.HandleFunc("/search-game", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		SearchGame(db, w, r)
	})))
	r.HandleFunc("/leave-queue", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		LeaveQueue(db, w, r)
	})))

	http.ListenAndServe(":8080", r)
}

func Register(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var player types.Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(player.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO players (name, password_hash) VALUES ($1, $2)", player.Name, hashedPassword)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to register player", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var player types.Player

	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM players WHERE name = $1 LIMIT 1", player.Name).Scan(&storedHash)
	if err != nil {
		http.Error(w, "Invalid name or password", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(player.Password, storedHash) {
		http.Error(w, "Invalid name or password", http.StatusUnauthorized)
		return
	}

	// Generate access and refresh tokens
	accessToken, err := auth.GenerateAccessToken(player.Name)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(player.Name)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Respond with both tokens
	response := types.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SearchGame(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	select {
	case opponent := <-waitingPlayers:
		// todo probably not the best way
		if username == opponent {
			removeFromQueue(username)
			http.Error(w, "Can't play with oneself", http.StatusInternalServerError)
			// TODO display waiting again
			// json.NewEncoder(w).Encode(map[string]string{"status": "waiting"})
			return
		}
		// Found an opponent, create a new game
		game, err := database.CreateNewGame(db, username, opponent)
		if err != nil {
			http.Error(w, "Failed to create game", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(game)
	default:
		// No opponent found, add self to waiting list
		mutex.Lock()
		defer mutex.Unlock()
		waitingPlayers <- username
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "waiting"})
	}
}

func LeaveQueue(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	err := json.NewDecoder(r.Body).Decode(&leaveQueueMessage)
	if err != nil {
		http.Error(w, "Invalid request leaveQueueMessage", http.StatusBadRequest)
		return
	}

	if leaveQueueMessage.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	removeFromQueue(leaveQueueMessage.Username)

	// Inform the client that they've been removed from the queue
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed from queue"})
}

func removeFromQueue(username string) {
	mutex.Lock()
	defer mutex.Unlock()

	// Create a temporary channel to hold players we want to keep
	tempChannel := make(chan string, len(waitingPlayers))

	// Iterate through the waiting players
	for len(waitingPlayers) > 0 {
		player := <-waitingPlayers
		if player != username {
			// If the player is not the one we want to remove, add them back to the temp channel
			tempChannel <- player
		}
	}

	// Replace the waitingPlayers channel with the temp channel
	waitingPlayers = tempChannel
}

func UpdateGameBoard(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameId := vars["id"]
	gameIdInt, _ := strconv.ParseInt(gameId, 0, 64)

	board := database.GetBoard(db, w, r)
	if board == [3][3]string{} {
		board = [3][3]string{
			{"", "", ""},
			{"", "", ""},
			{"", "", ""},
		}
	}
	victory, _ := gamerules.CheckVictory(board)
	turn, _ := database.GetGameTurn(db, gameIdInt)
	var game types.Game
	if victory {
		game = types.Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   turn,
			Status: types.StatusTerminated,
		}
		database.UpdateGameStatus(db, gameIdInt, types.StatusTerminated)
	} else {

		game = types.Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   turn,
			Status: types.StatusStarted,
		}
	}

	json.NewEncoder(w).Encode(game)
}
