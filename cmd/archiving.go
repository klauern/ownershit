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
	repos, err := client.QueryArchivableIssues(username)
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
