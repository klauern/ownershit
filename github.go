package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v29/github"
	"github.com/shurcooL/githubv4"
	"io/ioutil"
)

type GitHubClient struct {
	V3 *github.Client
	V4 *githubv4.Client
	Context context.Context
}

func (c *GitHubClient) addPermissions(repo string, organization string, perm Permissions) error {
	resp, err := c.V3.Teams.AddTeamRepoBySlug(c.Context, organization, perm.Team, organization, repo, &github.TeamAddTeamRepoOptions{Permission: string(perm.Level)})
	if err != nil {
		fmt.Printf("error adding %v as collaborator to %v: %v\n", perm.Team, repo, resp.Status)
		resp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body: %w", err)
		}
		fmt.Println(string(resp))
	}
	return nil
}

func (c *GitHubClient) getTeamSlug(settings *PermissionsSettings, team *Permissions) error {
	t, _, err := c.V3.Teams.GetTeamBySlug(c.Context, settings.Organization, team.Team)
	if err != nil {
		return fmt.Errorf("unable to get Team from organization: %w", err)
	}
	fmt.Printf("Team: %v ID: %v\n", team.Team, *t.ID)
	team.ID = *t.ID
	return nil
}

func (c *GitHubClient) GetRepository(name, owner string) (*githubv4.ID, error) {
	var query struct {
		Repository struct {
			ID                 githubv4.ID
			HasWikiEnabled     githubv4.Boolean
			HasIssuesEnabled   githubv4.Boolean
			HasProjectsEnabled githubv4.Boolean
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	err := c.V4.Query(c.Context, &query, map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	})
	if err != nil {
		fmt.Printf("error retrieving %s/%s: %v\n", owner, name, err)
		return nil, err
	}
	fmt.Printf("repository %s/%s: wiki:%v,issues:%v,projects:%v\n",
		owner, name, query.Repository.HasWikiEnabled, query.Repository.HasIssuesEnabled, query.Repository.HasProjectsEnabled)
	return &query.Repository.ID, nil
}

func (c *GitHubClient) SetRepository(id githubv4.ID, wiki, issues, project bool) error {
	var mutation struct {
		UpdateRepository struct {
			ClientMutationID githubv4.ID
			Repository       struct {
				ID                 githubv4.ID
				HasWikiEnabled     githubv4.Boolean
				HasIssuesEnabled   githubv4.Boolean
				HasProjectsEnabled githubv4.Boolean
			}
		} `graphql:"updateRepository(input: $input)"`
	}
	input := githubv4.UpdateRepositoryInput{
		RepositoryID:       id,
		HasWikiEnabled:     githubv4.NewBoolean(githubv4.Boolean(wiki)),
		HasIssuesEnabled:   githubv4.NewBoolean(githubv4.Boolean(issues)),
		HasProjectsEnabled: githubv4.NewBoolean(githubv4.Boolean(project)),
	}
	return c.mutate("updating repository", &mutation, input)
}

func (c *GitHubClient) SetBranchRules(id githubv4.ID, branchPattern string, approverCount int, requireCodeOwners, requiresApprovingReviews bool) error {
	var mutation struct {
		CreateBranchProtectionRule struct {
			ClientMutationID     githubv4.ID
			BranchProtectionRule struct {
				RepositoryID                 githubv4.ID
				Pattern                      githubv4.String
				RequiresApprovingReviews     githubv4.Boolean
				RequiresApprovingReviewCount githubv4.Int
				RequiresCodeOwnerReviews     githubv4.Boolean
			}
		} `graphql:"createBranchProtectionRule(input: $input)"`
	}
	input := githubv4.CreateBranchProtectionRuleInput{
		RepositoryID:                 id,
		Pattern:                      *githubv4.NewString(githubv4.String(branchPattern)),
		RequiresApprovingReviews:     githubv4.NewBoolean(githubv4.Boolean(requiresApprovingReviews)),
		RequiredApprovingReviewCount: githubv4.NewInt(githubv4.Int(approverCount)),
		RequiresCodeOwnerReviews:     githubv4.NewBoolean(githubv4.Boolean(requireCodeOwners)),
	}
	return c.mutate("updating repository branch protection", &mutation, input)
}

func (c *GitHubClient) mutate(errString string, mutation interface{}, input githubv4.Input) error {
	err := c.V4.Mutate(c.Context, &mutation, input, nil)
	if err != nil {
		fmt.Printf("error with mutation: %v: %v", errString, err)
		return err
	}
	return nil

}

