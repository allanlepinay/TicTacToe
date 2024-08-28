package types

type Game struct {
	ID        int64        `json:"id"`
	Board     [3][3]string `json:"board"`
	Turn      string       `json:"turn"`
	Status    int64        `json:"status"`
	PlayerXId int64        `json:"player_x_id"`
	PlayerOId int64        `json:"player_o_id"`
}

type Move struct {
	WebsocketMessage
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Turn string `json:"turn"`
}

type WebsocketMessage struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Username string `json:"username"`
	GameId   int64  `json:"gameId"`
}

type PlayerProfile struct {
	Type string `json:"type"`
	Player
	Games []Game `json:"games"`
}

type Player struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Password      string `json:"password"`
	Wins          int64  `json:"wins"`
	Loses         int64  `json:"loses"`
	Draw          int64  `json:"draw"`
	WebsocketConn string `json:"websocket_conn"`
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

var StatusName = map[GameStatus]string{
	StatusStarted:    "Started",
	StatusInProgress: "In-Progress",
	StatusTerminated: "Terminated",
}
