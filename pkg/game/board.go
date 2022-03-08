// Package game provides types for playing a game of Tic-tac-toe and maintaining board state.
package game

import (
	"fmt"

	"github.com/jeremyt135/tictactoe/pkg/tokens"
)

const (
	numToWin         int = 3 // number of cells to occupy in a row to win
	numRows, numCols int = 3, 3
)

// Board contains grid data for a tic-tac-toe game.
type Board struct {
	grid         [numRows][numCols]string
	numTokens    int
	winningToken string
}

// New creates a pointer to a properly initialized Board.
func New() (board *Board) {
	board = &Board{winningToken: tokens.Empty}
	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			board.grid[i][j] = tokens.Empty
		}
	}
	return
}

func (board *Board) String() string {
	output := ""
	for i := 0; i < numRows; i++ {
		output += fmt.Sprint(board.grid[i])
		if i+1 < numRows {
			output += "\n"
		}
	}
	if board.HasWinner() {
		output += fmt.Sprint("\nwinner:", board.winningToken)
	}
	return output
}

// IsFull returns true if the Board is full (every cell has a nonempty token).
func (board *Board) IsFull() bool {
	return board.numTokens == numRows*numCols
}

// HasWinner returns true if the Board has a winner
func (board *Board) HasWinner() bool {
	return board.winningToken != tokens.Empty
}

// WinningToken returns the winning token value.
func (board *Board) WinningToken() string {
	return board.winningToken
}

// At returns the value of the Board at (row,col).
//
// RangeError is returned when the given (row, col) pair is out of range.
func (board *Board) At(row, col int) (string, error) {
	if !inRange(row, col) {
		return "", &RangeError{row, col}
	}
	return board.grid[row][col], nil
}

// Put updates the given cell of the Board to the given string value. Successive
// calls to Put, or calls on a full or winning Board, will have no effect.
//
// Returns true if Put changed the Board successfully.
//
// RangeError is returned when the given (row, col) pair is out of range.
// TokenError is returned if the given token is not a valid token.
func (board *Board) Put(token string, row, col int) (bool, error) {
	if !inRange(row, col) {
		return false, &RangeError{row, col}
	}
	if !isToken(token) {
		return false, &TokenError{token}
	}
	if board.IsFull() || board.HasWinner() || board.grid[row][col] != tokens.Empty {
		return false, nil
	}

	board.grid[row][col] = token
	board.numTokens++
	if board.checkIfWinner(token, row, col) {
		board.winningToken = token
	}

	return true, nil
}

func inRange(row, col int) bool {
	return row >= 0 && row < numRows && col >= 0 && col < numCols
}

func isToken(value string) bool {
	return value == tokens.X || value == tokens.O
}

func (board *Board) sameAt(token string, row, col int) bool {
	return board.grid[row][col] == token
}

func (board *Board) checkHorizontal(token string, row, col int) bool {
	count := 1
	for r := row - 1; r >= 0; r-- {
		if board.sameAt(token, r, col) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}

	for r := row + 1; r < numRows; r++ {
		if board.sameAt(token, r, col) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}
	return false
}

func (board *Board) checkVertical(token string, row, col int) bool {
	count := 1
	for c := col - 1; c >= 0; c-- {
		if board.sameAt(token, row, c) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}

	for c := col + 1; c < numCols; c++ {
		if board.sameAt(token, row, c) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}
	return false
}

func (board *Board) checkDiagonals(token string, row, col int) bool {
	// check NE to SW diagonal
	count := 1
	for r, c := row-1, col-1; r >= 0 && c >= 0; r, c = r-1, c-1 {
		if board.sameAt(token, r, c) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}

	for r, c := row+1, col+1; r < numRows && c < numCols; r, c = r+1, c+1 {
		if board.sameAt(token, r, c) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}

	// check NW to SE diagonal
	count = 1
	for r, c := row+1, col-1; r < numRows && c >= 0; r, c = r+1, c-1 {
		if board.sameAt(token, r, c) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}

	for r, c := row-1, col+1; r >= 0 && c < numCols; r, c = r-1, c+1 {
		if board.sameAt(token, r, c) {
			count++
		} else {
			break
		}
	}
	if count >= numToWin {
		return true
	}
	return false
}

func (board *Board) checkIfWinner(token string, row, col int) bool {
	return board.checkHorizontal(token, row, col) ||
		board.checkVertical(token, row, col) ||
		board.checkDiagonals(token, row, col)
}

// RangeError occurs if a Board operation is performed with
// an invalid row or col value.
type RangeError struct {
	Row, Col int
}

func (rangeErr *RangeError) Error() string {
	return fmt.Sprintf("board index out of range: %v, %v", rangeErr.Row, rangeErr.Col)
}

// TokenError occurs when a token is not valid.
type TokenError struct {
	Value string
}

func (tokenErr *TokenError) Error() string {
	return fmt.Sprintf("token is invalid: %v, must be one of %v, %v", tokenErr.Value, tokens.X, tokens.O)
}
