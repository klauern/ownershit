package ownershit

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

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
	Repositories      []*Repository  `yaml:"repositories"`
	Organization      *string        `yaml:"organization"`
	DefaultLabels     []RepoLabel    `yaml:"default_labels"`
}

type Repository struct {
	Name     *string
	Wiki     *bool
	Issues   *bool
	Projects *bool
}

type RepoLabel struct {
	Name        string
	Color       string
	Emoji       string
	Description string
	oldLabel    string
}

// GitHub token validation errors.
var (
	ErrTokenEmpty    = errors.New("GitHub token is empty")
	ErrTokenInvalid  = errors.New("GitHub token format is invalid")
	ErrTokenNotFound = errors.New("GITHUB_TOKEN environment variable not set")
)

// ValidateGitHubToken validates a GitHub token format and content.
func ValidateGitHubToken(token string) error {
	if token == "" {
		return ErrTokenEmpty
	}

	// Remove whitespace
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTokenEmpty
	}

	// GitHub tokens have specific patterns:
	// - Classic tokens: ghp_[A-Za-z0-9]{36}
	// - Fine-grained tokens: github_pat_[A-Za-z0-9_]+
	// - GitHub App tokens: ghs_[A-Za-z0-9]{36}
	// - OAuth tokens: gho_[A-Za-z0-9]{36}
	// - Refresh tokens: ghr_[A-Za-z0-9]{36}
	// - SAML tokens: ghu_[A-Za-z0-9]{36}
	validPatterns := []string{
		`^ghp_[A-Za-z0-9]{36}$`,             // Classic personal access token
		`^github_pat_[A-Za-z0-9_]{22,255}$`, // Fine-grained personal access token
		`^ghs_[A-Za-z0-9]{36}$`,             // GitHub App token
		`^gho_[A-Za-z0-9]{36}$`,             // OAuth token
		`^ghr_[A-Za-z0-9]{36}$`,             // Refresh token
		`^ghu_[A-Za-z0-9]{36}$`,             // SAML token
	}

	for _, pattern := range validPatterns {
		matched, err := regexp.MatchString(pattern, token)
		if err != nil {
			return fmt.Errorf("token validation regex error: %w", err)
		}
		if matched {
			return nil
		}
	}

	return ErrTokenInvalid
}

// GetValidatedGitHubToken gets and validates the GitHub token from environment.
func GetValidatedGitHubToken() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", ErrTokenNotFound
	}

	if err := ValidateGitHubToken(token); err != nil {
		return "", fmt.Errorf("invalid GitHub token: %w", err)
	}

	return token, nil
}

// GitHubTokenEnv sets the GitHub Token from the environment variable.
// Deprecated: Use GetValidatedGitHubToken() for secure token handling.
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
						Str("operation", "addTeamPermissions").
						Msg("setting team permissions")
					// Continue processing other permissions even if one fails
				}
			}
		}
		if err := client.UpdateBranchPermissions(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
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
		if err := client.UpdateBranchPermissions(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
			log.Err(err).Str("repository", *repo.Name).Str("organization", *settings.Organization).Msg("updating repository settings")
		}
	}
}

func SyncLabels(settings *PermissionsSettings, client *GitHubClient) {
	for _, repo := range settings.Repositories {
		log.Info().
			Str("repository", *repo.Name).
			Msg("Updating Labels")
		if err := client.SyncLabels(*settings.Organization, *repo.Name, settings.DefaultLabels); err != nil {
			log.Err(err).Msg("synchronizing Labels")
		}
	}
}
