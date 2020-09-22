package main

import game "github.com/jeremyt135/tictactoe/pkg/game/server"

func main() {
	server := game.NewServer(game.DefaultOptions())
	defer server.Close()
	server.Listen(5432)
}
