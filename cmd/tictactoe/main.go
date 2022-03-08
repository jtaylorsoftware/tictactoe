package main

import (
	"log"

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

	l, err := server.ListenTcp(5432, logger)
	if err != nil {
		log.Fatalln(l)
	}

	if err := srv.Serve(l); err != nil {
		log.Fatalln(err)
	}
}
