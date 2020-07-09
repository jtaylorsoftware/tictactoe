package lobby

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jeremyt135/tictactoe/pkg/board"
	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/config"
	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/message"
	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/player"
	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/tokens"
)

// Lobby records an ongoing game and its players.
type Lobby struct {
	board         *board.Board
	players       player.Array
	logger        *log.Logger
	logPrefix     string
	id            int
	playing       bool
	currentPlayer int
}

var nextLobbyID int = 0

// New constructs a new game Lobby.
func New(logger *log.Logger) (lobby *Lobby) {
	lobby = &Lobby{board.New(),
		player.NewArray(),
		logger,
		fmt.Sprint("lobby ", nextLobbyID, ": "),
		nextLobbyID,
		false, -1}
	nextLobbyID++
	return
}

// IsFull returns true if the Lobby is full and cannot accept more players.
func (lobby *Lobby) IsFull() bool {
	return lobby.players.IsFull()
}

// IsPlaying returns true if the Lobby has a game in progress and cannot accept players.
func (lobby *Lobby) IsPlaying() bool {
	return lobby.playing
}

// IsAvailable returns true if the Lobby is available and can add players
func (lobby *Lobby) IsAvailable() bool {
	return !(lobby.IsFull() || lobby.IsPlaying())
}

// ID returns the Lobby's ID value
func (lobby *Lobby) ID() int {
	return lobby.id
}

// AddPlayer adds a player to the game Lobby.
//
// If the Lobby is full, returns an error.
func (lobby *Lobby) AddPlayer(conn player.Conn) error {
	if !lobby.IsAvailable() {
		return errors.New("lobby is not available")
	}
	player := player.New(conn)
	ind, err := lobby.players.Add(player)
	if err != nil {
		return fmt.Errorf("could not add player: %w", err)
	}

	player.ID = ind
	player.Token = tokens.FromIndex(ind)

	if lobby.IsFull() {
		go lobby.play()
	}
	return nil
}

func (lobby *Lobby) stop() {
	lobby.playing = false
	lobby.logln("stopping")
	// TODO - gracefully shut down & tell any connected players lobby shut down
}

func (lobby *Lobby) play() {
	lobby.playing = true
	lobby.logln("playing")

	lobby.identifyPlayers()

	// continue until game is over
	for !lobby.board.HasWinner() && lobby.playing {
		// tell next player to move
		player := lobby.nextPlayer()
		msg := message.TurnNotif{Token: player.Token}
		// TODO - ensure this message reaches player
		if n, err := player.Conn.WriteString(msg.String()); n != len(msg.String()) || err != nil {
			lobby.removePlayer(player, "could not write turn notification")
			lobby.stop()
			return
		}

		// try to get a valid move from player, give them a few tries
		var attempts = 0
		for ; attempts < config.MaxTurnAttempts; attempts++ {
			s, err := player.Conn.ReadString()
			if err != nil {
				lobby.logln(err)
				continue
			}
			msg, err := message.ParseTurnInfo(s)
			if err != nil {
				lobby.logln(err)
				continue
			}
			lobby.log(msg)

			turn := msg.(message.TurnInfo)
			if turn.Token != player.Token {
				// player trying to move as opponent
				lobby.logln("turn from", player.Token, "did not match turn token", turn.Token)

				// TODO - check write error
				player.Conn.WriteString("INVALID TOKEN\n")
				continue
			}

			// attempt move
			turnOk, err := lobby.board.Put(turn.Token, turn.Row, turn.Col)
			if err != nil {
				lobby.logln(err)
				switch err.(type) {
				case *board.TokenError:
					player.Conn.WriteString("INVALID TOKEN\n")
				case *board.RangeError:
					player.Conn.WriteString("INVALID RANGE\n")
				}
				continue
			}
			if !turnOk {
				lobby.logln("turn from", turn.Token, "did not change board")
				player.Conn.WriteString("INVALID FULL\n")
				continue
			} else {
				lobby.notifyTurnTaken(turn)
				break
			}
		}

		if attempts == config.MaxTurnAttempts {
			// assume player was trying to cheat and remove them
			lobby.removePlayer(player, "too many invalid moves")
			lobby.stop()
		}
	}
	// notify players of winner and shut down
	lobby.notifyWinner()
	lobby.stop()
}

func (lobby *Lobby) identifyPlayers() {
	// notify players of their identity
	for i := 0; i < lobby.players.Size(); i++ {
		player := lobby.players.At(i)
		msg := message.PlayerToken{Token: player.Token}
		if n, err := player.Conn.WriteString(msg.String()); n != len(msg.String()) || err != nil {
			lobby.removePlayer(player, "could not write identity")
			lobby.stop()
		}
	}
}

func (lobby *Lobby) notifyWinner() {
	// tell all players that there is a winner
	for i := 0; i < lobby.players.Size(); i++ {
		player := lobby.players.At(i)
		msg := message.GameOver{WinningToken: lobby.board.WinningToken()}
		// TODO - ensure this message reaches player
		player.Conn.WriteString(msg.String())
	}
}

func (lobby *Lobby) notifyTurnTaken(turn message.TurnInfo) {
	// tell all players that a turn was taken, except for the one who took turn
	for i := 0; i < lobby.players.Size(); i++ {
		player := lobby.players.At(i)
		if player.Token != turn.Token {
			// TODO - ensure this message reaches player
			player.Conn.WriteString(turn.String())
		}
	}
}

func (lobby *Lobby) nextPlayer() *player.Player {
	nextID := (lobby.currentPlayer + 1) % config.MaxPlayers
	lobby.currentPlayer = nextID
	return lobby.players.At(nextID)
}

func (lobby *Lobby) removePlayer(p *player.Player, why string) {
	lobby.logln("removing player ", p.ID, ", ", why)
	lobby.players.Remove(p.ID)
	p.Conn.Close()
}

func (lobby *Lobby) log(args ...interface{}) {
	lobby.logger.Print(lobby.logPrefix, fmt.Sprint(args...))
}

func (lobby *Lobby) logln(args ...interface{}) {
	lobby.logger.Print(lobby.logPrefix, fmt.Sprintln(args...))
}

// minutesFromNow is a utility function returning a time from now
func minutesFromNow(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * time.Minute)
}
