package cmd

import (
	"context"
	"errors"

	"github.com/davecgh/go-spew/spew"
	shit "github.com/klauern/ownershit"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var username string

var ArchiveSubcommands = []*cli.Command{
	{
		Name:   "query",
		Action: queryCommand,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &username,
				Aliases:     []string{"u"},
			},
			&cli.IntFlag{
				Name:  "forks",
				Usage: "maximum number of forks that a repo can have before it is considered archivalable",
				Value: 0,
			},
			&cli.IntFlag{
				Name:  "stars",
				Usage: "maximum number of stars that a repository should have before it is considered archivable",
				Value: 1,
			},
		},
	},
	{
		Name:   "execute",
		Action: executeCommand,
	},
}

func queryCommand(c *cli.Context) error {
	log.Info().Msgf("querying with parms")
	client := shit.NewGitHubClient(context.Background(), shit.GitHubTokenEnv)
	if username == "" {
		return errors.New("username not specified.  Please provide one to query.")
	}
	repos, err := client.QueryArchivableIssues(username, 1, 1)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		spew.Dump(repo)
	}
	return nil
}

func executeCommand(c *cli.Context) error {
	log.Info().Msgf("executing Archive with parms")
	return nil
}
