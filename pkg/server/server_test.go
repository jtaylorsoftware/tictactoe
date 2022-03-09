package server

import (
	"github.com/jeremyt135/tictactoe/pkg/protocol"
	"testing"
)

type fakeConn struct {
	send    chan string
	receive chan string
	poll    func(fakeConn)
}

func (c fakeConn) Send() chan<- string {
	return c.send
}

func (c fakeConn) Receive() <-chan string {
	return c.receive
}

type fakeListener struct {
	conns []fakeConn
	ch    chan Conn
}

func (l fakeListener) PollAccept() error {
	for _, c := range l.conns {
		l.ch <- c
		if c.poll != nil {
			go c.poll(c)
		}
	}
	close(l.ch)
	return nil
}

func (l fakeListener) Connections() <-chan Conn {
	return l.ch
}

func TestServerConnectionGreeting(t *testing.T) {
	// Create server with a fake listener that gives one fake connection
	s, _ := NewServer(nil)
	l := fakeListener{
		ch: make(chan Conn),
	}

	var msg string
	c := fakeConn{
		send:    make(chan string),
		receive: make(chan string),
		poll: func(conn fakeConn) {
			// Echo the greeting message
			msg = <-conn.send
			conn.receive <- msg
		},
	}
	l.conns = append(l.conns, c)

	// Run one connection and check what was sent
	s.Serve(l)

	if msg != protocol.Greeting {
		t.Errorf("server sent %s, expected %s", msg, protocol.Greeting)
	}
}

func TestServerAddToLobby(t *testing.T) {
	// Create server with 2 fake listeners that can be added to lobby
	s, _ := NewServer(nil)
	l := fakeListener{
		ch: make(chan Conn),
	}
	for i := 0; i < 2; i++ {
		c := fakeConn{
			send:    make(chan string),
			receive: make(chan string),
			poll: func(conn fakeConn) {
				// Echo the greeting message
				msg := <-conn.send
				conn.receive <- msg
			},
		}
		l.conns = append(l.conns, c)
	}

	// Run server
	s.Serve(l)

	// One lobby should be full
	lobby := s.lobbies[0]
	if !lobby.IsFull() {
		t.Errorf("server lobby %d was not full, expected to be full", lobby.ID())
	}
}
