package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v29/github"
	"github.com/shurcooL/githubv4"
	"io/ioutil"
)

func addPermissions(client *github.Client, ctx context.Context, repo string, organization string, perm Permissions) error {
	resp, err := client.Teams.AddTeamRepoBySlug(ctx, organization, perm.Team, organization, repo, &github.TeamAddTeamRepoOptions{Permission: string(perm.Level)})
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

func getTeamSlug(client *github.Client, ctx context.Context, settings *PermissionsSettings, team *Permissions) error {
	t, _, err := client.Teams.GetTeamBySlug(ctx, settings.Organization, team.Team)
	if err != nil {
		return fmt.Errorf("unable to get Team from organization: %w", err)
	}
	fmt.Printf("Team: %v ID: %v\n", team.Team, *t.ID)
	team.ID = *t.ID
	return nil
}

func GetRepository(client *githubv4.Client, ctx context.Context, name, owner string) (*githubv4.ID, error) {
	var query struct {
		Repository struct {
			ID                 githubv4.ID
			HasWikiEnabled     githubv4.Boolean
			HasIssuesEnabled   githubv4.Boolean
			HasProjectsEnabled githubv4.Boolean
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	err := client.Query(ctx, &query, map[string]interface{}{
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

func SetRepository(client *githubv4.Client, ctx context.Context, id githubv4.ID, wiki, issues, project bool) error {
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
	err := client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		fmt.Printf("error updating repository ID %v: %v", id, err)
		return err
	}
	return nil
}

func SetBranchRules(client *githubv4.Client, ctx context.Context, id githubv4.ID, branchPattern string, approverCount int, requireCodeOwners, requiresApprovingReviews bool) error {
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
	err := client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		fmt.Printf("error updating repository branch protection rule %v: %v", id, err)
		return err
	}
	return nil
}

