package ownershit

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/rs/zerolog/log"
)

func (c *GitHubClient) SetDefaultBranch(ctx context.Context, owner, repo, defaultBranchName string) error {
	r := &github.Repository{
		DefaultBranch: github.String(defaultBranchName),
	}
	repositoryResp, _, err := c.Repositories.Edit(ctx, owner, repo, r)
	if err != nil {
		return fmt.Errorf("setting repo to default branch %v: %w", defaultBranchName, err)
	}
	log.Info().Interface("repositoryResponse", repositoryResp).Msg("success updating default push branch")
	return nil
}

func (c *GitHubClient) SetRepositoryDefaults(
	ctx context.Context,
	owner, repo string,
	wiki, issues, projects bool) error {
	r := &github.Repository{
		HasWiki:     github.Bool(wiki),
		HasIssues:   github.Bool(issues),
		HasProjects: github.Bool(projects),
	}
	_, _, err := c.Repositories.Edit(ctx, owner, repo, r)
	if err != nil {
		return fmt.Errorf("setting defaults for %v/%v: %w", owner, repo, err)
	}

	return nil
}

// hub api repos/rm-you/test_branch_change_api -X PATCH -F name="test_branch_change_api" -F default_branch="new_branch"
// curl -s -H "Authorization: token $GITHUB_TOKEN" https://github.com/api/v3/repos/rm-you/test_branch_change_api -X \
// PATCH --data '{"name": "test_branch_change_api", "default_branch": "new_branch"}'
