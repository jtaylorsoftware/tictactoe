package logger

import (
	"os"
)

// Logger encapsulates interchangeable logging functionality needed by the server.
type Logger interface {
	// Info writes data with "Info" priority. It formats output like fmt.Sprint().
	Info(...interface{})

	// Warn writes data with "Warning" priority. It formats output like fmt.Sprint().
	Warn(...interface{})

	// Error writes data with "Error" priority. It formats output like fmt.Sprint().
	Error(...interface{})

	// Fatal writes data with "Fatal" priority and calls os.Exit(1). It formats output like fmt.Sprint().
	Fatal(...interface{})
}

type noOp struct{}

var noOpSingleton = &noOp{}

func (n *noOp) Warn(...interface{})  {}
func (n *noOp) Error(...interface{}) {}
func (n *noOp) Info(...interface{})  {}
func (n *noOp) Fatal(...interface{}) { os.Exit(1) }

// NoOpLogger returns an instance of a Logger that does not perform writes when called.
// The same instance is returned every time.
func NoOpLogger() Logger {
	return noOpSingleton
}
