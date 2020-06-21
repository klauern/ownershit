package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	shit "github.com/klauern/ownershit"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var GithubToken = os.Getenv("GITHUB_TOKEN")

func main() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := &cli.App{
		Name:      "ownershit",
		Usage:     "fix up team ownership of your repositories in an organization",
		UsageText: "ownershit --config repositories.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "repositories.yaml",
				Usage: "configuration of repository updates to perform",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "set output to debug logging",
				Value: false,
			},
		},
		Action:  runApp,
		Authors: []*cli.Author{{Name: "Nick Klauer", Email: "klauer@gmail.com"}},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("running app")
	}

}

func runApp(c *cli.Context) error {
	file, err := ioutil.ReadFile(c.String("config"))
	if err != nil {
		log.Err(err).Msg("config file error")
		return fmt.Errorf("config file error: %w", err)
	}
	if c.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	settings, client, err2 := MapTeams(file)
	if err2 != nil {
		return err2
	}

	shit.MapPermissions(settings, client)
	return nil
}

func MapTeams(file []byte) (*shit.PermissionsSettings, *shit.GitHubClient, error) {
	settings := &shit.PermissionsSettings{}
	if err := yaml.Unmarshal(file, settings); err != nil {
		log.Err(err).Msg("YAML unmarshal error with config file")
		return nil, nil, fmt.Errorf("config file yaml unmarshal error: %w", err)
	}

	ctx := context.Background()
	client := shit.NewGitHubClient(ctx, GithubToken)

	for _, team := range settings.TeamPermissions {
		err := client.SetTeamSlug(settings, team)
		if err != nil {
			fmt.Println(err)
		}
	}
	return settings, client, nil
}
