package ownershit

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

type PermissionsLevel string

const (
	Admin PermissionsLevel = "admin"
	Read  PermissionsLevel = "pull"
	Write PermissionsLevel = "push"

	// MaxApproverCount defines the maximum reasonable number of required approvers.
	MaxApproverCount = 100
)

type Permissions struct {
	Team  *string `yaml:"name"`
	Level *string `yaml:"level"`
}

type BranchPermissions struct {
	RequireCodeOwners         *bool `yaml:"require_code_owners"`
	ApproverCount             *int  `yaml:"require_approving_count"`
	RequirePullRequestReviews *bool `yaml:"require_pull_request_reviews"`
	AllowMergeCommit          *bool `yaml:"allow_merge_commit"`
	AllowSquashMerge          *bool `yaml:"allow_squash_merge"`
	AllowRebaseMerge          *bool `yaml:"allow_rebase_merge"`

	// Advanced Branch Protection Features
	RequireStatusChecks           *bool    `yaml:"require_status_checks"`
	StatusChecks                  []string `yaml:"status_checks"`
	RequireUpToDateBranch         *bool    `yaml:"require_up_to_date_branch"`
	EnforceAdmins                 *bool    `yaml:"enforce_admins"`
	RestrictPushes                *bool    `yaml:"restrict_pushes"`
	PushAllowlist                 []string `yaml:"push_allowlist"`
	RequireConversationResolution *bool    `yaml:"require_conversation_resolution"`
	RequireLinearHistory          *bool    `yaml:"require_linear_history"`
	AllowForcePushes              *bool    `yaml:"allow_force_pushes"`
	AllowDeletions                *bool    `yaml:"allow_deletions"`
}

type PermissionsSettings struct {
	BranchPermissions `yaml:"branches"`
	TeamPermissions   []*Permissions `yaml:"team"`
	Repositories      []*Repository  `yaml:"repositories"`
	Organization      *string        `yaml:"organization"`
	DefaultLabels     []RepoLabel    `yaml:"default_labels"`
}

type Repository struct {
	Name          *string `yaml:"name"`
	Wiki          *bool   `yaml:"wiki"`
	Issues        *bool   `yaml:"issues"`
	Projects      *bool   `yaml:"projects"`
	DefaultBranch *string `yaml:"default_branch"`
	Private       *bool   `yaml:"private"`
	Archived      *bool   `yaml:"archived"`
	Template      *bool   `yaml:"template"`
	Description   *string `yaml:"description"`
	Homepage      *string `yaml:"homepage"`
}

type RepoLabel struct {
	Name        string
	Color       string
	Emoji       string
	Description string
	oldLabel    string
}

// GitHub token validation errors.
var (
	ErrTokenEmpty    = errors.New("GitHub token is empty")
	ErrTokenInvalid  = errors.New("GitHub token format is invalid")
	ErrTokenNotFound = errors.New("GITHUB_TOKEN environment variable not set")
)

// ValidateGitHubToken validates a GitHub token format and content.
func ValidateGitHubToken(token string) error {
	if token == "" {
		return ErrTokenEmpty
	}

	// Remove whitespace
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTokenEmpty
	}

	// GitHub tokens have specific patterns:
	// - Classic tokens: ghp_[A-Za-z0-9]{36}
	// - Fine-grained tokens: github_pat_[A-Za-z0-9_]+
	// - GitHub App tokens: ghs_[A-Za-z0-9]{36}
	// - OAuth tokens: gho_[A-Za-z0-9]{36}
	// - Refresh tokens: ghr_[A-Za-z0-9]{36}
	// - SAML tokens: ghu_[A-Za-z0-9]{36}
	validPatterns := []string{
		`^ghp_[A-Za-z0-9]{36}$`,             // Classic personal access token
		`^github_pat_[A-Za-z0-9_]{22,255}$`, // Fine-grained personal access token
		`^ghs_[A-Za-z0-9]{36}$`,             // GitHub App token
		`^gho_[A-Za-z0-9]{36}$`,             // OAuth token
		`^ghr_[A-Za-z0-9]{36}$`,             // Refresh token
		`^ghu_[A-Za-z0-9]{36}$`,             // SAML token
	}

	for _, pattern := range validPatterns {
		matched, err := regexp.MatchString(pattern, token)
		if err != nil {
			return fmt.Errorf("token validation regex error: %w", err)
		}
		if matched {
			return nil
		}
	}

	return ErrTokenInvalid
}

// GetValidatedGitHubToken gets and validates the GitHub token from environment.
func GetValidatedGitHubToken() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", ErrTokenNotFound
	}

	if err := ValidateGitHubToken(token); err != nil {
		return "", fmt.Errorf("invalid GitHub token: %w", err)
	}

	return token, nil
}

// GetRequiredTokenPermissions returns the minimum required GitHub token permissions
// for ownershit operations, categorized by token type.
func GetRequiredTokenPermissions() map[string][]string {
	return map[string][]string{
		"classic_token_scopes": {
			"repo",            // Full repository access
			"admin:org",       // Organization administration
			"read:org",        // Read organization data
			"user",            // User profile access
		},
		"fine_grained_permissions": {
			"Repository permissions:",
			"- Administration: Write",     // Manage repository settings
			"- Metadata: Read",           // Read repository metadata  
			"- Contents: Read",           // Read repository contents
			"- Pull requests: Write",     // Manage branch protection
			"- Issues: Write",            // Manage repository issues
			"",
			"Organization permissions:",
			"- Administration: Read",     // Read org settings
			"- Members: Read",           // Read organization members
			"- Team membership: Read",   // Read team memberships
		},
		"operations_requiring_permissions": {
			"Sync repositories: repo, admin:org",
			"Import repository config: repo, read:org", 
			"Manage branch protection: repo",
			"Archive repositories: repo, admin:org",
			"Manage teams: admin:org",
		},
	}
}

// ValidateBranchPermissions validates branch protection configuration.
func ValidateBranchPermissions(perms *BranchPermissions) error {
	if perms == nil {
		return NewConfigValidationError("branch_permissions", nil, "branch permissions cannot be nil", nil)
	}

	// Validate approver count
	if perms.ApproverCount != nil {
		if *perms.ApproverCount < 0 {
			return NewConfigValidationError("require_approving_count", *perms.ApproverCount,
				"approver count cannot be negative", nil)
		}
		if *perms.ApproverCount > MaxApproverCount {
			return NewConfigValidationError("require_approving_count", *perms.ApproverCount,
				"approver count seems unreasonably high", nil)
		}
	}

	// Validate status checks configuration
	if perms.RequireStatusChecks != nil && *perms.RequireStatusChecks {
		if len(perms.StatusChecks) == 0 {
			return NewConfigValidationError("status_checks", perms.StatusChecks,
				"RequireStatusChecks is enabled but no status checks specified", nil)
		}

		// Check for duplicates and validate individual status check names
		seen := make(map[string]bool)
		for i, check := range perms.StatusChecks {
			if strings.TrimSpace(check) == "" {
				return NewConfigValidationError(fmt.Sprintf("status_checks[%d]", i), check,
					"empty status check name", nil)
			}
			if seen[check] {
				return NewConfigValidationError("status_checks", check,
					"duplicate status check", nil)
			}
			seen[check] = true
		}
	} else if len(perms.StatusChecks) > 0 {
		return NewConfigValidationError("status_checks", perms.StatusChecks,
			"status checks specified but RequireStatusChecks is disabled", nil)
	}

	// Validate push allowlist configuration
	if perms.RestrictPushes != nil && *perms.RestrictPushes {
		if len(perms.PushAllowlist) == 0 {
			return NewConfigValidationError("push_allowlist", perms.PushAllowlist,
				"RestrictPushes is enabled but no users/teams specified in PushAllowlist", nil)
		}

		// Check for duplicates and validate push allowlist entries
		seen := make(map[string]bool)
		for i, actor := range perms.PushAllowlist {
			trimmed := strings.TrimSpace(actor)
			if trimmed == "" {
				return NewConfigValidationError(fmt.Sprintf("push_allowlist[%d]", i), actor,
					"empty entry in PushAllowlist", nil)
			}
			if seen[actor] {
				return NewConfigValidationError("push_allowlist", actor,
					"duplicate entry in PushAllowlist", nil)
			}
			seen[actor] = true
		}
	}

	// Validate logical consistency
	if perms.RequirePullRequestReviews != nil && !*perms.RequirePullRequestReviews {
		if perms.ApproverCount != nil && *perms.ApproverCount > 0 {
			return NewConfigValidationError("require_approving_count", *perms.ApproverCount,
				"cannot require approving reviews when require_pull_request_reviews is false", nil)
		}
		if perms.RequireCodeOwners != nil && *perms.RequireCodeOwners {
			return NewConfigValidationError("require_code_owners", *perms.RequireCodeOwners,
				"cannot require code owner reviews when require_pull_request_reviews is false", nil)
		}
	}

	// Validate up-to-date branch requirement
	if perms.RequireUpToDateBranch != nil && *perms.RequireUpToDateBranch {
		if perms.RequireStatusChecks == nil || !*perms.RequireStatusChecks {
			return NewConfigValidationError("require_up_to_date_branch", *perms.RequireUpToDateBranch,
				"RequireUpToDateBranch requires RequireStatusChecks to be enabled", nil)
		}
	}

	// Validate conflicting settings
	if perms.RequireLinearHistory != nil && *perms.RequireLinearHistory &&
		perms.AllowForcePushes != nil && *perms.AllowForcePushes {
		return NewConfigValidationError("allow_force_pushes", *perms.AllowForcePushes,
			"RequireLinearHistory and AllowForcePushes cannot both be enabled", nil)
	}

	return nil
}

// ValidatePermissionsSettings validates the overall permissions configuration.
func ValidatePermissionsSettings(settings *PermissionsSettings) error {
	if settings == nil {
		return NewConfigValidationError("settings", nil, "permissions settings cannot be nil", nil)
	}

	// Validate organization field
	if settings.Organization == nil || *settings.Organization == "" {
		return NewConfigValidationError("organization", settings.Organization,
			"organization must be specified and cannot be empty", nil)
	}

	// Validate organization name format (GitHub usernames/orgs)
	orgName := strings.TrimSpace(*settings.Organization)
	if len(orgName) < 1 || len(orgName) > 39 {
		return NewConfigValidationError("organization", orgName,
			"organization name must be between 1 and 39 characters", nil)
	}

	// Validate branch permissions
	if err := ValidateBranchPermissions(&settings.BranchPermissions); err != nil {
		return err
	}

	// Validate repositories
	if len(settings.Repositories) == 0 {
		return NewConfigValidationError("repositories", settings.Repositories,
			"at least one repository must be specified", nil)
	}

	// Validate each repository
	repoNames := make(map[string]bool)
	for i, repo := range settings.Repositories {
		if repo == nil {
			return NewConfigValidationError(fmt.Sprintf("repositories[%d]", i), nil,
				"repository cannot be nil", nil)
		}
		if repo.Name == nil || *repo.Name == "" {
			return NewConfigValidationError(fmt.Sprintf("repositories[%d].name", i), repo.Name,
				"repository name must be specified and cannot be empty", nil)
		}

		repoName := strings.TrimSpace(*repo.Name)
		if repoNames[repoName] {
			return NewConfigValidationError(fmt.Sprintf("repositories[%d].name", i), repoName,
				"duplicate repository name", nil)
		}
		repoNames[repoName] = true
	}

	return nil
}

// GitHubTokenEnv was removed for security reasons.
// Use GetValidatedGitHubToken() for secure token handling instead.

func MapPermissions(settings *PermissionsSettings, client *GitHubClient) {
	// Validate complete configuration
	if err := ValidatePermissionsSettings(settings); err != nil {
		log.Err(err).Msg("configuration validation failed")
		return
	}

	for _, repo := range settings.Repositories {
		if len(settings.TeamPermissions) > 0 {
			for _, perm := range settings.TeamPermissions {
				log.Info().
					Str("repository", *repo.Name).
					Msg("Adding Permissions to repository")
				log.Debug().
					Str("repository", *repo.Name).
					Str("permissions-level", *perm.Level).
					Str("permissions-team", *perm.Team).
					Msg("permissions to add to repository")
				err := client.AddPermissions(*settings.Organization, *repo.Name, perm)
				if err != nil {
					log.Err(err).
						Str("repository", *repo.Name).
						Str("permissions-level", *perm.Level).
						Str("permissions-team", *perm.Team).
						Str("operation", "addTeamPermissions").
						Msg("setting team permissions")
					// Continue processing other permissions even if one fails
				}
			}
		}
		if err := client.UpdateBranchPermissions(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
			log.Err(err).
				Str("repository", *repo.Name).
				Str("organization", *settings.Organization).
				Msg("updating repository settings")
		}

		// Get repository ID once for both operations
		repoID, err := client.GetRepository(repo.Name, settings.Organization)
		if err != nil {
			log.Err(err).
				Str("repository", *repo.Name).
				Msg("getting repository")
		} else {
			log.Debug().
				Interface("repoID", repoID).
				Msg("Repository ID")

			// Apply enhanced branch protection rules via GraphQL
			if err := client.SetEnhancedBranchProtection(repoID, "main", &settings.BranchPermissions); err != nil {
				log.Err(err).
					Str("repository", *repo.Name).
					Str("organization", *settings.Organization).
					Msg("setting enhanced branch protection")
			}

			// Apply advanced features via REST API fallback
			if err := client.SetBranchProtectionFallback(*settings.Organization, *repo.Name, "main", &settings.BranchPermissions); err != nil {
				log.Err(err).
					Str("repository", *repo.Name).
					Str("organization", *settings.Organization).
					Msg("setting branch protection fallback via REST API")
			}

			// Set repository features (wiki, issues, projects)
			err := client.SetRepository(&repoID, repo.Wiki, repo.Issues, repo.Projects)
			if err != nil {
				log.Err(err).
					Interface("repoID", repoID).
					Bool("wikiEnabled", *repo.Wiki).
					Bool("issuesEnabled", *repo.Issues).
					Bool("projectsEnabled", *repo.Projects).
					Msg("setting repository fields")
			}
		}
	}
}

func UpdateBranchMergeStrategies(settings *PermissionsSettings, client *GitHubClient) {
	// Validate branch permissions configuration
	if err := ValidateBranchPermissions(&settings.BranchPermissions); err != nil {
		log.Err(err).Msg("branch permissions validation failed")
		return
	}

	for _, repo := range settings.Repositories {
		log.Info().
			Str("repository", *repo.Name).
			Bool("squash-commits", *settings.AllowSquashMerge).
			Bool("merges", *settings.AllowMergeCommit).
			Bool("rebase-merge", *settings.AllowRebaseMerge).
			Msg("Updating settings")
		if err := client.UpdateBranchPermissions(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
			log.Err(err).Str("repository", *repo.Name).Str("organization", *settings.Organization).Msg("updating repository settings")
		}
	}
}

func SyncLabels(settings *PermissionsSettings, client *GitHubClient) {
	for _, repo := range settings.Repositories {
		log.Info().
			Str("repository", *repo.Name).
			Msg("Updating Labels")
		if err := client.SyncLabels(*settings.Organization, *repo.Name, settings.DefaultLabels); err != nil {
			log.Err(err).Msg("synchronizing Labels")
		}
	}
}
