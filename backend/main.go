package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

type Game struct {
    ID      int64   `json:"id"`
    Board   [3][3]string `json:"board"`
    Turn    string   `json:"turn"`
    Status  int64   `json:"status"`
}

type Move struct {
    X      int    `json:"x"`
    Y      int    `json:"y"`
    Player string `json:"player"`
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

func main() {
    // TODO clean
    connStr := "user=postgres password=postgres dbname=tictactoe sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Println("Error connecting to the database:", err)
        return
    }
    defer db.Close()

    r := mux.NewRouter()
    r.HandleFunc("/game", withCORS(func(w http.ResponseWriter, r *http.Request) {
        CreateGame(db, w, r)
    }))
    r.HandleFunc("/game/{id}/move", withCORS(func(w http.ResponseWriter, r *http.Request) {
        MakeMove(db, w, r)
    }))

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

func CreateGame(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    res, err := db.Query("INSERT INTO games (status) VALUES ($1) RETURNING id", statusName[StatusStarted])
    if err != nil {
        fmt.Println(err)
        http.Error(w, "Failed to create game", http.StatusInternalServerError)
        return
    }

    var id int64
    for res.Next() {
        err = res.Scan(&id)
        if err != nil {
            fmt.Println(err)
            http.Error(w, "Failed to scan id", http.StatusInternalServerError)
            return
        }
    }


    game := Game{
        ID: id,
        Board: [3][3]string{
            {"", "", ""},
            {"", "", ""},
            {"", "", ""},
        },
        Turn: "X",
        Status: StatusStarted,
    }

    json.NewEncoder(w).Encode(game)
}

func MakeMove(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    gameId := vars["id"]
    move := Move{}
    err := json.NewDecoder(r.Body).Decode(&move)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("INSERT INTO moves (game_id, player, x, y) VALUES ($1, $2, $3, $4)", gameId, move.Player, move.X, move.Y)
    if err != nil {
        fmt.Println(err)
        http.Error(w, "Failed to make move", http.StatusInternalServerError)
        return
    }

    board := GetBoard(db,w,r)
    gameIdInt, _ := strconv.ParseInt(gameId, 0, 64)
    victory, _ := CheckVictory(board)
    var game Game

    if victory {
        game = Game{
            ID: gameIdInt,
            Board: board,
            Turn: move.Player,
            Status: StatusTerminated,
        }
        UpdateGameStatus(db, gameIdInt, StatusTerminated)
    } else {
        var turn string

        if move.Player == "X" {
            turn = "O"
        } else {
            turn = "X"
        }

        game = Game{
            ID: gameIdInt,
            Board: board,
            Turn: turn,
            Status: StatusInProgress,
        }
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

