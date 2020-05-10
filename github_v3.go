package ownershit

//go:generate mockgen -source=github_v3.go -destination mocks/github_v3_mocks.go -package mocks

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/v31/github"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	Teams   TeamsService
	Graph   GraphQLClient
	V3      *github.Client
	V4      *githubv4.Client
	Context context.Context
}

type TeamsService interface {
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error)
	AddTeamRepoBySlug(ctx context.Context, org, slug, owner, repo string, opts *github.TeamAddTeamRepoOptions) (*github.Response, error)
}

func NewGitHubClient(ctx context.Context, staticToken string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: staticToken})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	clientV4 := githubv4.NewClient(tc)

	return &GitHubClient{
		Teams:   client.Teams,
		V3:      client,
		V4:      clientV4,
		Graph:   clientV4,
		Context: ctx,
	}
}

func (c *GitHubClient) AddPermissions(repo string, organization string, perm Permissions) error {
	resp, err := c.Teams.AddTeamRepoBySlug(c.Context, organization, perm.Team, organization, repo, &github.TeamAddTeamRepoOptions{Permission: string(perm.Level)})
	if err != nil {
		log.Err(err).
			Str("team", perm.Team).
			Str("repo", repo).
			Str("response-status", resp.Status).
			Msg("error adding team as collaborator to repo")
		resp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Err(err).
				Msg("unable to read response body")
			return fmt.Errorf("unable to read response body: %w", err)
		}
		log.Debug().
			Str("response-body", string(resp))
	}
	log.Info().Fields(map[string]interface{}{
		"code": resp.StatusCode,
	})

	return nil
}

func (c *GitHubClient) GetTeamSlug(settings *PermissionsSettings, team *Permissions) error {
	t, _, err := c.Teams.GetTeamBySlug(c.Context, settings.Organization, team.Team)
	if err != nil {
		return fmt.Errorf("unable to get Team from organization: %w", err)
	}
	fmt.Printf("Team: %v ID: %v\n", team.Team, *t.ID)
	team.ID = *t.ID
	return nil
}
