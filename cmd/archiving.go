package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	shit "github.com/klauern/ownershit"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var ErrUsernameNotDefined = errors.New("username not defined for query")
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
			&cli.IntFlag{
				Name:  "days",
				Usage: "maximum number of days since last activity before considering for archival",
				Value: 365,
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
		return ErrUsernameNotDefined
	}
	repos, err := client.QueryArchivableIssues(username, c.Int("forks"), c.Int("stars"), c.Int("days"))
	if err != nil {
		return err
	}
	tableBuf := strings.Builder{}
	table := tablewriter.NewWriter(&tableBuf)
	table.SetHeader(
		[]string{"repository", "forks", "stars", "last updated"},
	)

	repos = shit.SortedRepositoryInfo(repos)
	for _, repo := range repos {
		forks := int(repo.ForkCount)
		stars := int(repo.StargazerCount)
		lastUpdated := repo.UpdatedAt.Time
		table.Append([]string{string(repo.Name), strconv.Itoa(forks), strconv.Itoa(stars), lastUpdated.String()})
	}
	table.Render()
	fmt.Println(tableBuf.String())
	return nil
}

func executeCommand(c *cli.Context) error {
	log.Info().Msgf("executing Archive with parms")
	return nil
}
