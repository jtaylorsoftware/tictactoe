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

func (c *TcpConn) Send() chan<- string {
	return c.send
}

func (c *TcpConn) Receive() <-chan string {
	return c.receive
}

func (c *TcpConn) Close() error {
	if atomic.LoadInt32(&c.closed) != 0 {
		return errors.New("repeated call to Close")
	}
	atomic.StoreInt32(&c.closed, 1)

	c.receive <- "DISCONNECT"
	close(c.receive)
	return c.conn.Close()
}

const connDeadlineMinutes = 1

func isTemporary(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Temporary()
	}
	return true
}

func (c *TcpConn) pollSocket() {
	defer c.Close()

	r := bufio.NewReader(c.conn)

	for {
		err := c.conn.SetReadDeadline(minutesFromNow(connDeadlineMinutes))
		if err != nil {
			c.logger.Error("error setting TCP read deadline: ", err)
			if !isTemporary(err) {
				break
			}
		}

		// Read from the socket
		msg, err := r.ReadString('\n')
		if err != nil {
			if !isTemporary(err) {
				c.logger.Error("error reading from TCP socket: ", err)
				break
			}
		}

		if atomic.LoadInt32(&c.closed) != 0 {
			break
		}

		// Forward to server
		c.receive <- msg
	}
}

func (c *TcpConn) pollMessages() {
	defer c.Close()

	for {
		// Read server message
		msg, ok := <-c.send
		if !ok {
			c.logger.Info("could not receive from c.send: closed")
			break
		}

		err := c.conn.SetWriteDeadline(minutesFromNow(connDeadlineMinutes))
		if err != nil {
			c.logger.Error("error setting TCP write deadline: ", err)
			if !isTemporary(err) {
				break
			}
		}

		// Forward to client
		_, err = c.conn.Write([]byte(msg))
		if err != nil {
			c.logger.Error("error writing to TCP socket: ", err)
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

func (c *TcpConn) poll() {
	go c.pollSocket()
	go c.pollMessages()
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

func (l *TcpListener) PollAccept() error {
	defer l.Close()

	l.logger.Info("waiting for TCP connections on port ", l.port)

	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if errors.Is(err, syscall.EINVAL) {
				return fmt.Errorf("accept error (unrecoverable): %w", err)
			}
			var ne net.Error
			if errors.As(err, &ne) && !ne.Temporary() {
				return fmt.Errorf("accept error (unrecoverable): %w", err)
			}
			if l.logger != nil {
				l.logger.Error("accept error (recovered):", err)
			}
		}
		tcpConn := newTcpConn(conn, l.logger)
		go tcpConn.poll()
		l.connections <- tcpConn
	}
}

func (l *TcpListener) Connections() <-chan Conn {
	return l.connections
}

func (l *TcpListener) Close() error {
	close(l.connections)
	if err := l.listener.Close(); err != nil {
		return fmt.Errorf("error closing: %w", err)
	}
	return nil
}

func minutesFromNow(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * time.Minute)
}
