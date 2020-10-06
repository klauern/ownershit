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

var client *shit.GitHubClient
var (
	forks int
	stars int
	days  int
)

var archiveFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "username",
		Usage:       "GitHub username",
		Destination: &username,
		Aliases:     []string{"u"},
	},
	&cli.IntFlag{
		Name:        "forks",
		Usage:       "maximum number of forks that a repo can have before it is considered archivalable",
		Destination: &forks,
		Value:       0,
	},
	&cli.IntFlag{
		Name:        "stars",
		Usage:       "maximum number of stars that a repository should have before it is considered archivable",
		Destination: &stars,
		Value:       0,
	},
	&cli.IntFlag{
		Name:        "days",
		Usage:       "maximum number of days since last activity before considering for archival",
		Destination: &days,
		Value:       365,
	},
}

var ArchiveSubcommands = []*cli.Command{
	{
		Name:   "query",
		Action: queryCommand,
		Flags:  archiveFlags,
		Before: func(c *cli.Context) error {
			if username == "" {
				return ErrUsernameNotDefined
			}
			client = shit.NewGitHubClient(context.Background(), shit.GitHubTokenEnv)
			return nil
		},
	},
	{
		Name:   "execute",
		Action: executeCommand,
		Flags:  archiveFlags,
	},
}

func queryCommand(c *cli.Context) error {
	log.
		Info().
		Fields(map[string]interface{}{
			"forks": forks,
			"stars": stars,
			"days":  days,
		}).
		Msgf("querying with parms")
	repos, err := client.QueryArchivableIssues(username, forks, stars, days)
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
		repoForks := int(repo.ForkCount)
		repoStars := int(repo.StargazerCount)
		lastUpdated := repo.UpdatedAt.Time
		table.Append([]string{string(repo.Name), strconv.Itoa(repoForks), strconv.Itoa(repoStars), lastUpdated.String()})
	}
	table.Render()
	fmt.Println(tableBuf.String())
	return nil
}

func executeCommand(c *cli.Context) error {
	log.Info().Msgf("executing Archive with parms")
	_, err := client.QueryArchivableIssues(username, forks, stars, days)
	if err != nil {
		return err
	}
	return nil
}
