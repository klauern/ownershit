package ownershit

//go:generate mockgen -source=github_v4.go -destination=mocks/github_v4_mocks.go -package mocks

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
)

type GraphQLClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

func (c *GitHubClient) SetRepository(id githubv4.ID, wiki, issues, project *bool) error {
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
		HasWikiEnabled:     githubv4.NewBoolean(githubv4.Boolean(*wiki)),
		HasIssuesEnabled:   githubv4.NewBoolean(githubv4.Boolean(*issues)),
		HasProjectsEnabled: githubv4.NewBoolean(githubv4.Boolean(*project)),
	}
	log.Debug().
		Interface("mutation", mutation).
		Interface("input", input).
		Interface("repositoryID", id).
		Msg("GitHubClient.SetRepository()")
	err := c.Graph.Mutate(c.Context, &mutation, input, nil)
	if err != nil {
		log.Err(err).
			Msg("updating repository")
		return err
	}
	return nil
}

func (c *GitHubClient) SetBranchRules(
	id githubv4.ID,
	branchPattern *string,
	approverCount *int,
	requireCodeOwners,
	requiresApprovingReviews *bool) error {
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
		Pattern:                      *githubv4.NewString(githubv4.String(*branchPattern)),
		RequiresApprovingReviews:     githubv4.NewBoolean(githubv4.Boolean(*requiresApprovingReviews)),
		RequiredApprovingReviewCount: githubv4.NewInt(githubv4.Int(*approverCount)),
		RequiresCodeOwnerReviews:     githubv4.NewBoolean(githubv4.Boolean(*requireCodeOwners)),
	}
	log.Debug().
		Interface("mutation", mutation).
		Interface("input", input).
		Interface("repositoryID", id).
		Msg("GitHubClient.SetBranchRules()")
	err := c.Graph.Mutate(c.Context, &mutation, input, nil)
	if err != nil {
		log.Err(err).
			Msg("setting branch rules")
		return err
	}
	return nil
}

type GetRepoQuery struct {
	Repository struct {
		ID                 githubv4.ID
		HasWikiEnabled     githubv4.Boolean
		HasIssuesEnabled   githubv4.Boolean
		HasProjectsEnabled githubv4.Boolean
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func (c *GitHubClient) GetRepository(name, owner *string) (githubv4.ID, error) {
	query := &GetRepoQuery{}
	err := c.Graph.Query(c.Context, query, map[string]interface{}{
		"owner": githubv4.String(*owner),
		"name":  githubv4.String(*name),
	})
	if err != nil {
		log.Err(err).
			Str("owner", *owner).
			Str("name", *name).
			Msg("error retrieving repository")
		return nil, err
	}
	log.Info().
		Str("repository", fmt.Sprintf("%s/%s", *owner, *name)).
		Bool("wiki", bool(query.Repository.HasWikiEnabled)).
		Bool("issues", bool(query.Repository.HasIssuesEnabled)).
		Bool("project", bool(query.Repository.HasProjectsEnabled)).
		Msg("get repository results")
	return query.Repository.ID, nil
}
