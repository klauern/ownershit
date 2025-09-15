// Package cmd provides command-line interface commands for the ownershit tool.
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

var (
	// ErrUsernameNotDefined is returned when no username is provided for a query.
	ErrUsernameNotDefined = errors.New("username not defined for query")
	username              string
)

var client *shit.GitHubClient
var (
	forks    int
	stars    int
	days     int
	watchers int
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
		Usage:       "maximum number of forks that a repo can have before it is considered archivable",
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
}

// ArchiveSubcommands defines the available archive-related CLI commands.
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

// userClientSetup validates the username parameter and sets up the GitHub client.
func userClientSetup(c *cli.Context) error {
	if strings.TrimSpace(username) == "" {
		return ErrUsernameNotDefined
	}
	secureClient, err := shit.NewSecureGitHubClient(context.Background())
	if err != nil {
		log.Err(err).
			Str("operation", "initializeSecureGitHubClient").
			Msg("failed to initialize secure GitHub client")
		return fmt.Errorf("failed to initialize secure GitHub client: %w", err)
	}
	client = secureClient
	return nil
}

// queryCommand queries repositories that match the archiving criteria.
func queryCommand(c *cli.Context) error {
	log.
		Info().
		Fields(map[string]interface{}{
			"username": username,
			"forks":    forks,
			"stars":    stars,
			"days":     days,
			"watchers": watchers,
		}).
		Msg("querying with parms")
	repos, err := client.QueryArchivableRepos(username, forks, stars, days, watchers)
	if err != nil {
		log.Err(err).
			Str("username", username).
			Str("operation", "queryArchivableRepos").
			Msg("failed to query archivable repositories")
		return fmt.Errorf("failed to query archivable repositories for user %s: %w", username, err)
	}
	tableBuf := strings.Builder{}
	table := tablewriter.NewWriter(&tableBuf)
	table.Header([]string{"repository", "forks", "stars", "watchers", "last updated"})

	repos = shit.SortedRepositoryInfo(repos)
	for _, repo := range repos {
		repoForks := strconv.Itoa(int(repo.ForkCount))
		repoStars := strconv.Itoa(int(repo.StargazerCount))
		repoWatchers := strconv.Itoa(int(repo.Watchers.TotalCount))
		lastUpdated := repo.UpdatedAt.Time
		if err := table.Append([]string{
			string(repo.Name), repoForks, repoStars, repoWatchers, lastUpdated.String(),
		}); err != nil {
			log.Warn().Err(err).Msg("failed to append table row")
		}
	}
	if err := table.Render(); err != nil {
		log.Warn().Err(err).Msg("failed to render table")
	}
	fmt.Println(tableBuf.String())
	return nil
}

// executeCommand executes the archiving process for repositories matching the criteria.
func executeCommand(c *cli.Context) error {
	log.
		Info().
		Fields(map[string]interface{}{
			"username": username,
			"forks":    forks,
			"stars":    stars,
			"days":     days,
			"watchers": watchers,
		}).
		Msg("executing Archive with parms")
	issues, err := client.QueryArchivableRepos(username, forks, stars, days, watchers)
	if err != nil {
		log.Err(err).
			Str("username", username).
			Str("operation", "queryArchivableRepos").
			Msg("failed to query archivable repositories for execution")
		return fmt.Errorf("failed to query archivable repositories for user %s: %w", username, err)
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
		repo, ok := repoMapping[r]
		if !ok {
			log.Warn().Str("repo", r).Msg("selected repository not found in mapping; skipping")
			continue
		}
		err = client.MutateArchiveRepository(repo)
		if err != nil {
			return fmt.Errorf("archiving %v: %w", r, err)
		}
	}
	return nil
}

// mapRepoNames creates a map of repository names to RepositoryInfo for quick lookup.
func mapRepoNames(r []shit.RepositoryInfo) map[string]shit.RepositoryInfo {
	repos := make(map[string]shit.RepositoryInfo, len(r))
	for _, v := range r {
		repos[string(v.Name)] = v
	}
	return repos
}
