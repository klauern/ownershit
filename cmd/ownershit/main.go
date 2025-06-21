package main

import (
	"fmt"
	"os"

	shit "github.com/klauern/ownershit"
	"github.com/klauern/ownershit/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	settings     *shit.PermissionsSettings
	githubClient *shit.GitHubClient
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

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
			{
				Name:   "label",
				Usage:  "set default labels for repositories",
				Action: labelCommand,
			},
			{
				Name:   "ratelimit",
				Usage:  "get ratelimit information for the GitHub GraphQL v4 API",
				Action: rateLimitCommand,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "repositories.yaml",
				Usage: "configuration of repository updates to perform",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				EnvVars: []string{"OWNERSHIT_DEBUG"},
				Usage:   "set output to debug logging",
				Value:   false,
			},
		},
		Name:    "ownershit",
		Usage:   "fix up team ownership of your repositories in an organization",
		Action:  cli.ShowAppHelp,
		Authors: []*cli.Author{{Name: "Nick Klauer", Email: "klauer@gmail.com"}},
		Before:  configureClient,
		Version: version,
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"date":    date,
				"builtBy": builtBy,
				"commit":  commit,
			}
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("running app")
	}
}

func configureClient(c *cli.Context) error {
	err := readConfig(c)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	if c.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	client, err := shit.NewSecureGitHubClient(c.Context)
	if err != nil {
		log.Err(err).
			Str("operation", "initializeGitHubClient").
			Msg("GitHub client initialization failed")
		return fmt.Errorf("failed to initialize GitHub client: %w", err)
	}
	githubClient = client
	return nil
}

func readConfig(c *cli.Context) error {
	configPath := c.String("config")
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Err(err).
			Str("configPath", configPath).
			Str("operation", "readConfigFile").
			Msg("config file error")
		return shit.NewConfigFileError(configPath, "read", "failed to read configuration file", err)
	}

	settings = &shit.PermissionsSettings{}
	if err := yaml.Unmarshal(file, settings); err != nil {
		log.Err(err).
			Str("configPath", configPath).
			Str("operation", "parseConfigFile").
			Msg("YAML unmarshal error with config file")
		return shit.NewConfigFileError(configPath, "parse", "failed to parse YAML configuration", err)
	}

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

func labelCommand(c *cli.Context) error {
	log.Info().Msg("synchronizing labels on repositories")
	shit.SyncLabels(settings, githubClient)
	return nil
}

func rateLimitCommand(c *cli.Context) error {
	log.Info().Msg("getting ratelimit information")
	githubClient.GetRateLimit()
	return nil
}
