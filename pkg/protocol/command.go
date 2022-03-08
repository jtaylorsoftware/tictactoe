// Package protocol defines the structure for client-server communication when playing tic-tac-toe.
package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jeremyt135/tictactoe/pkg/tokens"
)

// Command is an interface for strings of data that will be sent
// to/from the game server.
type Command interface {
	// Op returns the string for the type of Command.
	Op() string

	String() string
}

// Greeting is the required message to start a Tic-Tac-Toe match. It must be the first
// message.
const Greeting = "TICTACTOE\n"

// InternalError is an error response that occurs when an error occurs that is not the responsibility
// of the player to fix.
var InternalError = errors.New("INTERNAL ERROR\n")

// TokenError occurs when a player attempts to take a move without using their assigned token.
var TokenError = errors.New("INVALID TOKEN\n")

// RangeError indicates a player's move does not exist within the confines of the board.
var RangeError = errors.New("INVALID RANGE\n")

// SpaceTakenError is an error response indicating that the player's desired space on the board
// already has a token in it.
var SpaceTakenError = errors.New("INVALID FULL\n")

// ParseError indicates that a player's move does not
// have the expected format and could not be parsed.
type ParseError struct {
	failedStr string
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("could not parse string %v into turnInfo", pe.failedStr)
}

func (pe *ParseError) AsResponse() string {
	return "INVALID FORMAT\n"
}

// PlayerToken is a Command that contains a player's Token
type PlayerToken struct {
	Token string
}

// Op returns "PLAYER" as a PlayerToken Command's type of operation.
func (pt PlayerToken) Op() string {
	return "PLAYER"
}

func (pt PlayerToken) String() string {
	return fmt.Sprintln(pt.Op(), pt.Token)
}

// TurnInfo records data for one player's turn.
type TurnInfo struct {
	// The player that moved
	Token    string
	Row, Col int
}

// Op returns "TURN" as a TurnInfo Command's type of operation.
func (ti TurnInfo) Op() string {
	return "TURN"
}

func (ti TurnInfo) String() string {
	return fmt.Sprintln(ti.Op(), ti.Token, ti.Row, ti.Col)
}

// ParseTurnInfo attempts to parse a TurnInfo from a given string.
// Returns a Command of TurnInfo type if the string was valid.
// Returns nil and an error if the string was not formatted properly.
func ParseTurnInfo(s string) (Command, error) {
	s = strings.TrimSuffix(s, "\n")
	const numFields = 4
	turn := strings.SplitN(s, " ", numFields)
	cmd := TurnInfo{}
	// check if it at least somewhat matches "TURN TOKEN VAL VAL"
	if len(turn) == numFields && turn[0] == cmd.Op() {
		// validate token
		token := turn[1]
		validToken := token == tokens.X || token == tokens.O
		// try to parse last 2 values as int
		row, rowErr := strconv.Atoi(turn[2])
		col, colErr := strconv.Atoi(turn[3])
		if validToken && rowErr == nil && colErr == nil {
			cmd.Token, cmd.Row, cmd.Col = token, row, col
			return cmd, nil
		}
	}

	return nil, &ParseError{failedStr: s}
}

// TurnNotif is a command to send when a player should move.
type TurnNotif struct {
	Token string
}

// Op returns "MOVE" as a TurnNotif Command's type of operation.
func (tn TurnNotif) Op() string {
	return "MOVE"
}

func (tn TurnNotif) String() string {
	return fmt.Sprintln(tn.Op(), tn.Token)
}

// GameOver is a command indicating that the game has ended. It
// records the winning token.
type GameOver struct {
	WinningToken string
}

// Op returns "WINNER" as a GameOver Command's type of operation.
func (g GameOver) Op() string {
	return "WINNER"
}

func (g GameOver) String() string {
	return fmt.Sprintln(g.Op(), g.WinningToken)
}
