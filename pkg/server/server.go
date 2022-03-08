// Package server provides a server type for playing Tic-Tac-Toe over the internet.
package server

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jeremyt135/tictactoe/pkg/logger"
	"github.com/jeremyt135/tictactoe/pkg/protocol"
	"github.com/jeremyt135/tictactoe/pkg/server/internal/lobby"
	"github.com/jeremyt135/tictactoe/pkg/server/internal/player"
)

// Conn is a generic wrapper for an incoming connection.
type Conn interface {
	// Send returns a channel to send the Conn data.
	Send() chan<- string

	// Receive returns a channel to retrieve data from the Conn.
	Receive() <-chan string

	// Close hangs up on the connection.
	Close() error
}

// Listener supplies a Server with incoming connections.
type Listener interface {
	// PollAccept tells the Listener to start accepting connections.
	PollAccept() error

	// Connections provides a channel of incoming connections.
	Connections() <-chan Conn

	// Close stops the Listener accepting connections and closes resources.
	Close() error
}

// Server runs a tic-tac-toe server that accepts bare TCP connections.
type Server struct {
	listener Listener
	port     int
	lobbies  []*lobby.Lobby
	mux      sync.Mutex
	logger   logger.Logger
}

func validateOptions(opt *Options) error {
	if opt.NumLobbies <= 0 {
		return errors.New("lobbies must be positive")
	}
	return nil
}

// NewServer configures a Server, so it will be ready to Listen on a port.
//
// Options may be passed to configure operations such as logging.
// If nil, default options will be used.
func NewServer(opt *Options) (*Server, error) {
	s := &Server{}

	if opt == nil {
		opt = DefaultOptions()
	} else {
		if err := validateOptions(opt); err != nil {
			return nil, fmt.Errorf("could not create Server with opt: %w", err)
		}
	}

	if opt.Logger == nil {
		s.logger = logger.NoOpLogger()
	} else {
		s.logger = opt.Logger
	}

	s.lobbies = make([]*lobby.Lobby, opt.NumLobbies)
	for i := 0; i < len(s.lobbies); i++ {
		s.lobbies[i] = lobby.New()
	}

	return s, nil
}

// Close shuts down a Server.
func (s *Server) Close() {
	defer s.listener.Close()
	s.logger.Info("shutting down")
}

func (s *Server) Serve(l Listener) error {
	if l == nil {
		return errors.New("can not create Server with nil listener")
	}
	s.listener = l
	go s.pollConnections()
	return s.listener.PollAccept()
}

func (s *Server) pollConnections() {
	defer s.Close()

	s.logger.Info("server waiting for connections")

	conns := s.listener.Connections()
	for {
		c, ok := <-conns
		if !ok {
			break
		}
		go s.handleConnection(c)
	}
}

func (s *Server) handleConnection(c Conn) {
	s.logger.Info("received connection")

	// Wrap in utility NetConn type
	p := player.New(c.Send(), c.Receive())

	if ok := confirmConnection(p); ok {
		l := s.nextAvailableLobby()
		if l == nil {
			s.logger.Info("received connection but could not find an open lobby")
			c.Close()
			return
		}
		if err := l.AddPlayer(p); err != nil {
			s.logger.Error("error adding a client to lobby: ", err)
			c.Close()
			return
		}
		s.logger.Info("added a client to lobby ", l.ID())
	} else {
		s.logger.Error("received invalid response or could not write to client")
		c.Close()
		return
	}
}

func confirmConnection(p *player.Player) bool {
	// Perform handshake - both sides must send protocol.Greeting
	p.Send <- protocol.Greeting

	if res, ok := <-p.Receive; !ok || res != protocol.Greeting {
		return false
	}

	return true
}

func (s *Server) nextAvailableLobby() (openLobby *lobby.Lobby) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, l := range s.lobbies {
		if l.IsAvailable() {
			openLobby = l
			return
		}
	}
	return nil
}
