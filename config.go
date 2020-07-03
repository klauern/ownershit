package ownershit

import (
	"os"

	"github.com/rs/zerolog/log"
)

type PermissionsLevel string

const (
	Admin PermissionsLevel = "admin"
	Read  PermissionsLevel = "pull"
	Write PermissionsLevel = "push"
)

type Permissions struct {
	Team  *string `yaml:"name"`
	Level *string `yaml:"level"`
}

type BranchPermissions struct {
	RequireCodeOwners         *bool `yaml:"require_code_owners"`
	ApproverCount             *int  `yaml:"require_approving_count"`
	RequirePullRequestReviews *bool `yaml:"require_pull_request_reviews"`
	AllowMergeCommit          *bool `yaml:"allow_merge_commit"`
	AllowSquashMerge          *bool `yaml:"allow_squash_merge"`
	AllowRebaseMerge          *bool `yaml:"allow_rebase_merge"`
}

type PermissionsSettings struct {
	BranchPermissions `yaml:"branches"`
	TeamPermissions   []*Permissions `yaml:"team"`
	Repositories      []struct {
		Name     *string
		Wiki     *bool
		Issues   *bool
		Projects *bool
	} `yaml:"repositories"`
	Organization *string `yaml:"organization"`
}

// GitHubTokenEnv sets the GitHub Token from the environment variable.
var GitHubTokenEnv = os.Getenv("GITHUB_TOKEN")

func MapPermissions(settings *PermissionsSettings, client *GitHubClient) {
	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				log.Info().
					Str("repository", *repo.Name).
					Msg("Adding Permissions to repository")
				log.Debug().
					Str("repository", *repo.Name).
					Str("permissions-level", *perm.Level).
					Str("permissions-team", *perm.Team).
					Msg("permissions to add to repository")
				err := client.AddPermissions(*settings.Organization, *repo.Name, perm)
				if err != nil {
					log.Err(err).
						Str("repository", *repo.Name).
						Str("permissions-level", *perm.Level).
						Str("permissions-team", *perm.Team).
						Msg("setting team permissions")
				}
			}
		}
		if err := client.UpdateRepositorySettings(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
			log.Err(err).
				Str("repository", *repo.Name).
				Str("organization", *settings.Organization).
				Msg("updating repository settings")
		}
		repoID, err := client.GetRepository(repo.Name, settings.Organization)
		if err != nil {
			log.Err(err).
				Str("repository", *repo.Name).
				Msg("getting repository")
		} else {
			log.Debug().
				Interface("repoID", repoID).
				Msg("Repository ID")
			err := client.SetRepository(&repoID, repo.Wiki, repo.Issues, repo.Projects)
			if err != nil {
				log.Err(err).
					Interface("repoID", repoID).
					Bool("wikiEnabled", *repo.Wiki).
					Bool("issuesEnabled", *repo.Issues).
					Bool("projectsEnabled", *repo.Projects).
					Msg("setting repository fields")
			}
		}
	}
}

func UpdateBranchMergeStrategies(settings *PermissionsSettings, client *GitHubClient) {
	for _, repo := range settings.Repositories {
		log.Info().
			Str("repository", *repo.Name).
			Bool("squash-commits", *settings.AllowSquashMerge).
			Bool("merges", *settings.AllowMergeCommit).
			Bool("rebase-merge", *settings.AllowRebaseMerge).
			Msg("Updating settings")
		if err := client.UpdateRepositorySettings(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
			log.Err(err).Str("repository", *repo.Name).Str("organization", *settings.Organization).Msg("updating repository settings")
		}
	}
}
