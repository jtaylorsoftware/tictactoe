package player

// Player keeps a record of a connection and its identity.
type Player struct {
	Token string
	ID    int
	Conn  Conn
}

// New returns a pointer to a Player that will use the given connection.
func New(c Conn) *Player {
	return &Player{Conn: c}
}
