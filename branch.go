package ownershit

import (
	"context"
	"fmt"
)

func (c *GitHubClient) SetDefaultBranch(ctx context.Context, repo, org, defaultBranchName string) {
	_, err := c.V3.NewRequest("PATCH", fmt.Sprintf("repos/%s/%s", org, repo), struct {
		name          string
		defaultBranch string
	}{
		repo,
		defaultBranchName,
	})
	if err != nil {
		panic(err)
	}
}

// hub api repos/rm-you/test_branch_change_api -X PATCH -F name="test_branch_change_api" -F default_branch="new_branch"
// curl -s -H "Authorization: token $GITHUB_TOKEN" https://github.com/api/v3/repos/rm-you/test_branch_change_api -X \
// PATCH --data '{"name": "test_branch_change_api", "default_branch": "new_branch"}'