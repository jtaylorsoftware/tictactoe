package server

import "github.com/jeremyt135/tictactoe/pkg/logger"

// WSListener listens for incoming WebSocket connections.
type WSListener struct {

}

// ListenWS creates a new WSListener listening for HTTP connections at the server root on the given port.
func ListenWS(port int, logger logger.Logger) (*TcpListener,error){
	return nil, nil
}

func (ws *WSListener) PollAccept() error {
	panic("implement me")
}

func (ws *WSListener) Connections() <-chan Conn {
	panic("implement me")
}

func (ws *WSListener) Close() error {
	panic("implement me")
}

