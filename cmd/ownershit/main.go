package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	shit "github.com/klauern/ownershit"
	"github.com/klauern/ownershit/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

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
				Action:    syncCommand,
			},
			{
				Name:        "archive",
				Usage:       "Archive repositories",
				Subcommands: cmd.ArchiveSubcommands,
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

func syncCommand(c *cli.Context) error {
	log.Info().Msg("mapping all permissions for repositories")
	shit.MapPermissions(settings, githubClient)
	return nil
}

func branchCommand(c *cli.Context) error {
	log.Info().Msg("performing branch updates on repositories")
	shit.UpdateBranchMergeStrategies(settings, githubClient)
	return nil
}

func SetGithubClient(ctx context.Context) {
	githubClient = shit.NewGitHubClient(ctx, shit.GitHubTokenEnv)
}
