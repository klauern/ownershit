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
}

// NewGitHubClient creates a new GitHub context using OAuth2.
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
		log.Err(err).
			Str("team", *perm.Team).
			Str("repo", repo).
			Str("response-status", resp.Status).
			Msg("error adding team as collaborator to repo")
		resp, _ := httputil.DumpResponse(resp.Response, true)
		if resp != nil {
			log.Debug().Str("response-body", string(resp))
		}
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
		log.Err(err).
			Str("org", org).
			Str("repo", repo).
			Str("response-status", resp.Status).
			Msg("Error updating repository settings")
		resp, _ := httputil.DumpResponse(resp.Response, true)
		log.Debug().Str("response-body", string(resp))
		return err
	}

	log.Info().Fields(map[string]interface{}{
		"code": resp.StatusCode,
	})

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
			return fmt.Errorf("listing labels for %v/%v: %w", org, repo, err)
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
			return fmt.Errorf("listing labels for %v/%v: %w", org, repo, err)
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
