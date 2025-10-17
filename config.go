package ownershit

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
)

// PermissionsLevel represents the access level for GitHub repository permissions.
type PermissionsLevel string

const (
	// Admin permission level provides full administrative access.
	Admin PermissionsLevel = "admin"
	// Read permission level provides pull access.
	Read PermissionsLevel = "pull"
	// Write permission level provides push access.
	Write PermissionsLevel = "push"

	// MaxApproverCount defines the maximum reasonable number of required approvers.
	MaxApproverCount = 100

	// CurrentSchemaVersion is the current configuration schema version.
	CurrentSchemaVersion = "1.0"

	// LegacySchemaVersion for migration support.
	LegacySchemaVersion = "0.9"
)

// Permissions defines team access permissions for a repository.
type Permissions struct {
	Team  *string `yaml:"name"`
	Level *string `yaml:"level"`
}

// BranchPermissions defines protection rules for repository branches.
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

// PermissionsSettings contains the complete configuration for repository permissions.
type PermissionsSettings struct {
	Version           *string `yaml:"version,omitempty"`
	BranchPermissions `yaml:"branches"`
	TeamPermissions   []*Permissions `yaml:"team"`
	Repositories      []*Repository  `yaml:"repositories"`
	Organization      *string        `yaml:"organization"`
	DefaultLabels     []RepoLabel    `yaml:"default_labels"`
}

// OperationOptions controls how mutating operations are executed.
type OperationOptions struct {
	DryRun bool
}

func (opts OperationOptions) logDryRun(action, repository string, extra map[string]interface{}) {
	if !opts.DryRun {
		return
	}
	event := log.Info().
		Str("mode", "dry-run").
		Str("action", action)
	if repository != "" {
		event = event.Str("repository", repository)
	}
	for key, value := range extra {
		event = event.Interface(key, value)
	}
	event.Msg("dry run: no changes applied")
}

// Repository defines the configuration for a single GitHub repository.
type Repository struct {
	Name                   *string `yaml:"name"`
	Wiki                   *bool   `yaml:"wiki"`
	Issues                 *bool   `yaml:"issues"`
	Projects               *bool   `yaml:"projects"`
	DefaultBranch          *string `yaml:"default_branch"`
	Private                *bool   `yaml:"private"`
	Archived               *bool   `yaml:"archived"`
	Template               *bool   `yaml:"template"`
	Description            *string `yaml:"description"`
	Homepage               *string `yaml:"homepage"`
	DeleteBranchOnMerge    *bool   `yaml:"delete_branch_on_merge"`
	HasDiscussionsEnabled  *bool   `yaml:"discussions_enabled"`
	HasSponsorshipsEnabled *bool   `yaml:"sponsorships_enabled,omitempty"`
}

// RepoLabel defines a label that can be applied to GitHub repositories.
type RepoLabel struct {
	Name        string
	Color       string
	Emoji       string
	Description string
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
			"repo",      // Full repository access
			"admin:org", // Organization administration
			"read:org",  // Read organization data
			"user",      // User profile access
		},
		"fine_grained_permissions": {
			"Repository permissions:",
			"- Administration: Write", // Manage repository settings
			"- Metadata: Read",        // Read repository metadata
			"- Contents: Read",        // Read repository contents
			"- Pull requests: Write",  // Manage branch protection
			"- Issues: Write",         // Manage repository issues
			"",
			"Organization permissions:",
			"- Administration: Read",  // Read org settings
			"- Members: Read",         // Read organization members
			"- Team membership: Read", // Read team memberships
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
	if err := validateApproverCount(perms); err != nil {
		return err
	}
	if err := validateStatusChecks(perms); err != nil {
		return err
	}
	if err := validatePushAllowlist(perms); err != nil {
		return err
	}
	if err := validateLogicalConsistency(perms); err != nil {
		return err
	}
	if err := validateUpToDateRequirement(perms); err != nil {
		return err
	}
	if err := validateConflicts(perms); err != nil {
		return err
	}
	return nil
}

func validateApproverCount(perms *BranchPermissions) error {
	if perms.ApproverCount == nil {
		return nil
	}
	if *perms.ApproverCount < 0 {
		return NewConfigValidationError("require_approving_count", *perms.ApproverCount,
			"approver count cannot be negative", nil)
	}
	if *perms.ApproverCount > MaxApproverCount {
		return NewConfigValidationError("require_approving_count", *perms.ApproverCount,
			"approver count seems unreasonably high", nil)
	}
	return nil
}

func validateStatusChecks(perms *BranchPermissions) error {
	if perms.RequireStatusChecks != nil && *perms.RequireStatusChecks {
		if len(perms.StatusChecks) == 0 {
			return NewConfigValidationError("status_checks", perms.StatusChecks,
				"RequireStatusChecks is enabled but no status checks specified", nil)
		}
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
		return nil
	}
	if len(perms.StatusChecks) > 0 {
		return NewConfigValidationError("status_checks", perms.StatusChecks,
			"status checks specified but RequireStatusChecks is disabled", nil)
	}
	return nil
}

func validatePushAllowlist(perms *BranchPermissions) error {
	if perms.RestrictPushes == nil || !*perms.RestrictPushes {
		return nil
	}
	if len(perms.PushAllowlist) == 0 {
		return NewConfigValidationError("push_allowlist", perms.PushAllowlist,
			"RestrictPushes is enabled but no users/teams specified in PushAllowlist", nil)
	}
	seen := make(map[string]bool)
	for i, actor := range perms.PushAllowlist {
		trimmed := strings.TrimSpace(actor)
		if trimmed == "" {
			return NewConfigValidationError(fmt.Sprintf("push_allowlist[%d]", i), trimmed,
				"empty entry in PushAllowlist", nil)
		}
		if seen[trimmed] {
			return NewConfigValidationError("push_allowlist", trimmed,
				"duplicate entry in PushAllowlist", nil)
		}
		seen[trimmed] = true
	}
	return nil
}

func validateLogicalConsistency(perms *BranchPermissions) error {
	if perms.RequirePullRequestReviews == nil || *perms.RequirePullRequestReviews {
		return nil
	}
	if perms.ApproverCount != nil && *perms.ApproverCount > 0 {
		return NewConfigValidationError("require_approving_count", *perms.ApproverCount,
			"cannot require approving reviews when require_pull_request_reviews is false", nil)
	}
	if perms.RequireCodeOwners != nil && *perms.RequireCodeOwners {
		return NewConfigValidationError("require_code_owners", *perms.RequireCodeOwners,
			"cannot require code owner reviews when require_pull_request_reviews is false", nil)
	}
	return nil
}

func validateUpToDateRequirement(perms *BranchPermissions) error {
	if perms.RequireUpToDateBranch == nil || !*perms.RequireUpToDateBranch {
		return nil
	}
	if perms.RequireStatusChecks == nil || !*perms.RequireStatusChecks {
		return NewConfigValidationError("require_up_to_date_branch", *perms.RequireUpToDateBranch,
			"RequireUpToDateBranch requires RequireStatusChecks to be enabled", nil)
	}
	return nil
}

func validateConflicts(perms *BranchPermissions) error {
	if perms.RequireLinearHistory != nil && *perms.RequireLinearHistory &&
		perms.AllowForcePushes != nil && *perms.AllowForcePushes {
		return NewConfigValidationError("allow_force_pushes", *perms.AllowForcePushes,
			"RequireLinearHistory and AllowForcePushes cannot both be enabled", nil)
	}
	return nil
}

// ValidateTeamPermissions validates team configuration.
func ValidateTeamPermissions(teams []*Permissions) error {
	if len(teams) == 0 {
		// Empty team list is allowed
		return nil
	}

	teamNames := make(map[string]bool)
	validLevels := map[string]bool{
		"admin":    true,
		"push":     true,
		"pull":     true,
		"triage":   true,
		"maintain": true,
	}

	for i, team := range teams {
		if team == nil {
			return NewConfigValidationError(fmt.Sprintf("team[%d]", i), nil,
				"team entry cannot be nil", nil)
		}

		// Validate team name
		if team.Team == nil || *team.Team == "" {
			return NewConfigValidationError(fmt.Sprintf("team[%d].name", i), team.Team,
				"team name must be specified and cannot be empty", nil)
		}

		teamName := strings.TrimSpace(*team.Team)
		if teamName == "" {
			return NewConfigValidationError(fmt.Sprintf("team[%d].name", i), teamName,
				"team name cannot be empty or whitespace only", nil)
		}

		// Check for duplicate team names
		if teamNames[teamName] {
			return NewConfigValidationError(fmt.Sprintf("team[%d].name", i), teamName,
				"duplicate team name", nil)
		}
		teamNames[teamName] = true

		// Validate permission level
		if team.Level == nil || *team.Level == "" {
			return NewConfigValidationError(fmt.Sprintf("team[%d].level", i), team.Level,
				"team permission level must be specified and cannot be empty", nil)
		}

		level := strings.TrimSpace(strings.ToLower(*team.Level))
		if !validLevels[level] {
			return NewConfigValidationError(fmt.Sprintf("team[%d].level", i), level,
				"invalid permission level, must be one of: admin, push, pull, triage, maintain", nil)
		}
	}

	return nil
}

// ValidateDefaultLabels validates default label configuration.
func ValidateDefaultLabels(labels []RepoLabel) error {
	if len(labels) == 0 {
		// Empty label list is allowed
		return nil
	}

	labelNames := make(map[string]bool)
	hexColorPattern := regexp.MustCompile(`^[0-9a-fA-F]{6}$`)

	for i, label := range labels {
		// Validate label name
		name := strings.TrimSpace(label.Name)
		if name == "" {
			return NewConfigValidationError(fmt.Sprintf("default_labels[%d].name", i), label.Name,
				"label name cannot be empty", nil)
		}

		// Check for duplicate label names
		if labelNames[name] {
			return NewConfigValidationError(fmt.Sprintf("default_labels[%d].name", i), name,
				"duplicate label name", nil)
		}
		labelNames[name] = true

		// Validate color format (6-character hex code without #)
		color := strings.TrimSpace(label.Color)
		if color == "" {
			return NewConfigValidationError(fmt.Sprintf("default_labels[%d].color", i), label.Color,
				"label color cannot be empty", nil)
		}

		// Remove # prefix if present
		color = strings.TrimPrefix(color, "#")
		if !hexColorPattern.MatchString(color) {
			return NewConfigValidationError(fmt.Sprintf("default_labels[%d].color", i), label.Color,
				"label color must be a 6-character hexadecimal color code (e.g., 'd73a4a' or '#d73a4a')", nil)
		}

		// Description is optional but validate length if present
		if len(label.Description) > 100 {
			return NewConfigValidationError(fmt.Sprintf("default_labels[%d].description", i), label.Description,
				"label description cannot exceed 100 characters", nil)
		}

		// Emoji is optional - no validation needed as any string is acceptable
	}

	return nil
}

// ValidatePermissionsSettings validates the overall permissions configuration.
func ValidatePermissionsSettings(settings *PermissionsSettings) error {
	if settings == nil {
		return NewConfigValidationError("settings", nil, "permissions settings cannot be nil", nil)
	}

	// Validate schema version
	if err := ValidateSchemaVersion(settings.Version); err != nil {
		return err
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

	// Validate team permissions
	if err := ValidateTeamPermissions(settings.TeamPermissions); err != nil {
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

	// Validate default labels
	if err := ValidateDefaultLabels(settings.DefaultLabels); err != nil {
		return err
	}

	return nil
}

// ValidateSchemaVersion validates the configuration schema version.
func ValidateSchemaVersion(version *string) error {
	if version == nil {
		// Assume legacy version if not specified
		return nil
	}

	versionStr := strings.TrimSpace(*version)
	switch versionStr {
	case CurrentSchemaVersion, LegacySchemaVersion:
		return nil
	case "":
		// Empty version is allowed (defaults to current)
		return nil
	default:
		return NewConfigValidationError("version", versionStr,
			fmt.Sprintf("unsupported schema version, supported versions: %s, %s",
				CurrentSchemaVersion, LegacySchemaVersion), nil)
	}
}

// MigrateConfigurationSchema migrates configuration from older schema versions to current.
func MigrateConfigurationSchema(settings *PermissionsSettings) error {
	if settings == nil {
		return NewConfigValidationError("settings", nil, "settings cannot be nil for migration", nil)
	}

	// Determine the current version
	currentVersion := LegacySchemaVersion // Default for configs without version
	if settings.Version != nil && *settings.Version != "" {
		currentVersion = strings.TrimSpace(*settings.Version)
	}

	// Validate the version is supported
	if err := ValidateSchemaVersion(&currentVersion); err != nil {
		return fmt.Errorf("cannot migrate unsupported schema version: %w", err)
	}

	// No migration needed if already current
	if currentVersion == CurrentSchemaVersion {
		return nil
	}

	// Migrate from legacy version to current
	if currentVersion == LegacySchemaVersion {
		log.Info().
			Str("fromVersion", currentVersion).
			Str("toVersion", CurrentSchemaVersion).
			Msg("migrating configuration schema")

		// Migration steps from 0.9 to 1.0
		// In this case, no structural changes are needed - just version bump
		newVersion := CurrentSchemaVersion
		settings.Version = &newVersion

		log.Info().
			Str("version", CurrentSchemaVersion).
			Msg("configuration schema migration completed")
	}

	return nil
}

// GetSchemaVersion returns the schema version, defaulting to legacy if not set.
func GetSchemaVersion(settings *PermissionsSettings) string {
	if settings == nil || settings.Version == nil || *settings.Version == "" {
		return LegacySchemaVersion
	}
	return strings.TrimSpace(*settings.Version)
}

// IsCurrentSchemaVersion checks if the configuration uses the current schema version.
func IsCurrentSchemaVersion(settings *PermissionsSettings) bool {
	return GetSchemaVersion(settings) == CurrentSchemaVersion
}

// CollectConfigWarnings inspects the configuration and reports non-fatal warnings
// such as usage of deprecated schema versions or features that will be removed.
// originalVersion should reflect the schema version before any migrations were run.
func CollectConfigWarnings(settings *PermissionsSettings, originalVersion string) []string {
	warnings := []string{}
	version := strings.TrimSpace(originalVersion)
	if version == "" {
		version = GetSchemaVersion(settings)
	}
	if version == LegacySchemaVersion {
		warnings = append(warnings,
			fmt.Sprintf("configuration schema version %s is deprecated; automatically migrated to %s", LegacySchemaVersion, CurrentSchemaVersion))
	}

	// Security best practice warnings
	warnings = append(warnings, collectSecurityWarnings(settings)...)

	return warnings
}

// collectSecurityWarnings checks for security best practice violations.
func collectSecurityWarnings(settings *PermissionsSettings) []string {
	warnings := []string{}

	if settings == nil {
		return warnings
	}

	bp := &settings.BranchPermissions

	// Branch protection warnings
	if bp.EnforceAdmins == nil || !*bp.EnforceAdmins {
		warnings = append(warnings,
			"Security: enforce_admins is not enabled - administrators can bypass branch protection rules")
	}

	if bp.AllowForcePushes != nil && *bp.AllowForcePushes {
		warnings = append(warnings,
			"Security: allow_force_pushes is enabled - this can rewrite history and cause data loss")
	}

	if bp.AllowDeletions != nil && *bp.AllowDeletions {
		warnings = append(warnings,
			"Security: allow_deletions is enabled - branches can be permanently deleted")
	}

	if bp.RequirePullRequestReviews == nil || !*bp.RequirePullRequestReviews {
		warnings = append(warnings,
			"Security: require_pull_request_reviews is not enabled - code can be merged without review")
	}

	if bp.RequireCodeOwners != nil && *bp.RequireCodeOwners {
		if bp.RequirePullRequestReviews == nil || !*bp.RequirePullRequestReviews {
			warnings = append(warnings,
				"Security: require_code_owners requires require_pull_request_reviews to be effective")
		}
	}

	if bp.RequireStatusChecks == nil || !*bp.RequireStatusChecks {
		warnings = append(warnings,
			"Best practice: require_status_checks is not enabled - CI/CD checks will not be enforced")
	}

	if bp.ApproverCount != nil && *bp.ApproverCount == 0 {
		if bp.RequirePullRequestReviews != nil && *bp.RequirePullRequestReviews {
			warnings = append(warnings,
				"Best practice: require_approving_count is 0 - pull requests do not require approval")
		}
	}

	if bp.ApproverCount != nil && *bp.ApproverCount == 1 {
		warnings = append(warnings,
			"Best practice: require_approving_count is 1 - consider requiring 2+ approvers for critical repositories")
	}

	// Team permission warnings
	for _, team := range settings.TeamPermissions {
		if team != nil && team.Level != nil && strings.ToLower(*team.Level) == "admin" {
			teamName := "unknown"
			if team.Team != nil {
				teamName = *team.Team
			}
			warnings = append(warnings,
				fmt.Sprintf("Security: team '%s' has admin permissions - consider using push/maintain for most teams", teamName))
		}
	}

	return warnings
}

// GitHubTokenEnv was removed for security reasons.
// Use GetValidatedGitHubToken() for secure token handling instead.

// MapPermissions applies the permissions settings to the target repositories.
// It validates the configuration, adds team permissions, updates branch settings,
// applies enhanced branch protection, and sets repository-level features.
func MapPermissions(settings *PermissionsSettings, client *GitHubClient, opts OperationOptions) error {
	// Validate complete configuration
	if err := ValidatePermissionsSettings(settings); err != nil {
		log.Err(err).Msg("configuration validation failed")
		return err
	}

	for _, repo := range settings.Repositories {
		applyTeamPermissions(settings, repo, client, opts)
		updateRepoBranchSettings(settings, repo, client, opts)
		if opts.DryRun {
			repoName := ""
			if repo != nil && repo.Name != nil {
				repoName = *repo.Name
			}
			opts.logDryRun("set_enhanced_branch_protection", repoName, nil)
			opts.logDryRun("set_branch_protection_fallback", repoName, nil)
			opts.logDryRun("set_repository_features", repoName, nil)
			extra := map[string]interface{}{}
			if repo != nil && repo.DeleteBranchOnMerge != nil {
				extra["delete_branch_on_merge"] = *repo.DeleteBranchOnMerge
			}
			opts.logDryRun("set_advanced_repository_settings", repoName, extra)
			continue
		}
		repoID, ok := getRepositoryID(settings, repo, client)
		if !ok {
			continue
		}
		applyEnhancedBranchProtection(settings, repo, repoID, client)
		// REST fallback for unsupported fields or existing rule conflicts
		applyBranchProtectionFallback(settings, repo, client)
		setRepositoryFeatures(repo, repoID, client)
		setAdvancedRepoSettings(settings, repo, client)
	}

	return nil
}

func safeString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func applyTeamPermissions(settings *PermissionsSettings, repo *Repository, client *GitHubClient, opts OperationOptions) {
	if len(settings.TeamPermissions) == 0 {
		return
	}
	if repo == nil || repo.Name == nil {
		log.Warn().Msg("repository entry missing name during team permissions update")
		return
	}
	for _, perm := range settings.TeamPermissions {
		log.Info().Str("repository", *repo.Name).Msg("Adding Permissions to repository")
		log.Debug().
			Str("repository", *repo.Name).
			Str("permissions-level", *perm.Level).
			Str("permissions-team", *perm.Team).
			Msg("permissions to add to repository")
		opts.logDryRun("add_team_permissions", *repo.Name, map[string]interface{}{
			"team":  safeString(perm.Team),
			"level": safeString(perm.Level),
		})
		if opts.DryRun {
			continue
		}
		if err := client.AddPermissions(*settings.Organization, *repo.Name, perm); err != nil {
			log.Err(err).
				Str("repository", *repo.Name).
				Str("permissions-level", *perm.Level).
				Str("permissions-team", *perm.Team).
				Str("operation", "addTeamPermissions").
				Msg("setting team permissions")
		}
	}
}

func updateRepoBranchSettings(settings *PermissionsSettings, repo *Repository, client *GitHubClient, opts OperationOptions) {
	if repo == nil || repo.Name == nil {
		log.Warn().Msg("repository entry missing name during branch settings update")
		return
	}
	opts.logDryRun("update_branch_permissions", *repo.Name, nil)
	if opts.DryRun {
		return
	}
	if err := client.UpdateBranchPermissions(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
		log.Err(err).
			Str("repository", *repo.Name).
			Str("organization", *settings.Organization).
			Msg("updating repository settings")
	}
}

func getRepositoryID(settings *PermissionsSettings, repo *Repository, client *GitHubClient) (githubv4.ID, bool) {
	repoID, err := client.GetRepository(repo.Name, settings.Organization)
	if err != nil {
		log.Err(err).Str("repository", *repo.Name).Msg("getting repository")
		return githubv4.ID(""), false
	}
	log.Debug().Interface("repoID", repoID).Msg("Repository ID")
	return repoID, true
}

func applyEnhancedBranchProtection(settings *PermissionsSettings, repo *Repository, repoID githubv4.ID, client *GitHubClient) {
	if err := client.SetEnhancedBranchProtection(repoID, "main", &settings.BranchPermissions); err != nil {
		log.Err(err).
			Str("repository", *repo.Name).
			Str("organization", *settings.Organization).
			Msg("setting enhanced branch protection")
	}
}

func applyBranchProtectionFallback(settings *PermissionsSettings, repo *Repository, client *GitHubClient) {
	if err := client.SetBranchProtectionFallback(*settings.Organization, *repo.Name, "main", &settings.BranchPermissions); err != nil {
		log.Err(err).
			Str("repository", *repo.Name).
			Str("organization", *settings.Organization).
			Msg("setting branch protection fallback via REST API")
	}
}

func setRepositoryFeatures(repo *Repository, repoID githubv4.ID, client *GitHubClient) {
	if err := client.SetRepository(
		repoID,
		repo.Wiki,
		repo.Issues,
		repo.Projects,
		repo.HasDiscussionsEnabled,
		repo.HasSponsorshipsEnabled,
	); err != nil {
		logEvent := log.Err(err).Interface("repoID", repoID)
		if repo.Wiki != nil {
			logEvent = logEvent.Bool("wikiEnabled", *repo.Wiki)
		}
		if repo.Issues != nil {
			logEvent = logEvent.Bool("issuesEnabled", *repo.Issues)
		}
		if repo.Projects != nil {
			logEvent = logEvent.Bool("projectsEnabled", *repo.Projects)
		}
		if repo.HasDiscussionsEnabled != nil {
			logEvent = logEvent.Bool("discussionsEnabled", *repo.HasDiscussionsEnabled)
		}
		if repo.HasSponsorshipsEnabled != nil {
			logEvent = logEvent.Bool("sponsorshipsEnabled", *repo.HasSponsorshipsEnabled)
		}
		logEvent.Msg("setting repository fields")
	}
}

func setAdvancedRepoSettings(settings *PermissionsSettings, repo *Repository, client *GitHubClient) {
	if repo.DeleteBranchOnMerge == nil {
		return
	}
	if err := client.SetRepositoryAdvancedSettings(*settings.Organization, *repo.Name, repo.DeleteBranchOnMerge); err != nil {
		log.Err(err).
			Str("repository", *repo.Name).
			Str("organization", *settings.Organization).
			Bool("deleteBranchOnMerge", *repo.DeleteBranchOnMerge).
			Msg("setting advanced repository settings")
	}
}

// UpdateBranchMergeStrategies updates merge strategy settings (merge, rebase, squash)
// for all repositories in the provided configuration using the REST API.
func UpdateBranchMergeStrategies(settings *PermissionsSettings, client *GitHubClient, opts OperationOptions) error {
	// Validate branch permissions configuration
	if err := ValidateBranchPermissions(&settings.BranchPermissions); err != nil {
		log.Err(err).Msg("branch permissions validation failed")
		return err
	}

	var multiErr error
	for _, repo := range settings.Repositories {
		if repo == nil || repo.Name == nil {
			log.Warn().Msg("repository entry missing name during branch merge strategy update")
			continue
		}
		b := func(p *bool) bool {
			if p == nil {
				return false
			}
			return *p
		}
		log.Info().
			Str("repository", *repo.Name).
			Bool("squash-commits", b(settings.AllowSquashMerge)).
			Bool("merges", b(settings.AllowMergeCommit)).
			Bool("rebase-merge", b(settings.AllowRebaseMerge)).
			Msg("Updating settings")
		opts.logDryRun("update_branch_merge_strategies", *repo.Name, map[string]interface{}{
			"allow_squash_merge": b(settings.AllowSquashMerge),
			"allow_merge_commit": b(settings.AllowMergeCommit),
			"allow_rebase_merge": b(settings.AllowRebaseMerge),
		})
		if opts.DryRun {
			continue
		}
		if err := client.UpdateBranchPermissions(*settings.Organization, *repo.Name, &settings.BranchPermissions); err != nil {
			log.Err(err).
				Str("repository", *repo.Name).
				Str("organization", *settings.Organization).
				Msg("updating repository settings")
			multiErr = errors.Join(multiErr, err)
		}
	}
	return multiErr
}

// SyncLabels synchronizes labels for each repository in the configuration,
// creating, updating, or deleting labels to match the desired state.
func SyncLabels(settings *PermissionsSettings, client *GitHubClient, opts OperationOptions) error {
	var multiErr error
	for _, repo := range settings.Repositories {
		if repo == nil || repo.Name == nil {
			log.Warn().Msg("repository entry missing name during label sync")
			continue
		}
		log.Info().
			Str("repository", *repo.Name).
			Msg("Updating Labels")
		opts.logDryRun("sync_labels", *repo.Name, map[string]interface{}{
			"label_count": len(settings.DefaultLabels),
		})
		if opts.DryRun {
			continue
		}
		if err := client.SyncLabels(*settings.Organization, *repo.Name, settings.DefaultLabels); err != nil {
			log.Err(err).Msg("synchronizing Labels")
			multiErr = errors.Join(multiErr, err)
		}
	}
	return multiErr
}
