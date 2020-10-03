package cmd

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	shit "github.com/klauern/ownershit"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var ArchiveSubcommands = []*cli.Command{
	{
		Name:   "query",
		Action: queryCommand,
	},
	{
		Name:   "execute",
		Action: executeCommand,
	},
}

func queryCommand(c *cli.Context) error {
	log.Info().Msgf("querying with parms")
	client := shit.NewGitHubClient(context.Background(), shit.GitHubTokenEnv)
	repos, err := client.QueryArchivableIssues("klauern")
	if err != nil {
		return err
	}
	// sb := strings.Builder{}
	for _, repo := range repos {
		spew.Dump(repo)
	}
	return nil
}

func executeCommand(c *cli.Context) error {
	log.Info().Msgf("executing Archive with parms")
	return nil
}
