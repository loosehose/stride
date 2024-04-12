package main

import (
	"fmt"
	"net/http"

	"github.com/loosehose/stride/stride-server/internal/deployment/common"
	"github.com/loosehose/stride/stride-server/logging"
	"github.com/rs/zerolog/log"
)

const port = 8080

func init() {
	logging.InitLogger()
}

type application struct {
	Domain    string
	wsManager *common.WebSocketManager
}

func main() {
	app := &application{
		Domain:    "example.com",
		wsManager: common.NewWebSocketManager(),
	}
	// Start the server
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	log.Info().Msgf("Server started on port %d", port)
	if err != nil {
		log.Fatal().Msgf("Error starting server: ", err)
	}
}
