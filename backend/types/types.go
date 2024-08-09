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

var StatusName = map[GameStatus]string{
	StatusStarted:    "Started",
	StatusInProgress: "In-Progress",
	StatusTerminated: "Terminated",
}
