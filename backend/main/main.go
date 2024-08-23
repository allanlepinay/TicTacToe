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
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var leaveQueueMessage struct {
	Username string `json:"username"`
}

var waitingPlayers = make(chan string, 100)
var mutex = &sync.Mutex{}

type Client struct {
	conn     *websocket.Conn
	username string
	gameID   int64
}

var clients = make(map[*websocket.Conn]*Client)

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
	r.HandleFunc("/verify-token", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
	})))
	r.HandleFunc("/leave-queue", auth.WithCORS(auth.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		LeaveQueue(db, w, r)
	})))
	r.HandleFunc("/ws", auth.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		_, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		handleWebSocket(db, w, r)
	}))

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

	board := database.GetBoard(db, types.Move{})
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

func handleWebSocket(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := &Client{
		conn:     conn,
		username: "",
		gameID:   -1,
	}
	if _, ok := clients[conn]; !ok {
		clients[conn] = client
	}

	for {
		// Read client message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}

		var message types.Move
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Println("Error unmarshaling message:", err)
			continue
		}

		switch message.Type {
		case "JoinQueue":
			clients[conn].username = message.Username
			client = clients[conn]
			if client.username == "" {
				client.conn.WriteJSON(types.WebsocketMessage{
					Type:     "error",
					Message:  "No username found",
					Username: "",
					GameId:   -1})
				return
			}
			database.StorePlayerWebsocketConn(db, message.Username, conn)

			waitingPlayers <- client.username
			select {
			case player1 := <-waitingPlayers:
				select {
				case player2 := <-waitingPlayers:
					if player1 != player2 {
						game, err := database.CreateNewGame(db, player1, player2)
						if err != nil {
							fmt.Println("Failed to create game:", err)
							continue
						}
						// Update clients with new game information (only concerned clients)
						for _, c := range clients {
							if c.username == player1 || c.username == player2 {
								c.gameID = game.ID
								c.conn.WriteJSON(types.WebsocketMessage{
									Type:     "gameCreated",
									Message:  "",
									Username: "",
									GameId:   game.ID})
							}
						}
					} else {
						// Can't play with oneself
						waitingPlayers <- player1
						client.conn.WriteJSON(types.WebsocketMessage{
							Type:     "waiting",
							Message:  "Can't play with oneself",
							Username: "",
							GameId:   -1})
					}
				default:
					// No second player yet
					waitingPlayers <- player1
					client.conn.WriteJSON(types.WebsocketMessage{
						Type:     "waiting",
						Message:  "",
						Username: "",
						GameId:   -1})
				}
			default:
				client.conn.WriteJSON(types.WebsocketMessage{
					Type:     "waiting",
					Message:  "",
					Username: "",
					GameId:   -1})
			}
		case "ping":
			client.conn.WriteJSON(types.WebsocketMessage{
				Type:     "message",
				Message:  "pong",
				Username: "",
				GameId:   0})
		case "move":
			database.StorePlayerWebsocketConn(db, message.Username, conn)

			if message.GameId != -1 && message.GameId != 0 {
				var move types.Move
				moveData, err := json.Marshal(message)
				if err != nil {
					fmt.Println("Error marshalling move:", err)
					continue
				}
				err = json.Unmarshal(moveData, &move)
				if err != nil {
					fmt.Println("Error unmarshalling move:", err)
					continue
				}
				game := database.MakeMove(db, move)

				players, _ := database.GetPlayersByGameId(db, move.GameId)

				// Send game to players (only concerned players)
				for _, player := range players {
					for _, client := range clients {
						if player.WebsocketConn == client.conn.RemoteAddr().String() {
							gameJSON, _ := json.Marshal(game)
							client.conn.WriteJSON(types.WebsocketMessage{
								Type:     "move",
								Message:  string(gameJSON),
								Username: client.username,
								GameId:   client.gameID})
						}
					}
				}
			}
		}
	}
}
