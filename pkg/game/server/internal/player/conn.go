package player

import (
	"bufio"
	"net"
	"time"

	"github.com/jeremyt135/tictactoe/pkg/game/server/internal/config"
)

// Conn wraps a net.Conn with useful utilities for IO ops.
type Conn struct {
	c  net.Conn
	rw *bufio.ReadWriter
}

// NewConn creates a new wrapping for a net.Conn.
func NewConn(conn net.Conn) Conn {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	return Conn{conn, bufio.NewReadWriter(r, w)}
}

// ReadString reads one line (\n delimited) of input from the connection.
// Returns the read string or an error if a delimited string could not be read.
func (conn Conn) ReadString() (string, error) {
	conn.updateDeadine()
	return conn.rw.ReadString('\n')
}

// WriteString writes a string to the connection.
// Returns the length of the string written. If an error occurs it returns
// the possible length written and the error.
func (conn Conn) WriteString(s string) (int, error) {
	conn.updateDeadine()
	n, err := conn.rw.WriteString(s)
	if err == nil {
		// flush if the Writer op succeeded, may also cause an error
		err = conn.rw.Flush()
	}
	return n, err
}

// Close closes the player connection.
func (conn Conn) Close() error {
	return conn.c.Close()
}

func (conn Conn) updateDeadine() {
	conn.c.SetDeadline(minutesFromNow(config.ConnDeadlineMinutes))
}

// minutesFromNow is a utility function returning a time from now
func minutesFromNow(minutes int) time.Time {
	return time.Now().Add(time.Duration(minutes) * time.Minute)
}
