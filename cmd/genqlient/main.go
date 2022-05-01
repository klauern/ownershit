package main

import (
	"os"

	"github.com/klauern/ownershit/v4api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	client := v4api.NewGHv4Client()
	client.GetRateLimit()
	client.GetTeams()
}
