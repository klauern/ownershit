package ownershit

//go:generate mockgen -source=import.go -destination=mocks/import_mocks.go -package mocks

import (
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/rs/zerolog/log"
)

// ImportRepositoryConfig extracts repository configuration from GitHub APIs
// and returns it in the PermissionsSettings format used by ownershit.
func ImportRepositoryConfig(owner, repo string, client *GitHubClient) (*PermissionsSettings, error) {
	log.Info().
		Str("owner", owner).
		Str("repo", repo).
		Msg("importing repository configuration from GitHub")

	// Get repository details
	repoDetails, err := getRepositoryDetails(client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository details: %w", err)
	}

	// Get team permissions
	teamPermissions, err := getTeamPermissions(client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get team permissions: %w", err)
	}

	// Get branch protection rules
	branchPermissions, err := getBranchProtectionRules(client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch protection rules: %w", err)
	}

	// Create PermissionsSettings structure
	config := &PermissionsSettings{
		Organization:      &owner,
		BranchPermissions: *branchPermissions,
		TeamPermissions:   teamPermissions,
		Repositories: []*Repository{
			{
				Name:     &repo,
				Wiki:     repoDetails.Wiki,
				Issues:   repoDetails.Issues,
				Projects: repoDetails.Projects,
			},
		},
		DefaultLabels: []RepoLabel{}, // Empty for now, could be populated later
	}

	return config, nil
}

// repositoryDetails holds the basic repository configuration.
type repositoryDetails struct {
	Wiki     *bool
	Issues   *bool
	Projects *bool
}

// getRepositoryDetails retrieves basic repository settings via GitHub v3 API.
func getRepositoryDetails(client *GitHubClient, owner, repo string) (*repositoryDetails, error) {
	log.Debug().
		Str("owner", owner).
		Str("repo", repo).
		Msg("fetching repository details")

	repoInfo, _, err := client.v3.Repositories.Get(client.Context, owner, repo)
	if err != nil {
		return nil, NewGitHubAPIError(0, "get repository", repo, "failed to fetch repository details", err)
	}

	details := &repositoryDetails{
		Wiki:     repoInfo.HasWiki,
		Issues:   repoInfo.HasIssues,
		Projects: repoInfo.HasProjects,
	}

	log.Debug().
		Interface("details", details).
		Msg("repository details retrieved")

	return details, nil
}

// getTeamPermissions retrieves team permissions for the repository
func getTeamPermissions(client *GitHubClient, owner, repo string) ([]*Permissions, error) {
	log.Debug().
		Str("owner", owner).
		Str("repo", repo).
		Msg("fetching team permissions")

	teams, _, err := client.v3.Repositories.ListTeams(client.Context, owner, repo, nil)
	if err != nil {
		return nil, NewGitHubAPIError(0, "list teams", repo, "failed to fetch team permissions", err)
	}

	var permissions []*Permissions
	for _, team := range teams {
		perm := &Permissions{
			Team:  team.Name,
			Level: convertPermissionLevel(team.Permission),
		}
		permissions = append(permissions, perm)
	}

	log.Debug().
		Int("teamCount", len(permissions)).
		Interface("permissions", permissions).
		Msg("team permissions retrieved")

	return permissions, nil
}

// convertPermissionLevel converts GitHub API permission strings to ownershit permission levels
func convertPermissionLevel(ghPermission *string) *string {
	if ghPermission == nil {
		return nil
	}

	var level string
	switch *ghPermission {
	case string(Admin):
		level = string(Admin)
	case "push", "write":
		level = string(Write)
	case string(Read), "read":
		level = string(Read)
	default:
		level = string(Read) // default to read access
	}

	return &level
}

// getBranchProtectionRules retrieves branch protection configuration
func getBranchProtectionRules(client *GitHubClient, owner, repo string) (*BranchPermissions, error) {
	log.Debug().
		Str("owner", owner).
		Str("repo", repo).
		Msg("fetching branch protection rules")

	// Try to get branch protection for main branch first, then master
	branches := []string{"main", "master"}
	var protection *github.Protection
	var err error

	for _, branch := range branches {
		protection, _, err = client.v3.Repositories.GetBranchProtection(client.Context, owner, repo, branch)
		if err == nil {
			log.Debug().
				Str("branch", branch).
				Msg("found branch protection rules")
			break
		}
	}

	// If no protection found, return default settings
	if protection == nil {
		log.Debug().Msg("no branch protection rules found, using defaults")
		return &BranchPermissions{}, nil
	}

	// Convert GitHub protection to ownershit format
	branchPerms := convertBranchProtection(protection)

	// Get merge settings from repository details
	repoInfo, _, err := client.v3.Repositories.Get(client.Context, owner, repo)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get repository merge settings")
	} else {
		branchPerms.AllowMergeCommit = repoInfo.AllowMergeCommit
		branchPerms.AllowSquashMerge = repoInfo.AllowSquashMerge
		branchPerms.AllowRebaseMerge = repoInfo.AllowRebaseMerge
	}

	log.Debug().
		Interface("branchPermissions", branchPerms).
		Msg("branch protection rules retrieved")

	return branchPerms, nil
}

// convertBranchProtection converts GitHub branch protection to ownershit format
func convertBranchProtection(protection *github.Protection) *BranchPermissions {
	if protection == nil {
		return &BranchPermissions{}
	}

	perms := &BranchPermissions{}

	// Pull request reviews
	if protection.RequiredPullRequestReviews != nil {
		trueVal := true
		perms.RequirePullRequestReviews = &trueVal
		perms.ApproverCount = &protection.RequiredPullRequestReviews.RequiredApprovingReviewCount
		perms.RequireCodeOwners = &protection.RequiredPullRequestReviews.RequireCodeOwnerReviews
	}

	// Status checks
	if protection.RequiredStatusChecks != nil {
		trueVal := true
		perms.RequireStatusChecks = &trueVal
		perms.RequireUpToDateBranch = &protection.RequiredStatusChecks.Strict
		if protection.RequiredStatusChecks.Contexts != nil && len(*protection.RequiredStatusChecks.Contexts) > 0 {
			perms.StatusChecks = make([]string, len(*protection.RequiredStatusChecks.Contexts))
			copy(perms.StatusChecks, *protection.RequiredStatusChecks.Contexts)
		}
	}

	// Admin enforcement
	if protection.EnforceAdmins != nil {
		perms.EnforceAdmins = &protection.EnforceAdmins.Enabled
	}

	// Restrictions
	if protection.Restrictions != nil {
		trueVal := true
		perms.RestrictPushes = &trueVal
		// Note: GitHub API doesn't provide push allowlist in the same format
		// This would need additional API calls to get team/user restrictions
	}

	// Advanced settings (these may require additional API calls or GraphQL)
	// For now, we'll leave them as nil since they require more complex queries

	return perms
}
