package ownershit

//go:generate mockgen -source=github_v3.go -destination mocks/github_v3_mocks.go -package mocks

import (
	"context"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GitHubClient is a wrapper client to both the V3 REST API and V4 GraphQL API, with support for a subset of GitHub services,
// mainly around repository management itself.
type GitHubClient struct {
	Teams        TeamsService
	Repositories RepositoriesService
	Issues       IssuesService
	Graph        GraphQLClient
	v3           *github.Client
	v4           *githubv4.Client
	Context      context.Context
}

// TeamsService is a wrapper interface for the GitHub V3 API to support mocking and testing for the Teams API endpoints.
type TeamsService interface {
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error)
	AddTeamRepoBySlug(ctx context.Context, org, slug, owner, repo string, opts *github.TeamAddTeamRepoOptions) (*github.Response, error)
}

// IssuesService is a wrapper interface for the GitHub V3 REST API for Issues management.  This interface is used for
// mocking and testing.
type IssuesService interface {
	CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error)
	EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error)
	ListLabels(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.Label, *github.Response, error)
	// DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error)
}

// RepositoriesService is a wrapper interface for the GitHub V3 API to support mocking and testing for the Repository API endpoints.
type RepositoriesService interface {
	Edit(ctx context.Context, org, repo string, repository *github.Repository) (*github.Repository, *github.Response, error)
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	ListTeams(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.Team, *github.Response, error)
	GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, *github.Response, error)
}

// NewGitHubClient creates a new GitHub context using OAuth2.
// Deprecated: Use NewSecureGitHubClient() for secure token handling.
func NewGitHubClient(ctx context.Context, staticToken string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: staticToken})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	clientV4 := githubv4.NewClient(tc)

	return &GitHubClient{
		Teams:        client.Teams,
		Repositories: client.Repositories,
		Issues:       client.Issues,
		v3:           client,
		v4:           clientV4,
		Graph:        clientV4,
		Context:      ctx,
	}
}

// NewSecureGitHubClient creates a new GitHub client with validated token from environment.
func NewSecureGitHubClient(ctx context.Context) (*GitHubClient, error) {
	token, err := GetValidatedGitHubToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get validated GitHub token: %w", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	clientV4 := githubv4.NewClient(tc)

	return &GitHubClient{
		Teams:        client.Teams,
		Repositories: client.Repositories,
		Issues:       client.Issues,
		v3:           client,
		v4:           clientV4,
		Graph:        clientV4,
		Context:      ctx,
	}, nil
}

// AddPermissions adds a given team level repository permission.
func (c *GitHubClient) AddPermissions(organization, repo string, perm *Permissions) error {
    resp, err := c.Teams.
        AddTeamRepoBySlug(
            c.Context,
            organization,
            *perm.Team,
            organization,
            repo,
            &github.TeamAddTeamRepoOptions{Permission: *perm.Level})
    if err != nil {
        logEvent := log.Err(err).
            Str("team", *perm.Team).
            Str("repo", repo)
        if resp != nil {
            logEvent = logEvent.Str("response-status", resp.Status)
            if log.Debug().Enabled() {
                dumped, _ := httputil.DumpResponse(resp.Response, true)
                if len(dumped) > 0 {
                    log.Debug().Msg("response body follows")
                    log.Debug().Str("response-body", string(dumped))
                }
            }
        }
        logEvent.Msg("error adding team as collaborator to repo")
        return fmt.Errorf("adding team as collaborator to repo: %w", err)
    }
    log.Info().Int("status-code", resp.StatusCode).Msg("Successfully set repo")
    return nil
}

// UpdateBranchPermissions changes the settings for branch permissions from `perms`.
func (c *GitHubClient) UpdateBranchPermissions(org, repo string, perms *BranchPermissions) error {
	r := &github.Repository{}
	if perms.AllowMergeCommit != nil {
		r.AllowMergeCommit = perms.AllowMergeCommit
	}
	if perms.AllowRebaseMerge != nil {
		r.AllowRebaseMerge = perms.AllowRebaseMerge
	}
	if perms.AllowSquashMerge != nil {
		r.AllowSquashMerge = perms.AllowSquashMerge
	}
    _, resp, err := c.Repositories.Edit(c.Context, org, repo, r)
    if err != nil {
        statusCode := 0
        if resp != nil {
            statusCode = resp.StatusCode
        }
        logEvent := log.Err(err).
            Str("org", org).
            Str("repo", repo).
            Int("statusCode", statusCode).
            Str("operation", "updateRepositorySettings")
        if resp != nil && log.Debug().Enabled() {
            dumped, _ := httputil.DumpResponse(resp.Response, true)
            if len(dumped) > 0 {
                log.Debug().Msg("response body follows")
                log.Debug().Str("response-body", string(dumped))
            }
        }
        logEvent.Msg("Error updating repository settings")
        return NewGitHubAPIError(statusCode, "update repository settings",
            fmt.Sprintf("%s/%s", org, repo), "failed to update repository settings", err)
    }

	log.Info().Fields(map[string]interface{}{
		"code": resp.StatusCode,
	})

	return nil
}

// SetRepositoryAdvancedSettings configures repository settings that require REST API (not available in GraphQL).
func (c *GitHubClient) SetRepositoryAdvancedSettings(org, repo string, deleteBranchOnMerge *bool) error {
	if deleteBranchOnMerge == nil {
		return nil
	}

	r := &github.Repository{
		DeleteBranchOnMerge: deleteBranchOnMerge,
	}

    _, resp, err := c.Repositories.Edit(c.Context, org, repo, r)
    if err != nil {
        statusCode := 0
        if resp != nil {
            statusCode = resp.StatusCode
        }
        logEvent := log.Err(err).
            Str("org", org).
            Str("repo", repo).
            Int("statusCode", statusCode).
            Bool("deleteBranchOnMerge", *deleteBranchOnMerge).
            Str("operation", "setRepositoryAdvancedSettings")
        if resp != nil && log.Debug().Enabled() {
            dumped, _ := httputil.DumpResponse(resp.Response, true)
            if len(dumped) > 0 {
                log.Debug().Msg("response body follows")
                log.Debug().Str("response-body", string(dumped))
            }
        }
        logEvent.Msg("Error setting advanced repository settings")
        return NewGitHubAPIError(statusCode, "set advanced repository settings",
            fmt.Sprintf("%s/%s", org, repo), "failed to set advanced repository settings", err)
    }

	log.Info().
		Str("org", org).
		Str("repo", repo).
		Bool("deleteBranchOnMerge", *deleteBranchOnMerge).
		Int("statusCode", resp.StatusCode).
		Msg("Successfully set advanced repository settings")

	return nil
}

// SetBranchProtectionFallback applies advanced branch protection features via REST API
// that are not available in the GraphQL CreateBranchProtectionRuleInput.
func (c *GitHubClient) SetBranchProtectionFallback(org, repo, branch string, perms *BranchPermissions) error {
	if perms == nil {
		return nil
	}

	// Build protection request for REST API
	protection := &github.ProtectionRequest{}

	// Required pull request reviews
	if perms.RequirePullRequestReviews != nil && *perms.RequirePullRequestReviews {
		reviews := &github.PullRequestReviewsEnforcementRequest{
			RequiredApprovingReviewCount: 1, // Default
		}

		if perms.ApproverCount != nil {
			reviews.RequiredApprovingReviewCount = *perms.ApproverCount
		}

		if perms.RequireCodeOwners != nil {
			reviews.RequireCodeOwnerReviews = *perms.RequireCodeOwners
		}

		protection.RequiredPullRequestReviews = reviews
	}

	// Required status checks
	if perms.RequireStatusChecks != nil && *perms.RequireStatusChecks {
		statusChecks := &github.RequiredStatusChecks{
			Strict: false, // Default
		}

		if perms.RequireUpToDateBranch != nil {
			statusChecks.Strict = *perms.RequireUpToDateBranch
		}

		if len(perms.StatusChecks) > 0 {
			statusChecks.Contexts = &perms.StatusChecks
		}

		protection.RequiredStatusChecks = statusChecks
	}

	// Enforce admins
	if perms.EnforceAdmins != nil {
		protection.EnforceAdmins = *perms.EnforceAdmins
	}

	// Restrict pushes
	if perms.RestrictPushes != nil && *perms.RestrictPushes {
		restrictions := &github.BranchRestrictionsRequest{}

		// Add users/teams from allowlist
		if len(perms.PushAllowlist) > 0 {
			// Note: In a real implementation, you'd need to resolve team names to IDs
			// For now, assume they are team slugs
			restrictions.Teams = perms.PushAllowlist
		}

		protection.Restrictions = restrictions
	}

	// Advanced options only available in REST API
	if perms.RequireLinearHistory != nil {
		protection.RequireLinearHistory = perms.RequireLinearHistory
	}

	if perms.AllowForcePushes != nil {
		protection.AllowForcePushes = perms.AllowForcePushes
	}

	if perms.AllowDeletions != nil {
		protection.AllowDeletions = perms.AllowDeletions
	}

	log.Debug().
		Interface("protection", protection).
		Str("org", org).
		Str("repo", repo).
		Str("branch", branch).
		Msg("Setting branch protection via REST API")

	// Apply protection via REST API
    _, resp, err := c.v3.Repositories.UpdateBranchProtection(c.Context, org, repo, branch, protection)
    if err != nil {
        statusCode := 0
        if resp != nil {
            statusCode = resp.StatusCode
        }
        _ = log.Err(err).
            Str("org", org).
            Str("repo", repo).
            Str("branch", branch).
            Int("statusCode", statusCode).
            Str("operation", "updateBranchProtection")
        return NewGitHubAPIError(statusCode, "update branch protection",
            fmt.Sprintf("%s/%s", org, repo), "failed to set branch protection via REST API", err)
    }

	log.Info().
		Str("org", org).
		Str("repo", repo).
		Str("branch", branch).
		Int("statusCode", resp.StatusCode).
		Msg("Successfully set branch protection via REST API")

	return nil
}

// SyncLabels updates the list of labels that can be used on issues within a repository.
func (c *GitHubClient) SyncLabels(org, repo string, labels []RepoLabel) error {
	ghLabels, resp, err := c.Issues.ListLabels(c.Context, org, repo, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return fmt.Errorf("listing labels for %v/%v: %w", org, repo, err)
	}
	dumpedResp, _ := httputil.DumpResponse(resp.Response, true)
	log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
	var edits []RepoLabel
	var creates []RepoLabel
	for _, v := range labels {
		label := findLabel(v, ghLabels)
		if label != nil {
			v.oldLabel = *label.Name
			edits = append(edits, v)
		} else {
			creates = append(creates, v)
		}
	}

	for _, create := range creates {
		log.Debug().Msgf("Creating label %v", create.Name)
		_, resp, err := c.Issues.CreateLabel(c.Context, org, repo, &github.Label{
			Name:        github.String(fmt.Sprintf("%v %v", create.Emoji, create.Name)),
			Color:       &create.Color,
			Description: &create.Description,
		})
		if err != nil {
			return NewGitHubAPIError(0, "create label", fmt.Sprintf("%s/%s", org, repo), fmt.Sprintf("failed to create label %s", create.Name), err)
		}
		dumpedResp, _ := httputil.DumpResponse(resp.Response, true)
		log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
	}

	for _, edit := range edits {
		log.Debug().Msgf("Editing existing label %v", edit.Name)
		_, resp, err := c.Issues.EditLabel(c.Context, org, repo, edit.oldLabel, &github.Label{
			Name:        github.String(fmt.Sprintf("%v %v", edit.Emoji, edit.Name)),
			Color:       &edit.Color,
			Description: &edit.Description,
		})
		if err != nil {
			return NewGitHubAPIError(0, "edit label", fmt.Sprintf("%s/%s", org, repo), fmt.Sprintf("failed to edit label %s", edit.Name), err)
		}
		dumpedResp, _ := httputil.DumpResponse(resp.Response, true)
		log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
	}

	return nil
}

func findLabel(searchLabel RepoLabel, ghLabels []*github.Label) *github.Label {
	for _, v := range ghLabels {
		if strings.Contains(*v.Name, searchLabel.Name) {
			return v
		}
	}
	return nil
}
