// Package main provides a script to generate an ownershit configuration file
// from all repositories owned by a GitHub user.
package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Config mirrors the structure from config.go but simplified for generation
type Config struct {
	Version       string              `yaml:"version,omitempty"`
	Organization  string              `yaml:"organization"`
	Defaults      *RepositoryDefaults `yaml:"defaults,omitempty"`
	Branches      *BranchPermissions  `yaml:"branches,omitempty"`
	Team          []*TeamPermission   `yaml:"team,omitempty"`
	Repositories  []*Repository       `yaml:"repositories"`
	DefaultLabels []Label             `yaml:"default_labels,omitempty"`
	DefaultTopics []string            `yaml:"default_topics,omitempty"`
}

type RepositoryDefaults struct {
	Wiki                *bool `yaml:"wiki,omitempty"`
	Issues              *bool `yaml:"issues,omitempty"`
	Projects            *bool `yaml:"projects,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty"`
}

type BranchPermissions struct {
	RequireCodeOwners         *bool    `yaml:"require_code_owners,omitempty"`
	ApproverCount             *int     `yaml:"require_approving_count,omitempty"`
	RequirePullRequestReviews *bool    `yaml:"require_pull_request_reviews,omitempty"`
	AllowMergeCommit          *bool    `yaml:"allow_merge_commit,omitempty"`
	AllowSquashMerge          *bool    `yaml:"allow_squash_merge,omitempty"`
	AllowRebaseMerge          *bool    `yaml:"allow_rebase_merge,omitempty"`
	RequireStatusChecks       *bool    `yaml:"require_status_checks,omitempty"`
	StatusChecks              []string `yaml:"status_checks,omitempty"`
	RequireUpToDateBranch     *bool    `yaml:"require_up_to_date_branch,omitempty"`
	EnforceAdmins             *bool    `yaml:"enforce_admins,omitempty"`
	RestrictPushes            *bool    `yaml:"restrict_pushes,omitempty"`
	PushAllowlist             []string `yaml:"push_allowlist,omitempty"`
}

type TeamPermission struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

type Repository struct {
	Name                   string  `yaml:"name"`
	Wiki                   *bool   `yaml:"wiki,omitempty"`
	Issues                 *bool   `yaml:"issues,omitempty"`
	Projects               *bool   `yaml:"projects,omitempty"`
	DefaultBranch          *string `yaml:"default_branch,omitempty"`
	Private                *bool   `yaml:"private,omitempty"`
	Archived               *bool   `yaml:"archived,omitempty"`
	Template               *bool   `yaml:"template,omitempty"`
	Description            *string `yaml:"description,omitempty"`
	Homepage               *string `yaml:"homepage,omitempty"`
	DeleteBranchOnMerge    *bool   `yaml:"delete_branch_on_merge,omitempty"`
	HasDiscussionsEnabled  *bool   `yaml:"discussions_enabled,omitempty"`
	HasSponsorshipsEnabled *bool   `yaml:"sponsorships_enabled,omitempty"`
}

type Label struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	Emoji       string `yaml:"emoji,omitempty"`
	Description string `yaml:"description,omitempty"`
}

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Check for required arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <github-username> [output-file]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s klauern klauern-repositories.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN - GitHub Personal Access Token (required)\n")
		os.Exit(1)
	}

	username := os.Args[1]
	outputFile := username + "-repositories.yaml"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	// Get GitHub token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal().Msg("GITHUB_TOKEN environment variable not set")
	}

	// Create GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	log.Info().
		Str("username", username).
		Str("output", outputFile).
		Msg("fetching repositories from GitHub")

	// Fetch all repositories for the user
	repos, err := fetchUserRepositories(ctx, client, username)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to fetch repositories")
	}

	log.Info().Int("count", len(repos)).Msg("repositories found")

	// Generate configuration
	config := generateConfig(username, repos)

	// Write to YAML file
	if err := writeYAMLConfig(config, outputFile); err != nil {
		log.Fatal().Err(err).Msg("failed to write configuration file")
	}

	log.Info().
		Str("file", outputFile).
		Msg("configuration file generated successfully")
}

func fetchUserRepositories(ctx context.Context, client *github.Client, username string) ([]*github.Repository, error) {
	opt := &github.RepositoryListOptions{
		Type: "owner", // Only repositories owned by the user
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List(ctx, username, opt)
		if err != nil {
			return nil, fmt.Errorf("listing repositories: %w", err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Sort repositories by name for consistent output
	sort.Slice(allRepos, func(i, j int) bool {
		return strings.ToLower(allRepos[i].GetName()) < strings.ToLower(allRepos[j].GetName())
	})

	return allRepos, nil
}

func generateConfig(username string, repos []*github.Repository) *Config {
	// Calculate defaults based on most common settings
	defaults := calculateDefaults(repos)

	// Convert repositories
	var configRepos []*Repository
	for _, repo := range repos {
		configRepo := &Repository{
			Name: repo.GetName(),
		}

		// Only include fields that differ from defaults or are important metadata
		if repo.HasWiki != nil && (defaults.Wiki == nil || *repo.HasWiki != *defaults.Wiki) {
			configRepo.Wiki = repo.HasWiki
		}
		if repo.HasIssues != nil && (defaults.Issues == nil || *repo.HasIssues != *defaults.Issues) {
			configRepo.Issues = repo.HasIssues
		}
		if repo.HasProjects != nil && (defaults.Projects == nil || *repo.HasProjects != *defaults.Projects) {
			configRepo.Projects = repo.HasProjects
		}
		if repo.DeleteBranchOnMerge != nil && (defaults.DeleteBranchOnMerge == nil || *repo.DeleteBranchOnMerge != *defaults.DeleteBranchOnMerge) {
			configRepo.DeleteBranchOnMerge = repo.DeleteBranchOnMerge
		}

		// Always include these fields if they have meaningful values
		if repo.Private != nil && *repo.Private {
			configRepo.Private = repo.Private
		}
		if repo.Archived != nil && *repo.Archived {
			configRepo.Archived = repo.Archived
		}
		if repo.IsTemplate != nil && *repo.IsTemplate {
			configRepo.Template = repo.IsTemplate
		}
		if repo.Description != nil && *repo.Description != "" {
			configRepo.Description = repo.Description
		}
		if repo.Homepage != nil && *repo.Homepage != "" {
			configRepo.Homepage = repo.Homepage
		}
		if repo.DefaultBranch != nil && *repo.DefaultBranch != "" && *repo.DefaultBranch != "main" {
			configRepo.DefaultBranch = repo.DefaultBranch
		}
		if repo.HasDiscussions != nil && *repo.HasDiscussions {
			configRepo.HasDiscussionsEnabled = repo.HasDiscussions
		}

		configRepos = append(configRepos, configRepo)
	}

	// Build basic configuration structure
	config := &Config{
		Version:      "1.0",
		Organization: username,
		Defaults:     defaults,
		Repositories: configRepos,
		// Empty arrays to be filled in by user
		Team:          []*TeamPermission{},
		DefaultLabels: []Label{},
		DefaultTopics: []string{},
	}

	// Add minimal branch protection settings as examples
	falseVal := false
	trueVal := true
	config.Branches = &BranchPermissions{
		RequirePullRequestReviews: &falseVal,
		ApproverCount:             new(int), // 0 approvers
		RequireCodeOwners:         &falseVal,
		AllowMergeCommit:          &trueVal,
		AllowSquashMerge:          &trueVal,
		AllowRebaseMerge:          &trueVal,
	}

	return config
}

func calculateDefaults(repos []*github.Repository) *RepositoryDefaults {
	if len(repos) == 0 {
		return &RepositoryDefaults{}
	}

	// Count occurrences of each setting
	wikiCount := 0
	issuesCount := 0
	projectsCount := 0
	deleteBranchCount := 0

	for _, repo := range repos {
		if repo.HasWiki != nil && *repo.HasWiki {
			wikiCount++
		}
		if repo.HasIssues != nil && *repo.HasIssues {
			issuesCount++
		}
		if repo.HasProjects != nil && *repo.HasProjects {
			projectsCount++
		}
		if repo.DeleteBranchOnMerge != nil && *repo.DeleteBranchOnMerge {
			deleteBranchCount++
		}
	}

	total := len(repos)
	defaults := &RepositoryDefaults{}

	// Use majority (>50%) as default
	if wikiCount > total/2 {
		trueVal := true
		defaults.Wiki = &trueVal
	} else {
		falseVal := false
		defaults.Wiki = &falseVal
	}

	if issuesCount > total/2 {
		trueVal := true
		defaults.Issues = &trueVal
	} else {
		falseVal := false
		defaults.Issues = &falseVal
	}

	if projectsCount > total/2 {
		trueVal := true
		defaults.Projects = &trueVal
	} else {
		falseVal := false
		defaults.Projects = &falseVal
	}

	if deleteBranchCount > total/2 {
		trueVal := true
		defaults.DeleteBranchOnMerge = &trueVal
	} else {
		falseVal := false
		defaults.DeleteBranchOnMerge = &falseVal
	}

	return defaults
}

func writeYAMLConfig(config *Config, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	// Write header comment
	header := `# ownershit Configuration
# Generated configuration for GitHub user repositories
# 
# This file was auto-generated from your GitHub repositories.
# Please review and customize the following sections:
#   - team: Add your team permissions
#   - branches: Configure branch protection rules
#   - default_labels: Add labels to sync across repositories
#   - default_topics: Add topics to apply to all repositories
#
# For more information, see: https://github.com/klauern/ownershit

`
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Encode configuration to YAML
	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}

	return encoder.Close()
}
