package ownershit

//go:generate mockgen -source=import.go -destination=mocks/import_mocks.go -package mocks

import (
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/rs/zerolog/log"
)

// ImportRepositoryConfig extracts repository configuration from GitHub APIs
// ImportRepositoryConfig fetches configuration for the given GitHub repository and returns it as a PermissionsSettings.
//
// It retrieves repository metadata, branch protection rules, team permissions, and labels from GitHub and assembles
// them into a PermissionsSettings where Organization is set to owner and Repositories contains a single Repository for repo.
// If fetching team permissions fails the error is logged and an empty team permissions list is used; failures to fetch
// repository details, branch protection rules, or labels are returned as errors.
// ImportRepositoryConfig extracts repository configuration from GitHub APIs.
//
// If relaxTeamErrors is true, failures when fetching team permissions are logged and
// an empty permissions list is used. If false, such failures are returned as errors.
func ImportRepositoryConfig(owner, repo string, client *GitHubClient, relaxTeamErrors bool) (*PermissionsSettings, error) {
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
		if relaxTeamErrors {
			log.Warn().
				Str("owner", owner).
				Str("repo", repo).
				Err(err).
				Msg("Failed to get team permissions, continuing with empty team permissions")
			teamPermissions = []*Permissions{}
		} else {
			return nil, fmt.Errorf("failed to get team permissions: %w", err)
		}
	}

	// Get branch protection rules
	branchPermissions, err := getBranchProtectionRules(client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch protection rules: %w", err)
	}

	// Get repository labels
	repoLabels, err := getRepositoryLabels(client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository labels: %w", err)
	}

	// Create PermissionsSettings structure
	config := &PermissionsSettings{
		Organization:      &owner,
		BranchPermissions: *branchPermissions,
		TeamPermissions:   teamPermissions,
		Repositories: []*Repository{
			{
				Name:                  &repo,
				Wiki:                  repoDetails.Wiki,
				Issues:                repoDetails.Issues,
				Projects:              repoDetails.Projects,
				DefaultBranch:         repoDetails.DefaultBranch,
				Private:               repoDetails.Private,
				Archived:              repoDetails.Archived,
				Template:              repoDetails.Template,
				Description:           repoDetails.Description,
				Homepage:              repoDetails.Homepage,
				DeleteBranchOnMerge:   repoDetails.DeleteBranchOnMerge,
				HasDiscussionsEnabled: repoDetails.HasDiscussionsEnabled,
			},
		},
		DefaultLabels: repoLabels,
	}

	return config, nil
}

// repositoryDetails holds the basic repository configuration.
type repositoryDetails struct {
	Wiki                  *bool
	Issues                *bool
	Projects              *bool
	DefaultBranch         *string
	Private               *bool
	Archived              *bool
	Template              *bool
	Description           *string
	Homepage              *string
	DeleteBranchOnMerge   *bool
	HasDiscussionsEnabled *bool
}

// getRepositoryDetails retrieves basic repository settings via GitHub v3 API.
func getRepositoryDetails(client *GitHubClient, owner, repo string) (*repositoryDetails, error) {
	log.Debug().
		Str("owner", owner).
		Str("repo", repo).
		Msg("fetching repository details")

	repoInfo, _, err := client.Repositories.Get(client.Context, owner, repo)
	if err != nil {
		return nil, NewGitHubAPIError(0, "get repository", repo, "failed to fetch repository details", err)
	}

	details := &repositoryDetails{
		Wiki:                  repoInfo.HasWiki,
		Issues:                repoInfo.HasIssues,
		Projects:              repoInfo.HasProjects,
		DefaultBranch:         repoInfo.DefaultBranch,
		Private:               repoInfo.Private,
		Archived:              repoInfo.Archived,
		Template:              repoInfo.IsTemplate,
		Description:           repoInfo.Description,
		Homepage:              repoInfo.Homepage,
		DeleteBranchOnMerge:   repoInfo.DeleteBranchOnMerge,
		HasDiscussionsEnabled: repoInfo.HasDiscussions,
	}

	log.Debug().
		Interface("details", details).
		Msg("repository details retrieved")

	return details, nil
}

// getTeamPermissions retrieves team permissions for the repository.
func getTeamPermissions(client *GitHubClient, owner, repo string) ([]*Permissions, error) {
	log.Debug().
		Str("owner", owner).
		Str("repo", repo).
		Msg("fetching team permissions")

	teams, _, err := client.Repositories.ListTeams(client.Context, owner, repo, nil)
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

// convertPermissionLevel converts GitHub API permission strings to ownershit permission levels.
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

// getBranchProtectionRules retrieves branch protection configuration.
//
//nolint:unparam // keep error return for API symmetry and future use.
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
		protection, _, err = client.Repositories.GetBranchProtection(client.Context, owner, repo, branch)
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
	// TODO: accept repoInfo or merge flags as params to avoid a second API call
	repoInfo, _, err := client.Repositories.Get(client.Context, owner, repo)
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

// convertBranchProtection converts GitHub branch protection to ownershit format.
func convertBranchProtection(protection *github.Protection) *BranchPermissions {
	if protection == nil {
		// Return explicit false values for all boolean fields when no protection exists
		falseVal := false
		return &BranchPermissions{
			RequireStatusChecks:   &falseVal,
			RequireUpToDateBranch: &falseVal,
			RestrictPushes:        &falseVal,
		}
	}

	perms := &BranchPermissions{}
	mapPRReviews(protection, perms)
	mapStatusChecks(protection, perms)
	mapAdminEnforcement(protection, perms)
	mapRestrictions(protection, perms)
	mapAdvancedProtection(protection, perms)
	return perms
}

func mapPRReviews(protection *github.Protection, perms *BranchPermissions) {
	if protection.RequiredPullRequestReviews == nil {
		return
	}
	trueVal := true
	perms.RequirePullRequestReviews = &trueVal
	perms.ApproverCount = &protection.RequiredPullRequestReviews.RequiredApprovingReviewCount
	perms.RequireCodeOwners = &protection.RequiredPullRequestReviews.RequireCodeOwnerReviews
}

func mapStatusChecks(protection *github.Protection, perms *BranchPermissions) {
	if protection.RequiredStatusChecks != nil {
		trueVal := true
		perms.RequireStatusChecks = &trueVal
		perms.RequireUpToDateBranch = &protection.RequiredStatusChecks.Strict
		if protection.RequiredStatusChecks.Contexts != nil && len(*protection.RequiredStatusChecks.Contexts) > 0 {
			perms.StatusChecks = make([]string, len(*protection.RequiredStatusChecks.Contexts))
			copy(perms.StatusChecks, *protection.RequiredStatusChecks.Contexts)
		}
		return
	}
	// Explicitly set to false when status checks are disabled
	falseVal := false
	perms.RequireStatusChecks = &falseVal
	perms.RequireUpToDateBranch = &falseVal
}

func mapAdminEnforcement(protection *github.Protection, perms *BranchPermissions) {
	if protection.EnforceAdmins != nil {
		perms.EnforceAdmins = &protection.EnforceAdmins.Enabled
	}
}

func mapRestrictions(protection *github.Protection, perms *BranchPermissions) {
	if protection.Restrictions == nil {
		falseVal := false
		perms.RestrictPushes = &falseVal
		return
	}
	trueVal := true
	perms.RestrictPushes = &trueVal

	var allowlist []string
	for _, team := range protection.Restrictions.Teams {
		if team.Slug != nil {
			allowlist = append(allowlist, *team.Slug)
		}
	}
	for _, user := range protection.Restrictions.Users {
		if user.Login != nil {
			allowlist = append(allowlist, *user.Login)
		}
	}
	perms.PushAllowlist = allowlist
}

func mapAdvancedProtection(protection *github.Protection, perms *BranchPermissions) {
	if protection.RequiredConversationResolution != nil {
		perms.RequireConversationResolution = &protection.RequiredConversationResolution.Enabled
	}
	if protection.RequireLinearHistory != nil {
		perms.RequireLinearHistory = &protection.RequireLinearHistory.Enabled
	}
	if protection.AllowForcePushes != nil {
		perms.AllowForcePushes = &protection.AllowForcePushes.Enabled
	}
	if protection.AllowDeletions != nil {
		perms.AllowDeletions = &protection.AllowDeletions.Enabled
	}
}

// getRepositoryLabels retrieves all labels for the repository.
func getRepositoryLabels(client *GitHubClient, owner, repo string) ([]RepoLabel, error) {
	log.Debug().
		Str("owner", owner).
		Str("repo", repo).
		Msg("fetching repository labels")

	const defaultPageSize = 100
	opt := &github.ListOptions{PerPage: defaultPageSize}
	var allLabels []RepoLabel

	for {
		labels, resp, err := client.Issues.ListLabels(client.Context, owner, repo, opt)
		if err != nil {
			return nil, NewGitHubAPIError(0, "list labels", repo, "failed to fetch repository labels", err)
		}

		// Convert GitHub labels to RepoLabel format
		for _, label := range labels {
			// Guard against unexpected nil pointers from API to avoid panics.
			if label.Name == nil || label.Color == nil {
				log.Warn().Interface("label", label).Msg("skipping label with missing name or color")
				continue
			}
			repoLabel := RepoLabel{
				Name:  *label.Name,
				Color: *label.Color,
			}

			// Handle optional fields
			if label.Description != nil {
				repoLabel.Description = *label.Description
			}

			allLabels = append(allLabels, repoLabel)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.Debug().
		Int("labelCount", len(allLabels)).
		Interface("labels", allLabels).
		Msg("repository labels retrieved")

	return allLabels, nil
}
