package config

// MaxPlayers specifies the max allowed players in a lobby.
const MaxPlayers = 2

// MaxTurnAttempts is the number of tries a player can have at making
// a valid turn before they are disconnected.
const MaxTurnAttempts = 3

// ConnDeadlineMinutes is the deadline in minutes for a read or write
// operation on a Conn.
const ConnDeadlineMinutes = 1
