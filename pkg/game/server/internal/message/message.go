package message

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/tokens"
)

// Message is an interface for strings of data that will be sent
// to/from the game server.
type Message interface {
	// Op returns the string for the type of Message.
	Op() string

	String() string
}

// PlayerToken is a Message that contains a player's Token
type PlayerToken struct {
	Token string
}

// Op returns "PLAYER" as a PlayerToken Message's type of operation.
func (msg PlayerToken) Op() string {
	return "PLAYER"
}

func (msg PlayerToken) String() string {
	return fmt.Sprintln(msg.Op(), msg.Token)
}

// TurnInfo records data for one player's turn.
type TurnInfo struct {
	// The player that moved
	Token    string
	Row, Col int
}

// Op returns "TURN" as a TurnInfo Message's type of operation.
func (msg TurnInfo) Op() string {
	return "TURN"
}

func (msg TurnInfo) String() string {
	return fmt.Sprintln(msg.Op(), msg.Token, msg.Row, msg.Col)
}

// ParseTurnInfo attempts to parse a TurnInfo from a given string.
// Returns a Message of TurnInfo type if the string was valid.
// Returns nil and an error if the string was not formatted properly.
func ParseTurnInfo(s string) (Message, error) {
	s = strings.TrimSuffix(s, "\n")
	const numFields = 4
	turn := strings.SplitN(s, " ", numFields)
	msg := TurnInfo{}
	// check if it at least somewhat matches "TURN TOKEN VAL VAL"
	if len(turn) == numFields && turn[0] == msg.Op() {
		// validate token
		token := turn[1]
		validToken := token == tokens.X || token == tokens.O
		// try to parse last 2 values as int
		row, rowErr := strconv.Atoi(turn[2])
		col, colErr := strconv.Atoi(turn[3])
		if validToken && rowErr == nil && colErr == nil {
			msg.Token, msg.Row, msg.Col = token, row, col
			return msg, nil
		}
	}

	return nil, fmt.Errorf("could not parse string %v into turnInfo", s)
}

// TurnNotif is a message to send when a player should move.
type TurnNotif struct {
	Token string
}

// Op returns "MOVE" as a TurnNotif Message's type of operation.
func (msg TurnNotif) Op() string {
	return "MOVE"
}

func (msg TurnNotif) String() string {
	return fmt.Sprintln(msg.Op(), msg.Token)
}

// GameOver is a message indicating that the game has ended. It
// records the winning token.
type GameOver struct {
	WinningToken string
}

// Op returns "WINNER" as a GameOver Message's type of operation.
func (msg GameOver) Op() string {
	return "WINNER"
}

func (msg GameOver) String() string {
	return fmt.Sprintln(msg.Op(), msg.WinningToken)
}
