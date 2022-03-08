package lobby

import (
	"errors"
	"fmt"
	"github.com/jeremyt135/tictactoe/pkg/logger"
	"github.com/jeremyt135/tictactoe/pkg/server/internal/config"
	"github.com/jeremyt135/tictactoe/pkg/server/internal/player"

	"github.com/jeremyt135/tictactoe/pkg/game"
	"github.com/jeremyt135/tictactoe/pkg/protocol"
	"github.com/jeremyt135/tictactoe/pkg/tokens"
)

// Lobby records an ongoing game and its players.
type Lobby struct {
	board         *game.Board
	players       player.Array
	logger        logger.Logger
	id            int
	playing       bool
	currentPlayer int
}

var nextLobbyID = 0

// New constructs a new game Lobby.
func New() (lobby *Lobby) {
	lobby = &Lobby{
		board:         game.New(),
		players:       player.NewFixedArray(),
		id:            nextLobbyID,
		logger:        logger.NoOpLogger(),
		playing:       false,
		currentPlayer: -1,
	}

	nextLobbyID++
	return
}

// UseLogger changes the logger being used by the Lobby.
func (l *Lobby) UseLogger(logger logger.Logger) *Lobby {
	l.logger = logger
	return l
}

// IsFull returns true if the Lobby is full and cannot accept more players.
func (l *Lobby) IsFull() bool {
	return l.players.IsFull()
}

// IsPlaying returns true if the Lobby has a game in progress and cannot accept players.
func (l *Lobby) IsPlaying() bool {
	return l.playing
}

// IsAvailable returns true if the Lobby is available and can add players
func (l *Lobby) IsAvailable() bool {
	return !(l.IsFull() || l.IsPlaying())
}

// ID returns the Lobby's ID value
func (l *Lobby) ID() int {
	return l.id
}

// AddPlayer adds a player to the game Lobby.
//
// If the Lobby is full, returns an error.
func (l *Lobby) AddPlayer(p *player.Player) error {
	if !l.IsAvailable() {
		return errors.New("lobby is not available")
	}
	ind, err := l.players.Add(p)
	if err != nil {
		return fmt.Errorf("could not add p: %w", err)
	}

	p.ID = ind
	p.Token = tokens.FromIndex(ind)

	if l.IsFull() {
		go l.play()
	}
	return nil
}

func (l *Lobby) stop() {
	l.playing = false
	l.logger.Info("lobby ", l.id, " stopping...")
	// TODO - gracefully shut down & tell any connected players l shut down
	// maybe add option to rematch
	for i := 0; i < l.players.Size(); i++ {
		p := l.players.At(i)
		l.removePlayer(p, "lobby stopping")
	}
	l.logger.Info("lobby ", l.id, " resetting...")
	l.reset()
}

func (l *Lobby) reset() {
	l.board = game.New()
	l.currentPlayer = -1
}

// maxTurnAttempts is the number of tries a player can have at making
// a valid turn before they are disconnected.
const maxTurnAttempts = 3

func (l *Lobby) play() {
	l.playing = true
	l.logger.Info("lobby ", l.id, " playing")

	l.identifyPlayers()

	// continue until game is over
	for !l.board.HasWinner() && l.playing {
		p := l.nextPlayer()

		// First, notify player that it's their turn
		msg := protocol.TurnNotif{Token: p.Token}
		p.Send <- msg.String()

		// Wait for and validate their response. Give them a few tries.
		var attempts = 0
		for ; attempts < maxTurnAttempts; attempts++ {
			s, ok := <-p.Receive
			if !ok {
				l.logger.Error("lobby ", l.id, " could not receive move from ", p.Token, ": channel closed")
				l.removePlayer(p, "disconnected")
				l.stop()
				return
			}

			msg, err := protocol.ParseTurnInfo(s)
			var parseError *protocol.ParseError
			if err != nil {
				l.logger.Info("lobby ", l.id, " error in move from ", p.Token, ": ", err)
				if errors.As(err, &parseError) {
					p.Send <- parseError.AsResponse()
				} else {
					p.Send <- protocol.InternalError.Error()
				}
				continue
			}

			l.logger.Info("lobby ", l.id, " ", msg)

			turn := msg.(protocol.TurnInfo)
			if turn.Token != p.Token {
				// p trying to move as opponent
				l.logger.Info("lobby ", l.id, " turn from", p.Token, " did not match turn token ", turn.Token)

				p.Send <- protocol.TokenError.Error()
				continue
			}

			// attempt move
			turnOk, err := l.board.Put(turn.Token, turn.Row, turn.Col)
			if err != nil {
				l.logger.Info("lobby ", l.id, " : ", err)
				switch err.(type) {
				case *game.TokenError:
					p.Send <- protocol.TokenError.Error()
				case *game.RangeError:
					p.Send <- protocol.RangeError.Error()
				}
				continue
			}
			if !turnOk {
				l.logger.Info("lobby ", l.id, " turn from", turn.Token, " did not change board")
				p.Send <- protocol.SpaceTakenError.Error()
				continue
			} else {
				l.notifyTurnTaken(turn)
				break
			}
		}

		if attempts == maxTurnAttempts {
			// assume p was trying to cheat and remove them
			l.removePlayer(p, "too many invalid moves")
			// stop game and return, no winner
			l.stop()
			return
		}
	}
	// notify players of winner and shut down
	l.notifyWinner()
	l.stop()
}

func (l *Lobby) identifyPlayers() {
	// Send players the token that they have to use.
	for i := 0; i < l.players.Size(); i++ {
		p := l.players.At(i)
		msg := protocol.PlayerToken{Token: p.Token}
		p.Send <- msg.String()
		//if err != nil {
		//	l.removePlayer(p, "could not write identity")
		//	l.stop()
		//}
	}
}

func (l *Lobby) notifyWinner() {
	// Tell all players that there is a winner.
	for i := 0; i < l.players.Size(); i++ {
		p := l.players.At(i)
		msg := protocol.GameOver{WinningToken: l.board.WinningToken()}
		p.Send <- msg.String()
	}
}

func (l *Lobby) notifyTurnTaken(turn protocol.TurnInfo) {
	// Tell all players that a turn was taken, except for the one who took turn.
	for i := 0; i < l.players.Size(); i++ {
		p := l.players.At(i)
		if p.Token != turn.Token {
			p.Send <- turn.String()
		}
	}
}

func (l *Lobby) nextPlayer() *player.Player {
	nextID := (l.currentPlayer + 1) % config.MaxPlayers
	l.currentPlayer = nextID
	return l.players.At(nextID)
}

func (l *Lobby) removePlayer(p *player.Player, why string) {
	if p == nil {
		return
	}
	l.logger.Info("lobby ", l.id, " removing player ", p.ID, ", ", why, "\n")
	p.Send <- "REMOVED\n"
	close(p.Send)
	l.players.Remove(p.ID)
}
