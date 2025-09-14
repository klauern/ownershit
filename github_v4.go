package ownershit

//go:generate mockgen -source=github_v4.go -destination=mocks/github_v4_mocks.go -package mocks

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
)

type GraphQLClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

func (c *GitHubClient) SetRepository(id githubv4.ID, wiki, issues, project, discussions, sponsorships *bool) error {
	var mutation struct {
		UpdateRepository struct {
			ClientMutationID githubv4.ID
			Repository       struct {
				ID                     githubv4.ID
				HasWikiEnabled         githubv4.Boolean
				HasIssuesEnabled       githubv4.Boolean
				HasProjectsEnabled     githubv4.Boolean
				HasDiscussionsEnabled  githubv4.Boolean
				HasSponsorshipsEnabled githubv4.Boolean
			}
		} `graphql:"updateRepository(input: $input)"`
	}
	// Start with only the repository ID set
	inputRepo := githubv4.UpdateRepositoryInput{
		RepositoryID: id,
	}
	// Only set fields if corresponding pointers are non-nil
	if wiki != nil {
		inputRepo.HasWikiEnabled = githubv4.NewBoolean(githubv4.Boolean(*wiki))
	}
	if issues != nil {
		inputRepo.HasIssuesEnabled = githubv4.NewBoolean(githubv4.Boolean(*issues))
	}
	if project != nil {
		inputRepo.HasProjectsEnabled = githubv4.NewBoolean(githubv4.Boolean(*project))
	}

	// Only set optional fields if they are not nil
	if discussions != nil {
		inputRepo.HasDiscussionsEnabled = githubv4.NewBoolean(githubv4.Boolean(*discussions))
	}
	if sponsorships != nil {
		inputRepo.HasSponsorshipsEnabled = githubv4.NewBoolean(githubv4.Boolean(*sponsorships))
	}
	log.Debug().
		Interface("mutation", mutation).
		Interface("input", inputRepo).
		Interface("repositoryID", id).
		Msg("GitHubClient.SetRepository()")
	err := c.Graph.Mutate(c.Context, &mutation, inputRepo, nil)
	if err != nil {
		log.Err(err).
			Str("operation", "updateRepository").
			Interface("repositoryID", id).
			Msg("updating repository")
		return NewGitHubAPIError(0, "update repository", "", "failed to update repository settings", err)
	}
	return nil
}

func (c *GitHubClient) SetBranchRules(
	id githubv4.ID,
	branchPattern *string,
	approverCount *int,
	requireCodeOwners,
	requiresApprovingReviews *bool,
) error {
	// Deprecated: Use SetEnhancedBranchProtection instead
	return c.SetEnhancedBranchProtection(id, "main", &BranchPermissions{
		RequireCodeOwners:         requireCodeOwners,
		ApproverCount:             approverCount,
		RequirePullRequestReviews: requiresApprovingReviews,
	})
}

func (c *GitHubClient) SetEnhancedBranchProtection(id githubv4.ID, branchPattern string, perms *BranchPermissions) error {
	var mutation struct {
		CreateBranchProtectionRule struct {
			ClientMutationID     githubv4.ID
			BranchProtectionRule struct {
				ID                           githubv4.ID
				Pattern                      githubv4.String
				RequiresApprovingReviews     githubv4.Boolean
				RequiredApprovingReviewCount githubv4.Int
				RequiresCodeOwnerReviews     githubv4.Boolean
				RequiresStatusChecks         githubv4.Boolean
				RequiresStrictStatusChecks   githubv4.Boolean
				RequiredStatusCheckContexts  []githubv4.String
				Repository                   struct {
					ID githubv4.ID
				}
			}
		} `graphql:"createBranchProtectionRule(input: $input)"`
	}

	// Build the input with all available options
	input := githubv4.CreateBranchProtectionRuleInput{
		RepositoryID: id,
		Pattern:      *githubv4.NewString(githubv4.String(branchPattern)),
	}

	// Set required fields and handle existing functionality
	if perms.RequirePullRequestReviews != nil {
		input.RequiresApprovingReviews = githubv4.NewBoolean(githubv4.Boolean(*perms.RequirePullRequestReviews))
	}

	if perms.ApproverCount != nil {
		if *perms.ApproverCount > 2147483647 || *perms.ApproverCount < -2147483648 {
			log.Warn().Int("approverCount", *perms.ApproverCount).Msg("approverCount exceeds int32 range, using max value")
			input.RequiredApprovingReviewCount = githubv4.NewInt(githubv4.Int(2147483647))
		} else {
			input.RequiredApprovingReviewCount = githubv4.NewInt(githubv4.Int(int32(*perms.ApproverCount)))
		}
	}

	if perms.RequireCodeOwners != nil {
		input.RequiresCodeOwnerReviews = githubv4.NewBoolean(githubv4.Boolean(*perms.RequireCodeOwners))
	}

	// Set new advanced protection features (using fields available in GitHub GraphQL API)
	if perms.RequireStatusChecks != nil {
		input.RequiresStatusChecks = githubv4.NewBoolean(githubv4.Boolean(*perms.RequireStatusChecks))
	}

	if perms.RequireUpToDateBranch != nil {
		input.RequiresStrictStatusChecks = githubv4.NewBoolean(githubv4.Boolean(*perms.RequireUpToDateBranch))
	}

	if len(perms.StatusChecks) > 0 {
		statusChecks := make([]githubv4.String, len(perms.StatusChecks))
		for i, check := range perms.StatusChecks {
			statusChecks[i] = githubv4.String(check)
		}
		input.RequiredStatusCheckContexts = &statusChecks
	}

	// Note: Advanced features like EnforceAdmins, RestrictPushes, etc. may not be available
	// in the current githubv4 package version. These will be implemented via REST API fallback.
	// For now, log what features are being requested but not applied via GraphQL.

	if perms.EnforceAdmins != nil && *perms.EnforceAdmins {
		log.Info().Msg("EnforceAdmins requested - will be handled via REST API fallback")
	}

	if perms.RestrictPushes != nil && *perms.RestrictPushes {
		log.Info().Msg("RestrictPushes requested - will be handled via REST API fallback")
	}

	if perms.RequireConversationResolution != nil && *perms.RequireConversationResolution {
		log.Info().Msg("RequireConversationResolution requested - will be handled via REST API fallback")
	}

	if perms.RequireLinearHistory != nil && *perms.RequireLinearHistory {
		log.Info().Msg("RequireLinearHistory requested - will be handled via REST API fallback")
	}

	if perms.AllowForcePushes != nil {
		log.Info().Bool("allowForcePushes", *perms.AllowForcePushes).Msg("AllowForcePushes requested - will be handled via REST API fallback")
	}

	if perms.AllowDeletions != nil {
		log.Info().Bool("allowDeletions", *perms.AllowDeletions).Msg("AllowDeletions requested - will be handled via REST API fallback")
	}

	log.Debug().
		Interface("mutation", mutation).
		Interface("input", input).
		Interface("repositoryID", id).
		Str("pattern", branchPattern).
		Msg("GitHubClient.SetEnhancedBranchProtection() - Before GraphQL mutation")

	err := c.Graph.Mutate(c.Context, &mutation, input, nil)
	if err != nil {
		// Check if this is a "Name already protected" error - this is expected and we should fall back to REST API
		if strings.Contains(err.Error(), "Name already protected") {
			log.Info().
				Str("pattern", branchPattern).
				Str("reason", "branch protection rule already exists").
				Msg("GraphQL createBranchProtectionRule failed - will use REST API fallback")
			return NewGitHubAPIError(0, "create branch protection rule", "",
				fmt.Sprintf("branch protection rule already exists for pattern %s, falling back to REST API", branchPattern), err)
		}

		// For other errors, log detailed information
		log.Err(err).
			Str("operation", "createBranchProtectionRule").
			Interface("repositoryID", id).
			Str("pattern", branchPattern).
			Interface("input", input).
			Interface("mutationStruct", mutation).
			Msg("setting enhanced branch protection")
		return NewGitHubAPIError(0, "create branch protection rule", "",
			fmt.Sprintf("failed to create enhanced branch protection rule for pattern %s", branchPattern), err)
	}

	log.Debug().
		Interface("result", mutation).
		Str("pattern", branchPattern).
		Msg("GitHubClient.SetEnhancedBranchProtection() - GraphQL mutation successful")
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
			Str("operation", "getRepository").
			Msg("error retrieving repository")
		return nil, NewRepositoryNotFoundError(*owner, *name, fmt.Errorf("failed to retrieve repository %s/%s: %w", *owner, *name, err))
	}
	log.Info().
		Str("repository", fmt.Sprintf("%s/%s", *owner, *name)).
		Bool("wiki", bool(query.Repository.HasWikiEnabled)).
		Bool("issues", bool(query.Repository.HasIssuesEnabled)).
		Bool("project", bool(query.Repository.HasProjectsEnabled)).
		Msg("get repository results")
	return query.Repository.ID, nil
}

type GetRateLimitQuery struct {
	Viewer struct {
		Login githubv4.String
	}
	RateLimit struct {
		Limit     githubv4.Int
		Cost      githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

func (c *GitHubClient) GetRateLimit() {
	query := &GetRateLimitQuery{}
	err := c.Graph.Query(c.Context, query, nil)
	if err != nil {
		log.Err(err).
			Msg("error retrieving rate limit")
		return
	}
	log.Info().
		Str("login", string(query.Viewer.Login)).
		Int("limit", int(query.RateLimit.Limit)).
		Int("cost", int(query.RateLimit.Cost)).
		Int("remaining", int(query.RateLimit.Remaining)).
		Time("resetAt", query.RateLimit.ResetAt.Time).
		Msg("get rate limit results")
}
