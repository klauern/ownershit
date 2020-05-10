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
				//Destination: &repositoriesYAMLConfig,
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
	settings := &shit.PermissionsSettings{}
	if err = yaml.Unmarshal(file, settings); err != nil {
		log.Err(err).Msg("YAML unmarshal error with config file")
		return fmt.Errorf("config file yaml unmarshal error: %w", err)
	}

	ctx := context.Background()
	client := shit.NewGitHubClient(ctx, GithubToken)

	for _, team := range settings.TeamPermissions {
		err = client.GetTeamSlug(settings, team)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				log.Info().
					Interface("repository", repo.Name).
					Msg("Adding Permissions to repository")
				log.Debug().
					Interface("repository", repo.Name).
					Interface("permissions", perm).
					Msg("permissions to add to repository")
				err = client.AddPermissions(repo.Name, settings.Organization, *perm)
				if err != nil {
					log.Err(err).
						Interface("repository", repo.Name).
						Interface("permissions", perm).
						Msg("setting team permissions")
				}
			}
		}
		repoID, err := client.GetRepository(repo.Name, settings.Organization)
		if err != nil {
			log.Err(err).Str("repository", repo.Name).Msg("getting repository")
		} else {
			log.Debug().
				Interface("repoID", repoID).
				Msg("Repository ID")
			err := client.SetRepository(&repoID, repo.Wiki, repo.Issues, repo.Projects)
			if err != nil {
				log.
					Err(err).
					Interface("repoID", repoID).
					Bool("wikiEnabled", repo.Wiki).
					Bool("issuesEnabled", repo.Issues).
					Bool("projectsEnabled", repo.Projects).
					Msg("setting repository fields")
			}
		}
	}
	return nil
}
