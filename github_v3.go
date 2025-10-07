package ownershit

//go:generate mockgen -source=github_v3.go -destination mocks/github_v3_mocks.go -package mocks

import (
	"context"
	"errors"
	"fmt"
	"net/http/httputil"
	"sort"
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
	ListAllTopics(ctx context.Context, owner, repo string) ([]string, *github.Response, error)
	ReplaceAllTopics(ctx context.Context, owner, repo string, topics []string) ([]string, *github.Response, error)
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
	if perm == nil || perm.Team == nil || perm.Level == nil {
		return fmt.Errorf("%w", ErrInvalidPermissions)
	}
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
			if resp.Response != nil && log.Debug().Enabled() {
				dumped, _ := httputil.DumpResponse(resp.Response, false)
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
		if resp != nil && resp.Response != nil && log.Debug().Enabled() {
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

	log.Info().Fields(map[string]interface{}{"code": resp.StatusCode}).Msg("Updated repository settings")

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
		if resp != nil && resp.Response != nil && log.Debug().Enabled() {
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

	protection := buildProtectionRequest(perms)

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
		log.Err(err).
			Str("org", org).
			Str("repo", repo).
			Str("branch", branch).
			Int("statusCode", statusCode).
			Str("operation", "updateBranchProtection").
			Msg("Error updating branch protection via REST API")
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

// buildProtectionRequest constructs a github.ProtectionRequest based on the given permissions.
func buildProtectionRequest(perms *BranchPermissions) *github.ProtectionRequest {
	pr := &github.ProtectionRequest{}

	if r := buildPullRequestReviews(perms); r != nil {
		pr.RequiredPullRequestReviews = r
	}
	if rs := buildRequiredStatusChecks(perms); rs != nil {
		pr.RequiredStatusChecks = rs
	}
	if perms.EnforceAdmins != nil {
		pr.EnforceAdmins = *perms.EnforceAdmins
	}
	if rest := buildRestrictions(perms); rest != nil {
		pr.Restrictions = rest
	}
	applyAdvancedProtectionOptions(perms, pr)
	return pr
}

func buildPullRequestReviews(perms *BranchPermissions) *github.PullRequestReviewsEnforcementRequest {
	if perms.RequirePullRequestReviews == nil || !*perms.RequirePullRequestReviews {
		return nil
	}
	reviews := &github.PullRequestReviewsEnforcementRequest{
		RequiredApprovingReviewCount: 1, // Default
	}
	if perms.ApproverCount != nil {
		reviews.RequiredApprovingReviewCount = *perms.ApproverCount
	}
	if perms.RequireCodeOwners != nil {
		reviews.RequireCodeOwnerReviews = *perms.RequireCodeOwners
	}
	return reviews
}

func buildRequiredStatusChecks(perms *BranchPermissions) *github.RequiredStatusChecks {
	if perms.RequireStatusChecks == nil || !*perms.RequireStatusChecks {
		return nil
	}
	statusChecks := &github.RequiredStatusChecks{Strict: false}
	if perms.RequireUpToDateBranch != nil {
		statusChecks.Strict = *perms.RequireUpToDateBranch
	}
	if len(perms.StatusChecks) > 0 {
		statusChecks.Contexts = &perms.StatusChecks
	}
	return statusChecks
}

func buildRestrictions(perms *BranchPermissions) *github.BranchRestrictionsRequest {
	if perms.RestrictPushes == nil || !*perms.RestrictPushes {
		return nil
	}
	if len(perms.PushAllowlist) > 0 {
		restrictions := &github.BranchRestrictionsRequest{}
		// Note: Ideally resolve team/user IDs; we treat entries as team slugs/usernames.
		restrictions.Teams = perms.PushAllowlist
		return restrictions
	}
	// else: leave Restrictions nil to avoid invalid empty restrictions
	return nil
}

func applyAdvancedProtectionOptions(perms *BranchPermissions, pr *github.ProtectionRequest) {
	if perms.RequireLinearHistory != nil {
		pr.RequireLinearHistory = perms.RequireLinearHistory
	}
	if perms.RequireConversationResolution != nil {
		pr.RequiredConversationResolution = perms.RequireConversationResolution
	}
	if perms.AllowForcePushes != nil {
		pr.AllowForcePushes = perms.AllowForcePushes
	}
	if perms.AllowDeletions != nil {
		pr.AllowDeletions = perms.AllowDeletions
	}
}

// SyncLabels updates the list of labels that can be used on issues within a repository.
//
//nolint:gocyclo // multi-step label sync handles create/edit/delete flows.
func (c *GitHubClient) SyncLabels(org, repo string, labels []RepoLabel) error {
	ghLabels, resp, err := c.Issues.ListLabels(c.Context, org, repo, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return fmt.Errorf("listing labels for %v/%v: %w", org, repo, err)
	}
	if resp != nil && resp.Response != nil && log.Debug().Enabled() {
		dumpedResp, _ := httputil.DumpResponse(resp.Response, false)
		if len(dumpedResp) > 0 {
			log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
		}
	}
	var edits []struct {
		label   RepoLabel
		oldName string
	}
	var creates []RepoLabel
	for _, v := range labels {
		label := findLabel(v, ghLabels)
		if label != nil {
			edits = append(edits, struct {
				label   RepoLabel
				oldName string
			}{v, *label.Name})
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
		if resp != nil && resp.Response != nil && log.Debug().Enabled() {
			dumpedResp, _ := httputil.DumpResponse(resp.Response, false)
			if len(dumpedResp) > 0 {
				log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
			}
		}
	}

	for _, edit := range edits {
		log.Debug().Msgf("Editing existing label %v", edit.label.Name)
		_, resp, err := c.Issues.EditLabel(c.Context, org, repo, edit.oldName, &github.Label{
			Name:        github.String(fmt.Sprintf("%v %v", edit.label.Emoji, edit.label.Name)),
			Color:       &edit.label.Color,
			Description: &edit.label.Description,
		})
		if err != nil {
			return NewGitHubAPIError(0, "edit label", fmt.Sprintf("%s/%s", org, repo), fmt.Sprintf("failed to edit label %s", edit.label.Name), err)
		}
		if resp != nil && resp.Response != nil && log.Debug().Enabled() {
			dumpedResp, _ := httputil.DumpResponse(resp.Response, false)
			if len(dumpedResp) > 0 {
				log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
			}
		}
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

// SyncTopics updates repository topics, either additively or by replacement.
// When additive is true, new topics are merged with existing ones.
// When additive is false, all topics are replaced with the new set.
//
//nolint:gocyclo // Multi-step sync operation with validation
func (c *GitHubClient) SyncTopics(org, repo string, topics []string, additive bool) error {
	var finalTopics []string
	if additive {
		// Get existing topics using ListAllTopics (Repositories.Get does not return topics)
		existingTopics, resp, err := c.Repositories.ListAllTopics(c.Context, org, repo)
		if err != nil {
			return fmt.Errorf("listing topics for %v/%v: %w", org, repo, err)
		}
		if resp != nil && resp.Response != nil && log.Debug().Enabled() {
			dumpedResp, _ := httputil.DumpResponse(resp.Response, false)
			if len(dumpedResp) > 0 {
				log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
			}
		}

		// Merge existing and new topics, removing duplicates
		topicSet := make(map[string]bool)
		for _, topic := range existingTopics {
			topicSet[topic] = true
		}
		for _, topic := range topics {
			topicSet[topic] = true
		}
		for topic := range topicSet {
			finalTopics = append(finalTopics, topic)
		}
		// Sort for deterministic output in logs and tests
		sort.Strings(finalTopics)
		log.Debug().
			Strs("existing", existingTopics).
			Strs("new", topics).
			Strs("merged", finalTopics).
			Msg("merging topics")
	} else {
		// Replace mode: use only the new topics
		if len(topics) == 0 {
			log.Info().
				Str("repository", fmt.Sprintf("%s/%s", org, repo)).
				Msg("skipping topic replacement: empty topics list provided (use explicit empty list if you want to clear all topics)")
			return nil
		}
		finalTopics = topics
		log.Debug().
			Strs("new", topics).
			Msg("replacing topics")
	}

	// Update repository topics
	_, resp, err := c.Repositories.ReplaceAllTopics(c.Context, org, repo, finalTopics)
	if err != nil {
		return NewGitHubAPIError(0, "replace topics", fmt.Sprintf("%s/%s", org, repo), "failed to update repository topics", err)
	}
	if resp != nil && resp.Response != nil && log.Debug().Enabled() {
		dumpedResp, _ := httputil.DumpResponse(resp.Response, false)
		if len(dumpedResp) > 0 {
			log.Debug().Str("response-body", string(dumpedResp)).Msg("response body")
		}
	}

	log.Info().
		Str("repository", fmt.Sprintf("%s/%s", org, repo)).
		Strs("topics", finalTopics).
		Bool("additive", additive).
		Msg("successfully updated repository topics")

	return nil
}

// Package-level errors for consistent wrapping and lint compliance.
var (
	ErrInvalidPermissions = errors.New("invalid permissions: team and level must be provided")
)
