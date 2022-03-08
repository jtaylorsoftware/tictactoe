package server

import (
	"github.com/jeremyt135/tictactoe/pkg/logger"
)

// Options hold configuration data for a server.
type Options struct {
	NumLobbies int
	Logger     logger.Logger
}

// DefaultOptions returns default Options for configuring a server.
// The default Logger used does nothing.
func DefaultOptions() *Options {
	return &Options{NumLobbies: 2, Logger: logger.NoOpLogger()}
}
