package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jeremyt135/tictactoe/pkg/logger"
)

// TcpListener listens for incoming TCP connections.
type TcpListener struct {
	listener    net.Listener
	connections chan Conn
	logger      logger.Logger
	port        int
}

// TcpConn wraps an incoming connection and forwards data
// from it to channels.
type TcpConn struct {
	conn    net.Conn
	logger  logger.Logger
	send    chan string // channel for server to send messages to connected client
	receive chan string // channel for server to receive messages from client
	closed  int32
}

func newTcpConn(conn net.Conn, logger logger.Logger) *TcpConn {
	return &TcpConn{
		conn:    conn,
		send:    make(chan string, 10),
		receive: make(chan string, 10),
		logger:  logger,
	}
}

func (t *TcpConn) Send() chan<- string {
	return t.send
}

func (t *TcpConn) Receive() <-chan string {
	return t.receive
}

func (t *TcpConn) Close() error {
	if atomic.LoadInt32(&t.closed) != 0 {
		return errors.New("repeated call to Close")
	}
	atomic.StoreInt32(&t.closed, 1)

	t.receive <- "DISCONNECT"
	close(t.receive)
	return t.conn.Close()
}

const connDeadlineMinutes = 1

func isTemporary(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Temporary()
	}
	return true
}

func (t *TcpConn) pollSocket() {
	defer t.Close()

	r := bufio.NewReader(t.conn)

	for {
		err := t.conn.SetReadDeadline(minutesFromNow(connDeadlineMinutes))
		if err != nil {
			t.logger.Error("error setting TCP read deadline: ", err)
			if !isTemporary(err) {
				break
			}
		}

		// Read from the socket
		msg, err := r.ReadString('\n')
		if err != nil {
			if !isTemporary(err) {
				t.logger.Error("error reading from TCP socket: ", err)
				break
			}
		}

		// Forward to server
		t.receive <- msg
	}
}

func (t *TcpConn) pollMessages() {
	defer t.Close()

	for {
		// Read server message
		msg, ok := <-t.send
		if !ok {
			t.logger.Info("could not receive from t.send: closed")
			break
		}

		err := t.conn.SetWriteDeadline(minutesFromNow(connDeadlineMinutes))
		if err != nil {
			t.logger.Error("error setting TCP write deadline: ", err)
			if !isTemporary(err) {
				break
			}
		}

		// Forward to client
		_, err = t.conn.Write([]byte(msg))
		if err != nil {
			t.logger.Error("error writing to TCP socket: ", err)
			if !isTemporary(err) {
				break
			}
		}

		// Server removed client for some reason
		if msg == "REMOVED\n" {
			break
		}
	}
}

func (t *TcpConn) poll() {
	go t.pollSocket()
	go t.pollMessages()
}

// ListenTcp creates a new TcpListener listening on the given port.
// Logger may be nil in which case no log output will be generated.
func ListenTcp(port int, logger logger.Logger) (*TcpListener, error) {
	l, err := net.Listen("tcp4", fmt.Sprint(":", port))
	if err != nil {
		return nil, fmt.Errorf("could not create TcpListener: %w", err)
	}
	return &TcpListener{
		listener:    l,
		logger:      logger,
		connections: make(chan Conn, 100),
		port:        port,
	}, nil
}

func (t *TcpListener) PollAccept() error {
	defer t.Close()

	t.logger.Info("waiting for TCP connections on port ", t.port)

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if errors.Is(err, syscall.EINVAL) {
				return fmt.Errorf("accept error (unrecoverable): %w", err)
			}
			var ne net.Error
			if errors.As(err, &ne) && !ne.Temporary() {
				return fmt.Errorf("accept error (unrecoverable): %w", err)
			}
			if t.logger != nil {
				t.logger.Error("accept error (recovered):", err)
			}
		}
		tcpConn := newTcpConn(conn, t.logger)
		go tcpConn.poll()
		t.connections <- tcpConn
	}
}

func (t *TcpListener) Connections() <-chan Conn {
	return t.connections
}

func (t *TcpListener) Close() error {
	close(t.connections)
	if err := t.listener.Close(); err != nil {
		return fmt.Errorf("error closing: %w", err)
	}
	return nil
}

func minutesFromNow(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * time.Minute)
}
