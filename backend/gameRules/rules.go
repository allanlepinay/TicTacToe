package gamerules

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
