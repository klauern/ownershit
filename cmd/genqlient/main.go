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

	client, err := v4api.NewGHv4Client()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize v4 client")
	}
	if _, err := client.GetRateLimit(); err != nil {
		log.Error().Err(err).Msg("failed to get rate limit")
	}
	if _, err := client.GetTeams("github"); err != nil {
		log.Error().Err(err).Msg("failed to get teams")
	}
}
