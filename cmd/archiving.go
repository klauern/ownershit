package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	shit "github.com/klauern/ownershit"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var ErrUsernameNotDefined = errors.New("username not defined for query")
var username string

var client *shit.GitHubClient
var (
	forks    int
	stars    int
	days     int
	watchers int
	forking  bool
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
	&cli.IntFlag{
		Name:        "watchers",
		Usage:       "maximum number of watchers before considering for archival",
		Destination: &watchers,
		Value:       0,
	},
	&cli.BoolFlag{
		Name:        "forking",
		Usage:       "enable or disable the forking of repositories",
		Destination: &forking,
		Value:       true,
	},
}

var ArchiveSubcommands = []*cli.Command{
	{
		Name:   "query",
		Action: queryCommand,
		Flags:  archiveFlags,
		Before: userClientSetup,
	},
	{
		Name:   "execute",
		Action: executeCommand,
		Flags:  archiveFlags,
		Before: userClientSetup,
	},
}

func userClientSetup(c *cli.Context) error {
	if username == "" {
		return ErrUsernameNotDefined
	}
	client = shit.NewGitHubClient(context.Background(), shit.GitHubTokenEnv)
	return nil
}

func queryCommand(c *cli.Context) error {
	log.
		Info().
		Fields(map[string]interface{}{
			"username": username,
			"forks":    forks,
			"stars":    stars,
			"days":     days,
			"watchers": watchers,
			"forking":  forking,
		}).
		Msgf("querying with parms")
	repos, err := client.QueryArchivableRepos(username, forks, stars, days, watchers, forking)
	if err != nil {
		return err
	}
	tableBuf := strings.Builder{}
	table := tablewriter.NewWriter(&tableBuf)
	table.SetHeader(
		[]string{"repository", "forks", "stars", "watchers", "last updated"},
	)

	repos = shit.SortedRepositoryInfo(repos)
	for _, repo := range repos {
		repoForks := strconv.Itoa(int(repo.ForkCount))
		repoStars := strconv.Itoa(int(repo.StargazerCount))
		repoWatchers := strconv.Itoa(int(repo.Watchers.TotalCount))
		lastUpdated := repo.UpdatedAt.Time
		table.Append([]string{string(repo.Name), repoForks, repoStars, repoWatchers, lastUpdated.String()})
	}
	table.Render()
	fmt.Println(tableBuf.String())
	return nil
}

func executeCommand(c *cli.Context) error {
	log.
		Info().
		Fields(map[string]interface{}{
			"username": username,
			"forks":    forks,
			"stars":    stars,
			"days":     days,
			"watchers": watchers,
			"forking":  forking,
		}).
		Msgf("executing Archive with parms")
	issues, err := client.QueryArchivableRepos(username, forks, stars, days, watchers, forking)
	if err != nil {
		return err
	}
	var ops []string
	for _, v := range issues {
		ops = append(ops, string(v.Name))
	}
	archiving := &survey.MultiSelect{
		Message: "Which repositories do you wish to archive?",
		Options: ops,
	}
	repoNames := []string{}
	err = survey.AskOne(archiving, &repoNames)
	if err != nil {
		return fmt.Errorf("error from survey question: %w", err)
	}
	repoMapping := mapRepoNames(issues)

	for _, r := range repoNames {
		err = client.MutateArchiveRepository(repoMapping[r])
		if err != nil {
			return fmt.Errorf("archiving %v: %w", r, err)
		}
	}
	return nil
}

func mapRepoNames(r []shit.RepositoryInfo) map[string]shit.RepositoryInfo {
	repos := make(map[string]shit.RepositoryInfo, len(r))
	for _, v := range r {
		repos[string(v.Name)] = v
	}
	return repos
}
