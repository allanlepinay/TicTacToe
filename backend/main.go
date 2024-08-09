package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Game struct {
	ID        int64        `json:"id"`
	Board     [3][3]string `json:"board"`
	Turn      string       `json:"turn"`
	Status    int64        `json:"status"`
	PlayerXId int64        `json:"player_x_id"`
	PlayerOId int64        `json:"player_o_id"`
}

type Move struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Player   string `json:"player"`
	Username string `json:"username"`
}

type Player struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Wins     int64  `json:"wins"`
	Loses    int64  `json:"loses"`
	Draw     int64  `json:"draw"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GameStatus int

const (
	StatusStarted = iota
	StatusInProgress
	StatusTerminated
)

var statusName = map[GameStatus]string{
	StatusStarted:    "Started",
	StatusInProgress: "In-Progress",
	StatusTerminated: "Terminated",
}

var waitingPlayers = make(chan string, 100)
var mutex = &sync.Mutex{}

func main() {
	// .env load
	viper.SetConfigFile(".env")
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
	r.HandleFunc("/register", withCORS(func(w http.ResponseWriter, r *http.Request) {
		Register(db, w, r)
	}))
	r.HandleFunc("/login", withCORS(func(w http.ResponseWriter, r *http.Request) {
		Login(db, w, r)
	}))
	r.HandleFunc("/refresh-token", withCORS(RefreshTokenHandler))
	// Protected route
	r.HandleFunc("/game/{id}", withCORS(Authenticate(func(w http.ResponseWriter, r *http.Request) {
		UpdateGameBoard(db, w, r)
	})))
	r.HandleFunc("/game/{id}/move", withCORS(Authenticate(func(w http.ResponseWriter, r *http.Request) {
		MakeMove(db, w, r)
	})))
	r.HandleFunc("/verify-token", withCORS(Authenticate(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
	})))
	r.HandleFunc("/search-game", withCORS(Authenticate(func(w http.ResponseWriter, r *http.Request) {
		SearchGame(db, w, r)
	})))
	r.HandleFunc("/leave-queue", withCORS(Authenticate(func(w http.ResponseWriter, r *http.Request) {
		LeaveQueue(db, w, r)
	})))

	http.ListenAndServe(":8080", r)
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("Received %s request for %s\n", r.Method, r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func SearchGame(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	select {
	case opponent := <-waitingPlayers:
		// Found an opponent, create a new game
		game, err := createNewGame(db, username, opponent)
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
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
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

	// Inform the client that they've been removed from the queue
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed from queue"})
}

func UpdateGameBoard(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameId := vars["id"]
	gameIdInt, _ := strconv.ParseInt(gameId, 0, 64)

	board := GetBoard(db, w, r)
	if board == [3][3]string{} {
		board = [3][3]string{
			{"", "", ""},
			{"", "", ""},
			{"", "", ""},
		}
	}
	victory, _ := CheckVictory(board)
	turn, _ := GetGameTurn(db, gameIdInt)
	var game Game
	if victory {
		game = Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   turn,
			Status: StatusTerminated,
		}
		UpdateGameStatus(db, gameIdInt, StatusTerminated)
	} else {

		game = Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   turn,
			Status: StatusStarted,
		}
	}

	json.NewEncoder(w).Encode(game)
}

func GetGameTurn(db *sql.DB, gameId int64) (string, error) {
	var turn string
	err := db.QueryRow("SELECT turn FROM games WHERE id = $1", gameId).Scan(&turn)
	if err != nil {
		return "", err
	}
	return turn, nil
}

func UpdateGameTurn(db *sql.DB, gameId int64) error {
	_, err := db.Exec("UPDATE games SET turn = CASE WHEN turn = 'X' THEN 'O' ELSE 'X' END WHERE id = $1", gameId)
	if err != nil {
		return err
	}
	return nil
}

func getPlayerByName(db *sql.DB, player_name string) (Player, error) {
	var player Player
	err := db.QueryRow("SELECT id, name FROM players WHERE name = $1", player_name).Scan(&player.ID, &player.Name)
	if err != nil {
		return Player{}, err
	}

	return player, nil
}

func createNewGame(db *sql.DB, player_x_name string, player_o_name string) (Game, error) {
	playerO, _ := getPlayerByName(db, player_o_name)
	playerX, _ := getPlayerByName(db, player_x_name)

	res, err := db.Query("INSERT INTO games (status, player_x_id, player_o_id) VALUES ($1, $2, $3) RETURNING id", statusName[StatusStarted], playerX.ID, playerO.ID)
	if err != nil {
		return Game{}, err
	}

	var id int64
	for res.Next() {
		err = res.Scan(&id)
		if err != nil {
			return Game{}, err
		}
	}

	game := Game{
		ID: id,
		Board: [3][3]string{
			{"", "", ""},
			{"", "", ""},
			{"", "", ""},
		},
		Turn:      "X",
		Status:    StatusStarted,
		PlayerXId: playerX.ID,
		PlayerOId: playerO.ID,
	}

	return game, nil
}

func MakeMove(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameId := vars["id"]
	gameIdInt, _ := strconv.ParseInt(gameId, 0, 64)

	move := Move{}
	err := json.NewDecoder(r.Body).Decode(&move)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	player, _ := getPlayerByName(db, move.Username)

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
	victory, _ := CheckVictory(board)

	var game Game
	if victory {
		game = Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   gameDb.Turn,
			Status: StatusTerminated,
		}
		UpdateGameStatus(db, gameIdInt, StatusTerminated)
	} else {
		var turn string

		if gameDb.Turn == "X" {
			turn = "O"
		} else {
			turn = "X"
		}

		game = Game{
			ID:     gameIdInt,
			Board:  board,
			Turn:   turn,
			Status: StatusInProgress,
		}
		UpdateGameTurn(db, gameIdInt)
		// TODO don't really want to update everytime
		UpdateGameStatus(db, gameIdInt, StatusInProgress)
	}

	json.NewEncoder(w).Encode(game)
}

func GetBoard(db *sql.DB, w http.ResponseWriter, r *http.Request) [3][3]string {
	vars := mux.Vars(r)
	gameId := vars["id"]
	var moves []Move
	res, err := db.Query("SELECT x, y, player FROM moves WHERE game_id = $1", gameId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Moves not found", http.StatusNotFound)
		return [3][3]string{}
	}
	for res.Next() {
		var move Move
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

// TODO checkDraw
func CheckVictory(board [3][3]string) (bool, string) {
	// Check rows
	for _, row := range board {
		if row[0] != "" && row[0] == row[1] && row[1] == row[2] {
			return true, row[0]
		}
	}
	// Check columns
	for col := 0; col < 3; col++ {
		if board[0][col] != "" && board[0][col] == board[1][col] && board[1][col] == board[2][col] {
			return true, board[0][col]
		}
	}
	// Check diagonals
	if board[0][0] != "" && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return true, board[0][0]
	}
	if board[0][2] != "" && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return true, board[0][2]
	}
	return false, ""
}

func UpdateGameStatus(db *sql.DB, gameId int64, status GameStatus) error {
	_, err := db.Exec("UPDATE games SET status = $1, updated_at = $2 WHERE id = $3", statusName[status], time.Now(), gameId)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to update game status: %w", err)
	}
	return nil
}

func GetGame(db *sql.DB, gameId int64) (Game, error) {
	var game Game
	err := db.QueryRow("SELECT id, turn, player_x_id, player_o_id FROM games WHERE id = $1", gameId).Scan(&game.ID, &game.Turn, &game.PlayerXId, &game.PlayerOId)
	if err != nil {
		return Game{}, err
	}
	return game, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Register(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var player Player
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(player.Password)
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
	var player Player

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

	if !checkPasswordHash(player.Password, storedHash) {
		http.Error(w, "Invalid name or password", http.StatusUnauthorized)
		return
	}

	// Generate access and refresh tokens
	accessToken, err := GenerateAccessToken(player.Name)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := GenerateRefreshToken(player.Name)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Respond with both tokens
	response := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GenerateAccessToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Minute * 1).Unix(), // Short-lived access token
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(viper.GetString("JWT_SECRET_KEY"))
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Generate a new refresh token
func GenerateRefreshToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 5).Unix(), // Long-lived refresh token
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(viper.GetString("JWT_SECRET_KEY"))
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Validate JWT token
func ValidateToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(viper.GetString("JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// Authenticate middleware
func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		token, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", token.Claims.(jwt.MapClaims)["username"])
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Handle token refresh
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var request map[string]string
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	refreshTokenStr, ok := request["refresh_token"]
	if !ok {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	token, err := ValidateToken(refreshTokenStr)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	username := token.Claims.(jwt.MapClaims)["username"].(string)

	// Generate new tokens
	accessToken, err := GenerateAccessToken(username)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Respond with new tokens
	response := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
