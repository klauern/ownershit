package ownershit

//go:generate mockgen -source=github_v4.go -destination=mocks/github_v4_mocks.go -package mocks

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
)

// GraphQLClient defines the interface for GraphQL operations.
type GraphQLClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

// SetRepository updates repository-level feature flags (wiki, issues, projects,
// discussions, sponsorships) for the given repository ID using the GraphQL API.
// Only non-nil fields are applied.
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

// SetBranchRules configures a subset of branch protection settings for the given
// repository via GraphQL. Deprecated: prefer SetEnhancedBranchProtection.
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

// ErrNilBranchPermissions is returned when branch permissions are nil.
var ErrNilBranchPermissions = fmt.Errorf("nil branch permissions")

// SetEnhancedBranchProtection creates a branch protection rule via GraphQL for the
// provided repository ID and branch pattern using settings from perms. Some advanced
// features are logged and handled via REST fallbacks elsewhere.

// SetEnhancedBranchProtection creates a branch protection rule via GraphQL.
func (c *GitHubClient) SetEnhancedBranchProtection(id githubv4.ID, branchPattern string, perms *BranchPermissions) error {
	if perms == nil {
		return ErrNilBranchPermissions
	}
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

	input := buildBranchProtectionRuleInput(id, branchPattern, perms)
	logUnsupportedGraphQLFeatures(perms)

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

// buildBranchProtectionRuleInput assembles the GraphQL input from the provided permissions.
func buildBranchProtectionRuleInput(
	id githubv4.ID,
	branchPattern string,
	perms *BranchPermissions,
) githubv4.CreateBranchProtectionRuleInput {
	input := githubv4.CreateBranchProtectionRuleInput{
		RepositoryID: id,
		Pattern:      *githubv4.NewString(githubv4.String(branchPattern)),
	}
	if perms == nil {
		return input
	}
	if perms.RequirePullRequestReviews != nil {
		input.RequiresApprovingReviews = githubv4.NewBoolean(githubv4.Boolean(*perms.RequirePullRequestReviews))
	}
	if perms.ApproverCount != nil {
		input.RequiredApprovingReviewCount = githubv4.NewInt(githubv4.Int(clampToInt32(*perms.ApproverCount)))
	}
	if perms.RequireCodeOwners != nil {
		input.RequiresCodeOwnerReviews = githubv4.NewBoolean(githubv4.Boolean(*perms.RequireCodeOwners))
	}
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
	return input
}

// clampToInt32 ensures value fits in int32 range, logging if clamped.
func clampToInt32(v int) int32 {
	const maxInt32 = int32(2147483647)
	const minInt32 = int32(-2147483648)
	if v > int(maxInt32) {
		log.Warn().Int("approverCount", v).Msg("approverCount exceeds int32 range, using max value")
		return maxInt32
	}
	if v < int(minInt32) {
		log.Warn().Int("approverCount", v).Msg("approverCount below int32 range, using min value")
		return minInt32
	}
	return int32(v)
}

// logUnsupportedGraphQLFeatures logs features that require REST fallback.
func logUnsupportedGraphQLFeatures(perms *BranchPermissions) {
	if perms == nil {
		return
	}
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
}

// GetRepoQuery defines the GraphQL query structure for repository information.
type GetRepoQuery struct {
	Repository struct {
		ID                 githubv4.ID
		HasWikiEnabled     githubv4.Boolean
		HasIssuesEnabled   githubv4.Boolean
		HasProjectsEnabled githubv4.Boolean
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// GetRepository looks up a repository by owner/name and returns its GraphQL ID
// along with basic repository feature flags. Returns a RepositoryNotFoundError
// when the repository cannot be retrieved.
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

// GetRateLimitQuery defines the GraphQL query structure for rate limit information.
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

// GetRateLimit queries the current viewer and GraphQL API ratelimit status and
// logs the results for visibility.
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
