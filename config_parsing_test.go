package ownershit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigParsing_RealYAML_BranchProtection(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		validate    func(*PermissionsSettings) error
		description string
	}{
		{
			name: "complete Phase 1 branch protection configuration",
			yamlContent: `
organization: test-org
branches:
  require_pull_request_reviews: true
  require_approving_count: 2
  require_code_owners: true
  require_status_checks: true
  status_checks:
    - "ci/build"
    - "ci/test"
    - "security/scan"
  require_up_to_date_branch: true
  enforce_admins: true
  restrict_pushes: true
  push_allowlist:
    - "admin-team"
    - "release-team"
  require_conversation_resolution: true
  require_linear_history: true
  allow_force_pushes: false
  allow_deletions: false
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false
repositories:
  - name: test-repo
    wiki: false
    issues: true
`,
			wantErr:     false,
			description: "Test complete Phase 1 branch protection fields parse correctly",
			validate: func(s *PermissionsSettings) error {
				bp := &s.BranchPermissions

				// Validate all Phase 1 fields were parsed correctly
				if bp.RequirePullRequestReviews == nil || !*bp.RequirePullRequestReviews {
					return NewConfigValidationError("require_pull_request_reviews", true, "not parsed correctly", nil)
				}
				if bp.ApproverCount == nil || *bp.ApproverCount != 2 {
					return NewConfigValidationError("require_approving_count", 2, "not parsed correctly", nil)
				}
				if bp.RequireCodeOwners == nil || !*bp.RequireCodeOwners {
					return NewConfigValidationError("require_code_owners", true, "not parsed correctly", nil)
				}
				if bp.RequireStatusChecks == nil || !*bp.RequireStatusChecks {
					return NewConfigValidationError("require_status_checks", true, "not parsed correctly", nil)
				}
				if len(bp.StatusChecks) != 3 {
					return NewConfigValidationError("status_checks", 3, "expected 3 status checks", nil)
				}
				if bp.RequireUpToDateBranch == nil || !*bp.RequireUpToDateBranch {
					return NewConfigValidationError("require_up_to_date_branch", true, "not parsed correctly", nil)
				}
				if bp.EnforceAdmins == nil || !*bp.EnforceAdmins {
					return NewConfigValidationError("enforce_admins", true, "not parsed correctly", nil)
				}
				if bp.RestrictPushes == nil || !*bp.RestrictPushes {
					return NewConfigValidationError("restrict_pushes", true, "not parsed correctly", nil)
				}
				if len(bp.PushAllowlist) != 2 {
					return NewConfigValidationError("push_allowlist", 2, "expected 2 allowlist entries", nil)
				}
				if bp.RequireConversationResolution == nil || !*bp.RequireConversationResolution {
					return NewConfigValidationError("require_conversation_resolution", true, "not parsed correctly", nil)
				}
				if bp.RequireLinearHistory == nil || !*bp.RequireLinearHistory {
					return NewConfigValidationError("require_linear_history", true, "not parsed correctly", nil)
				}
				if bp.AllowForcePushes == nil || *bp.AllowForcePushes {
					return NewConfigValidationError("allow_force_pushes", false, "not parsed correctly", nil)
				}
				if bp.AllowDeletions == nil || *bp.AllowDeletions {
					return NewConfigValidationError("allow_deletions", false, "not parsed correctly", nil)
				}

				return nil
			},
		},
		{
			name: "minimal valid configuration",
			yamlContent: `
organization: minimal-org
branches:
  require_pull_request_reviews: true
repositories:
  - name: minimal-repo
`,
			wantErr:     false,
			description: "Test minimal configuration parses without error",
			validate: func(s *PermissionsSettings) error {
				if s.RequirePullRequestReviews == nil {
					return NewConfigValidationError("require_pull_request_reviews", true, "required field missing", nil)
				}
				return nil
			},
		},
		{
			name: "empty status checks list",
			yamlContent: `
organization: test-org
branches:
  require_status_checks: true
  status_checks: []
`,
			wantErr:     false,
			description: "Test empty status checks list parses (validation should catch this)",
			validate: func(s *PermissionsSettings) error {
				if s.RequireStatusChecks == nil || !*s.RequireStatusChecks {
					return NewConfigValidationError("require_status_checks", true, "not parsed correctly", nil)
				}
				if len(s.StatusChecks) != 0 {
					return NewConfigValidationError("status_checks", 0, "should be empty", nil)
				}
				return nil
			},
		},
		{
			name: "mixed merge strategies",
			yamlContent: `
organization: merge-org
branches:
  allow_merge_commit: true
  allow_squash_merge: false
  allow_rebase_merge: true
  require_linear_history: false
`,
			wantErr:     false,
			description: "Test mixed merge strategy configuration",
			validate: func(s *PermissionsSettings) error {
				bp := &s.BranchPermissions
				if bp.AllowMergeCommit == nil || !*bp.AllowMergeCommit {
					return NewConfigValidationError("allow_merge_commit", true, "not parsed correctly", nil)
				}
				if bp.AllowSquashMerge == nil || *bp.AllowSquashMerge {
					return NewConfigValidationError("allow_squash_merge", false, "not parsed correctly", nil)
				}
				if bp.AllowRebaseMerge == nil || !*bp.AllowRebaseMerge {
					return NewConfigValidationError("allow_rebase_merge", true, "not parsed correctly", nil)
				}
				return nil
			},
		},
		{
			name: "null values for optional fields",
			yamlContent: `
organization: null-org
branches:
  require_pull_request_reviews: true
  require_approving_count: null
  enforce_admins: null
`,
			wantErr:     false,
			description: "Test null values in YAML for optional fields",
			validate: func(s *PermissionsSettings) error {
				bp := &s.BranchPermissions
				if bp.RequirePullRequestReviews == nil || !*bp.RequirePullRequestReviews {
					return NewConfigValidationError("require_pull_request_reviews", true, "not parsed correctly", nil)
				}
				if bp.ApproverCount != nil {
					return NewConfigValidationError("require_approving_count", nil, "should be nil", nil)
				}
				if bp.EnforceAdmins != nil {
					return NewConfigValidationError("enforce_admins", nil, "should be nil", nil)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			var settings PermissionsSettings
			err := yaml.Unmarshal([]byte(tt.yamlContent), &settings)

			if (err != nil) != tt.wantErr {
				t.Errorf("yaml.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // Don't validate if we expected an error
			}

			// Run custom validation if provided
			if tt.validate != nil {
				if validationErr := tt.validate(&settings); validationErr != nil {
					t.Errorf("Configuration validation failed: %v", validationErr)
				}
			}

			// Always run the standard validation
			if branchValidationErr := ValidateBranchPermissions(&settings.BranchPermissions); branchValidationErr != nil {
				t.Logf("Branch validation error (expected for some cases): %v", branchValidationErr)
			}
		})
	}
}

func TestConfigParsing_FieldTypes_InvalidYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errorType   string
		description string
	}{
		{
			name: "string instead of bool for require_status_checks",
			yamlContent: `
branches:
  require_status_checks: "true"
`,
			wantErr:     true,
			errorType:   "type mismatch",
			description: "Test string value for boolean field causes parse error",
		},
		{
			name: "string instead of int for approver count",
			yamlContent: `
branches:
  require_approving_count: "two"
`,
			wantErr:     true,
			errorType:   "type mismatch",
			description: "Test string value for integer field causes parse error",
		},
		{
			name: "array instead of bool",
			yamlContent: `
branches:
  enforce_admins: []
`,
			wantErr:     true,
			errorType:   "type mismatch",
			description: "Test array value for boolean field causes parse error",
		},
		{
			name: "object instead of string array",
			yamlContent: `
branches:
  status_checks:
    ci: true
    test: false
`,
			wantErr:     true,
			errorType:   "type mismatch",
			description: "Test object value for string array field causes parse error",
		},
		{
			name: "negative number for approver count",
			yamlContent: `
branches:
  require_approving_count: -1
`,
			wantErr:     false, // YAML parsing should succeed, validation should catch this
			errorType:   "validation",
			description: "Test negative number parses but validation catches it",
		},
		{
			name: "very large number for approver count",
			yamlContent: `
branches:
  require_approving_count: 999999999999999999999
`,
			wantErr:     true,
			errorType:   "overflow",
			description: "Test very large number causes parse error",
		},
		{
			name: "malformed YAML structure",
			yamlContent: `
branches:
  require_status_checks: true
    status_checks:
  - ci/build
`,
			wantErr:     true,
			errorType:   "yaml syntax",
			description: "Test malformed YAML syntax causes parse error",
		},
		{
			name:        "tabs instead of spaces in YAML",
			yamlContent: "branches:\n\trequire_status_checks: true",
			wantErr:     true,
			errorType:   "yaml syntax",
			description: "Test tabs in YAML cause parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			var settings PermissionsSettings
			err := yaml.Unmarshal([]byte(tt.yamlContent), &settings)

			if (err != nil) != tt.wantErr {
				t.Errorf("yaml.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				t.Logf("Parse error (expected): %v", err)
			} else if tt.errorType == "validation" {
				// If parsing succeeded but we expected an error at validation level
				validationErr := ValidateBranchPermissions(&settings.BranchPermissions)
				if validationErr == nil {
					t.Error("Expected validation error but got none")
				}
			}
		})
	}
}

func TestConfigParsing_RealFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		content     string
		expectValid bool
		description string
	}{
		{
			name:     "valid complete config file",
			filename: "complete-config.yaml",
			content: `
organization: test-org
branches:
  require_pull_request_reviews: true
  require_approving_count: 2
  require_code_owners: true
  require_status_checks: true
  status_checks: ["ci/build", "ci/test"]
  require_up_to_date_branch: true
  enforce_admins: true
  restrict_pushes: true
  push_allowlist: ["admin-team"]
  require_conversation_resolution: true
  require_linear_history: true
  allow_force_pushes: false
  allow_deletions: false
team:
  - name: admin-team
    level: admin
repositories:
  - name: test-repo
    wiki: false
    issues: true
`,
			expectValid: true,
			description: "Test complete valid configuration file",
		},
		{
			name:     "config with validation errors",
			filename: "invalid-config.yaml",
			content: `
organization: invalid-org
branches:
  require_pull_request_reviews: false
  require_approving_count: 2  # Invalid: requires PR reviews
  require_status_checks: true
  status_checks: []           # Invalid: empty list with requirement
`,
			expectValid: false,
			description: "Test configuration that parses but fails validation",
		},
		{
			name:        "empty config file",
			filename:    "empty-config.yaml",
			content:     "",
			expectValid: true, // Empty config parses but validation should catch issues
			description: "Test empty configuration file",
		},
		{
			name:     "config with only organization",
			filename: "minimal-config.yaml",
			content: `
organization: minimal-org
`,
			expectValid: true,
			description: "Test minimal configuration with only organization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			// Write test file
			testFile := filepath.Join(tempDir, tt.filename)
			err := os.WriteFile(testFile, []byte(tt.content), 0o600)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Read and parse the file
			fileContent, err := os.ReadFile(testFile) // #nosec G304 - test file in temp directory
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			var settings PermissionsSettings
			parseErr := yaml.Unmarshal(fileContent, &settings)

			if tt.expectValid {
				if parseErr != nil {
					t.Errorf("Expected valid parse but got error: %v", parseErr)
					return
				}

				// For valid configs, also test validation
				if validationErr := ValidateBranchPermissions(&settings.BranchPermissions); validationErr != nil {
					t.Logf("Validation error (may be expected): %v", validationErr)
				}
			} else {
				// For invalid configs, we expect either parse error OR validation error
				validationErr := ValidateBranchPermissions(&settings.BranchPermissions)
				if parseErr == nil && validationErr == nil {
					t.Error("Expected either parse error or validation error, but got neither")
				}
			}
		})
	}
}

func TestConfigParsing_ErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		expectParseErr bool
		expectValErr   bool
		checkMessage   func(error) bool
		description    string
	}{
		{
			name: "helpful error for malformed YAML",
			yamlContent: `
branches:
  require_status_checks: true
    invalid_indentation: true
`,
			expectParseErr: true,
			description:    "Test that malformed YAML provides helpful error messages",
			checkMessage: func(err error) bool {
				return strings.Contains(err.Error(), "line") || strings.Contains(err.Error(), "column")
			},
		},
		{
			name: "helpful error for type mismatch",
			yamlContent: `
branches:
  require_approving_count: "not a number"
`,
			expectParseErr: true,
			description:    "Test that type mismatches provide helpful error messages",
			checkMessage: func(err error) bool {
				return strings.Contains(strings.ToLower(err.Error()), "type") ||
					strings.Contains(strings.ToLower(err.Error()), "cannot")
			},
		},
		{
			name: "validation error with field context",
			yamlContent: `
branches:
  require_pull_request_reviews: false
  require_approving_count: 2
`,
			expectValErr: true,
			description:  "Test that validation errors include field context",
			checkMessage: func(err error) bool {
				return strings.Contains(err.Error(), "require_approving_count") ||
					strings.Contains(err.Error(), "require_pull_request_reviews")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			var settings PermissionsSettings
			parseErr := yaml.Unmarshal([]byte(tt.yamlContent), &settings)

			if tt.expectParseErr {
				if parseErr == nil {
					t.Error("Expected parse error but got none")
					return
				}
				if tt.checkMessage != nil && !tt.checkMessage(parseErr) {
					t.Errorf("Parse error message doesn't meet criteria: %v", parseErr)
				}
			}

			if tt.expectValErr && parseErr == nil {
				valErr := ValidateBranchPermissions(&settings.BranchPermissions)
				if valErr == nil {
					t.Error("Expected validation error but got none")
					return
				}
				if tt.checkMessage != nil && !tt.checkMessage(valErr) {
					t.Errorf("Validation error message doesn't meet criteria: %v", valErr)
				}
			}
		})
	}
}

func TestConfigParsing_BranchPermissionsIsolation(t *testing.T) {
	// Test that BranchPermissions can be parsed independently
	tests := []struct {
		name        string
		yamlContent string
		description string
	}{
		{
			name: "standalone branch permissions",
			yamlContent: `
require_pull_request_reviews: true
require_approving_count: 1
require_status_checks: true
status_checks: ["ci/build"]
enforce_admins: true
`,
			description: "Test parsing BranchPermissions as standalone YAML",
		},
		{
			name: "partial branch permissions",
			yamlContent: `
require_pull_request_reviews: true
allow_force_pushes: false
`,
			description: "Test parsing partial BranchPermissions configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			var branchPerms BranchPermissions
			err := yaml.Unmarshal([]byte(tt.yamlContent), &branchPerms)
			if err != nil {
				t.Errorf("Failed to parse BranchPermissions directly: %v", err)
				return
			}

			// Validate the parsed permissions
			validationErr := ValidateBranchPermissions(&branchPerms)
			if validationErr != nil {
				t.Logf("Validation error (may be expected): %v", validationErr)
			}
		})
	}
}
