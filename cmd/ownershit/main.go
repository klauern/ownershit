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
var settings *shit.PermissionsSettings
var githubClient *shit.GitHubClient

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "branches",
				Usage:  "Perform branch management steps from repositories.yaml",
				Action: branchCommand,
			},
			{
				Name:      "sync",
				Usage:     "Synchronize branch, repo, owner and other configs on repositories",
				UsageText: "ownershit sync --config repositories.yaml",
				Action:    runApp,
			},
		},
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
		Name:    "ownershit",
		Usage:   "fix up team ownership of your repositories in an organization",
		Action:  cli.ShowAppHelp,
		Authors: []*cli.Author{{Name: "Nick Klauer", Email: "klauer@gmail.com"}},
		Before:  readConfigs,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("running app")
	}
}

func readConfigs(c *cli.Context) error {
	file, err := ioutil.ReadFile(c.String("config"))
	if err != nil {
		log.Err(err).Msg("config file error")
		return fmt.Errorf("config file error: %w", err)
	}
	if c.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	settings = &shit.PermissionsSettings{}
	if err := yaml.Unmarshal(file, settings); err != nil {
		log.Err(err).Msg("YAML unmarshal error with config file")
		return fmt.Errorf("config file yaml unmarshal error: %w", err)
	}

	SetGithubClient(context.Background())

	return nil
}

func runApp(c *cli.Context) error {
	shit.MapPermissions(settings, githubClient)
	for _, team := range settings.TeamPermissions {
		err := githubClient.SetTeamSlug(settings, team)
		if err != nil {
			log.Error().AnErr("team", err).Msg("setting Team Slug")
		}
	}
	return nil
}

func branchCommand(c *cli.Context) error {
	shit.UpdateBranchMergeStrategies(settings, githubClient)
	return nil
}

func SetGithubClient(context context.Context) {
	githubClient = shit.NewGitHubClient(context, GithubToken)
}
