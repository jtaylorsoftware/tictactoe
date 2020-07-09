package game

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/lobby"
	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/player"
)

// Server handles connections and tictactoe game lobbies.
type Server struct {
	listener net.Listener
	port     int
	lobbies  []*lobby.Lobby
	mux      sync.Mutex
	logger   *log.Logger
}

// Options hold configuration data for a server.
type Options struct {
	EnableLogging bool
	LogOutput     io.Writer
	NumLobbies    int
}

// DefaultOptions returns a Options set with useful defaults for logging.
//
// These defaults are:
//   EnableLogging = true
//   LogFile       = os.Stdout
// 	 NumLobbies    = 1
func DefaultOptions() *Options {
	return &Options{true, os.Stdout, 2}
}

// NewServer configures a Server so it will be ready to Listen on a port.
//
// Options may be passed to configure operations such as logging.
// If nil, default options will be used.
func NewServer(options *Options) (server *Server) {
	server = &Server{}

	if options == nil {
		options = DefaultOptions()
	}

	logOutput, flags := ioutil.Discard, 0
	if options.EnableLogging {
		logOutput, flags = options.LogOutput, log.LstdFlags|log.Lmsgprefix
	}
	server.logger = log.New(logOutput, "", flags)

	server.lobbies = make([]*lobby.Lobby, options.NumLobbies)
	for i := 0; i < len(server.lobbies); i++ {
		server.lobbies[i] = lobby.New(server.logger)
	}
	return
}

// Close shuts down a Server.
func (server *Server) Close() {
	defer server.listener.Close()
	server.logger.Println("shutting down")
}

// Listen causes a Server to start listening for connections until it is closed or
// an error occurs.
func (server *Server) Listen(port int) {
	server.port = port
	listener, err := net.Listen("tcp4", fmt.Sprint(":", server.port))
	if err != nil {
		server.logger.Fatalln("could not listen on port", server.port)
	}
	server.listener = listener

	server.logger.Println("listening on port", server.port)

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			server.logger.Println("error accepting a connection:", err)
			continue
		}
		server.logger.Println("incoming connection")
		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	server.logger.Println("received connection")

	// Wrap in utility Conn type
	playerConn := player.NewConn(conn)

	if ok := confirmConnection(playerConn); ok {
		lobby := server.nextAvailableLobby()
		if lobby == nil {
			server.logger.Println("received connection but could not find an open lobby")
			conn.Close()
			return
		}
		if err := lobby.AddPlayer(playerConn); err != nil {
			server.logger.Println("error adding a client to lobby", err)
			conn.Close()
			return
		}
		server.logger.Println("added a client to lobby", lobby.ID())
	} else {
		server.logger.Println("received invalid response or could not write to client")
		conn.Close()
		return
	}
}

func confirmConnection(conn player.Conn) bool {
	const greeting = "TICTACTOE\n"

	// Send an opening message to the client
	if n, err := conn.WriteString(greeting); n == 0 || n != len(greeting) || err != nil {
		return false
	}

	// Wait for response and check if it was expected
	if str, err := conn.ReadString(); len(str) == 0 || err != nil {
		return false
	}

	return true
}

func (server *Server) nextAvailableLobby() (openLobby *lobby.Lobby) {
	server.mux.Lock()
	defer server.mux.Unlock()

	for _, lobby := range server.lobbies {
		if lobby.IsAvailable() {
			openLobby = lobby
			return
		}
	}
	return nil
}
