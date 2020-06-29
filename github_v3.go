package ownershit

//go:generate mockgen -source=github_v3.go -destination mocks/github_v3_mocks.go -package mocks

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/v32/github"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	Teams        TeamsService
	Repositories RepositoriesService
	Graph        GraphQLClient
	V3           *github.Client
	V4           *githubv4.Client
	Context      context.Context
}

// TeamsService is a wrapepr interface for the GitHub V3 API to support mocking and testing.
type TeamsService interface {
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error)
	AddTeamRepoBySlug(ctx context.Context, org, slug, owner, repo string, opts *github.TeamAddTeamRepoOptions) (*github.Response, error)
}

type RepositoriesService interface {
	Edit(ctx context.Context, org, repo string, repository *github.Repository) (*github.Repository, *github.Response, error)
}

func NewGitHubClient(ctx context.Context, staticToken string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: staticToken})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	clientV4 := githubv4.NewClient(tc)

	return &GitHubClient{
		Teams:        client.Teams,
		Repositories: client.Repositories,
		V3:           client,
		V4:           clientV4,
		Graph:        clientV4,
		Context:      ctx,
	}
}

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
		respBody, respErr := ioutil.ReadAll(resp.Body)
		if respErr != nil {
			log.Err(err).
				Msg("unable to read response body")
			return fmt.Errorf("unable to read response body: %w", respErr)
		}
		log.Debug().
			Str("response-body", string(respBody))
		return fmt.Errorf("adding team as collaborator to repo: %w", err)
	}
	log.Info().Int("status-code", resp.StatusCode).Msg("Successfully set repo")
	return nil
}

func (c *GitHubClient) UpdateRepositorySettings(org, repo string, perms *BranchPermissions) error {
	_, resp, err := c.Repositories.Edit(c.Context, org, repo, &github.Repository{
		AllowMergeCommit: github.Bool(*perms.AllowMergeCommit),
		AllowRebaseMerge: github.Bool(*perms.AllowRebaseMerge),
		AllowSquashMerge: github.Bool(*perms.AllowMergeCommit),
	})
	if err != nil {
		log.Err(err).
			Str("org", org).
			Str("repo", repo).
			Str("response-status", resp.Status).
			Msg("Error updating repository settings")
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Err(err).
				Msg("unable to read response body")
			return fmt.Errorf("unable to read response body: %w", err)
		}
		log.Debug().Str("response-body", string(respBody))
	}

	log.Info().Fields(map[string]interface{}{
		"code": resp.StatusCode,
	})

	return nil
}
