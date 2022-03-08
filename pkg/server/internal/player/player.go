package player

// Player keeps a record of a connection and its identity.
type Player struct {
	Token   string
	ID      int
	Send    chan<- string
	Receive <-chan string
}

// New returns a pointer to a Player that will use the given connection.
func New(send chan<- string, receive <-chan string) *Player {
	return &Player{Send: send, Receive: receive}
}
