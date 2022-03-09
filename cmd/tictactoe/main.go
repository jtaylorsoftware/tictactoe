package main

import (
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/jeremyt135/tictactoe/pkg/server"
)

func setupLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("Could not create logger:", err)
	}
	sugar := logger.Sugar()
	return sugar
}

func main() {
	logger := setupLogger()
	defer logger.Sync()

	srv, err := server.NewServer(&server.Options{
		NumLobbies: 2,
		Logger:     logger,
	})
	if err != nil {
		log.Fatalln(err)
	}

	port := 42000
	if p, err := strconv.Atoi(os.Getenv("PORT")); err != nil && p != 0 {
		port = p
	}

	mode := "tcp"
	if m := os.Getenv("MODE"); m == "ws" {
		mode = m
	}

	var l server.Listener
	switch mode {
	case "tcp":
		tcp, err := server.ListenTcp(port, logger)
		if err != nil {
			log.Fatalln(err)
		}
		l = tcp
		defer tcp.Close()
	case "ws":
		ws, err := server.ListenWS(port, logger)
		if err != nil {
			log.Fatalln(err)
		}
		l = ws
		defer ws.Close()
	}

	if err := srv.Serve(l); err != nil {
		log.Fatalln(err)
	} else {
		defer srv.Close()
	}
}
