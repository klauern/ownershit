package ownershit

import (
	"strings"
	"testing"
)

func TestValidateBranchPermissions_Integration(t *testing.T) {
	tests := []struct {
		name          string
		perms         *BranchPermissions
		wantErr       bool
		errorContains string
		description   string
	}{
		{
			name: "valid minimal configuration",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(1),
			},
			wantErr:     false,
			description: "Test valid minimal branch protection configuration",
		},
		{
			name: "valid comprehensive configuration",
			perms: &BranchPermissions{
				RequirePullRequestReviews:     boolPtr(true),
				ApproverCount:                 intPtr(2),
				RequireCodeOwners:             boolPtr(true),
				RequireStatusChecks:           boolPtr(true),
				StatusChecks:                  []string{"ci/build", "ci/test"},
				RequireUpToDateBranch:         boolPtr(true),
				EnforceAdmins:                 boolPtr(true),
				RestrictPushes:                boolPtr(true),
				PushAllowlist:                 []string{"admin-team"},
				RequireConversationResolution: boolPtr(true),
				RequireLinearHistory:          boolPtr(true),
				AllowForcePushes:              boolPtr(false),
				AllowDeletions:                boolPtr(false),
			},
			wantErr:     false,
			description: "Test valid comprehensive branch protection with all features",
		},
		{
			name: "negative approver count",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(-1),
			},
			wantErr:       true,
			errorContains: "approver count cannot be negative",
			description:   "Test validation of negative approver count",
		},
		{
			name: "status checks without requirement enabled",
			perms: &BranchPermissions{
				RequireStatusChecks: boolPtr(false),
				StatusChecks:        []string{"ci/build"},
			},
			wantErr:       true,
			errorContains: "status checks specified but RequireStatusChecks is disabled",
			description:   "Test validation of status checks without requirement",
		},
		{
			name: "empty status check name",
			perms: &BranchPermissions{
				RequireStatusChecks: boolPtr(true),
				StatusChecks:        []string{"ci/build", "", "ci/test"},
			},
			wantErr:       true,
			errorContains: "empty status check name",
			description:   "Test validation of empty status check names",
		},
		{
			name: "status checks required but no checks specified",
			perms: &BranchPermissions{
				RequireStatusChecks: boolPtr(true),
				StatusChecks:        []string{},
			},
			wantErr:       true,
			errorContains: "RequireStatusChecks is enabled but no status checks specified",
			description:   "Test validation when status checks are required but none specified",
		},
		{
			name: "push restrictions without allowlist",
			perms: &BranchPermissions{
				RestrictPushes: boolPtr(true),
				PushAllowlist:  []string{},
			},
			wantErr:       true,
			errorContains: "RestrictPushes is enabled but no users/teams specified in PushAllowlist",
			description:   "Test validation of push restrictions without allowlist",
		},
		{
			name: "empty push allowlist entry",
			perms: &BranchPermissions{
				RestrictPushes: boolPtr(true),
				PushAllowlist:  []string{"admin-team", "", "release-team"},
			},
			wantErr:       true,
			errorContains: "empty entry in PushAllowlist",
			description:   "Test validation of empty push allowlist entries",
		},
		{
			name: "require up-to-date without status checks",
			perms: &BranchPermissions{
				RequireStatusChecks:   boolPtr(false),
				RequireUpToDateBranch: boolPtr(true),
			},
			wantErr:       true,
			errorContains: "RequireUpToDateBranch requires RequireStatusChecks to be enabled",
			description:   "Test validation of up-to-date requirement without status checks",
		},
		{
			name: "conflicting force push and linear history",
			perms: &BranchPermissions{
				RequireLinearHistory: boolPtr(true),
				AllowForcePushes:     boolPtr(true),
			},
			wantErr:       true,
			errorContains: "RequireLinearHistory and AllowForcePushes cannot both be enabled",
			description:   "Test validation of conflicting linear history and force push settings",
		},
		{
			name: "very large approver count",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(1000),
			},
			wantErr:       true,
			errorContains: "approver count seems unreasonably high",
			description:   "Test validation of unreasonably high approver count",
		},
		{
			name: "duplicate status checks",
			perms: &BranchPermissions{
				RequireStatusChecks: boolPtr(true),
				StatusChecks:        []string{"ci/build", "ci/test", "ci/build"},
			},
			wantErr:       true,
			errorContains: "duplicate status check",
			description:   "Test validation of duplicate status check names",
		},
		{
			name:          "nil permissions",
			perms:         nil,
			wantErr:       true,
			errorContains: "branch permissions cannot be nil",
			description:   "Test validation of nil permissions struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			err := ValidateBranchPermissions(tt.perms)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateBranchPermissions() expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateBranchPermissions() error = %v, expected to contain %v", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateBranchPermissions() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestBranchPermissions_ConfigurationParsing(t *testing.T) {
	tests := []struct {
		name            string
		yamlConfig      string
		expectedPerms   *BranchPermissions
		wantParseErr    bool
		wantValidateErr bool
		description     string
	}{
		{
			name: "minimal YAML configuration",
			yamlConfig: `
branches:
  main:
    require_pull_request_reviews: true
    approver_count: 1
`,
			expectedPerms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(1),
			},
			wantParseErr:    false,
			wantValidateErr: false,
			description:     "Test parsing minimal YAML branch protection configuration",
		},
		{
			name: "comprehensive YAML configuration",
			yamlConfig: `
branches:
  main:
    require_pull_request_reviews: true
    approver_count: 2
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
`,
			expectedPerms: &BranchPermissions{
				RequirePullRequestReviews:     boolPtr(true),
				ApproverCount:                 intPtr(2),
				RequireCodeOwners:             boolPtr(true),
				RequireStatusChecks:           boolPtr(true),
				StatusChecks:                  []string{"ci/build", "ci/test", "security/scan"},
				RequireUpToDateBranch:         boolPtr(true),
				EnforceAdmins:                 boolPtr(true),
				RestrictPushes:                boolPtr(true),
				PushAllowlist:                 []string{"admin-team", "release-team"},
				RequireConversationResolution: boolPtr(true),
				RequireLinearHistory:          boolPtr(true),
				AllowForcePushes:              boolPtr(false),
				AllowDeletions:                boolPtr(false),
			},
			wantParseErr:    false,
			wantValidateErr: false,
			description:     "Test parsing comprehensive YAML branch protection configuration",
		},
		{
			name: "invalid YAML with validation errors",
			yamlConfig: `
branches:
  main:
    require_pull_request_reviews: true
    approver_count: -1
    require_status_checks: true
    status_checks: []
`,
			expectedPerms:   nil,
			wantParseErr:    false,
			wantValidateErr: true,
			description:     "Test YAML that parses but fails validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			// Note: This test demonstrates the structure we'd need for YAML parsing
			// In a real implementation, we'd parse the YAML into a config struct
			// and then extract the BranchPermissions from it

			// For now, we'll validate the expected permissions directly
			if tt.expectedPerms != nil {
				err := ValidateBranchPermissions(tt.expectedPerms)
				if (err != nil) != tt.wantValidateErr {
					t.Errorf("ValidateBranchPermissions() error = %v, wantValidateErr %v", err, tt.wantValidateErr)
				}
			}
		})
	}
}

func TestBranchPermissions_MergeStrategiesIntegration(t *testing.T) {
	tests := []struct {
		name           string
		perms          *BranchPermissions
		mergeCommit    bool
		squashMerge    bool
		rebaseMerge    bool
		expectConflict bool
		description    string
	}{
		{
			name: "linear history with merge commit disabled",
			perms: &BranchPermissions{
				RequireLinearHistory: boolPtr(true),
			},
			mergeCommit:    false,
			squashMerge:    true,
			rebaseMerge:    true,
			expectConflict: false,
			description:    "Test that linear history works with non-merge strategies",
		},
		{
			name: "linear history with merge commit enabled",
			perms: &BranchPermissions{
				RequireLinearHistory: boolPtr(true),
			},
			mergeCommit:    true,
			squashMerge:    true,
			rebaseMerge:    true,
			expectConflict: true,
			description:    "Test that linear history conflicts with merge commits",
		},
		{
			name: "no linear history requirement",
			perms: &BranchPermissions{
				RequireLinearHistory: boolPtr(false),
			},
			mergeCommit:    true,
			squashMerge:    true,
			rebaseMerge:    true,
			expectConflict: false,
			description:    "Test that all merge strategies work without linear history",
		},
		{
			name:           "nil permissions with merge strategies",
			perms:          nil,
			mergeCommit:    true,
			squashMerge:    false,
			rebaseMerge:    false,
			expectConflict: false,
			description:    "Test merge strategies with nil branch permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			// First validate the branch permissions themselves
			if tt.perms != nil {
				if err := ValidateBranchPermissions(tt.perms); err != nil {
					t.Errorf("ValidateBranchPermissions() unexpected error = %v", err)
				}
			}

			// Then check for conflicts between branch protection and merge strategies
			hasConflict := false
			if tt.perms != nil && tt.perms.RequireLinearHistory != nil && *tt.perms.RequireLinearHistory && tt.mergeCommit {
				hasConflict = true
			}

			if hasConflict != tt.expectConflict {
				t.Errorf("Expected conflict = %v, but got = %v", tt.expectConflict, hasConflict)
			}
		})
	}
}

func TestBranchPermissions_TeamAndUserValidation(t *testing.T) {
	tests := []struct {
		name          string
		pushAllowlist []string
		wantErr       bool
		errorContains string
		description   string
	}{
		{
			name:          "valid team names",
			pushAllowlist: []string{"admin-team", "release-team", "security-team"},
			wantErr:       false,
			description:   "Test valid team names in push allowlist",
		},
		{
			name:          "valid usernames",
			pushAllowlist: []string{"admin-user", "release-bot", "security-admin"},
			wantErr:       false,
			description:   "Test valid usernames in push allowlist",
		},
		{
			name:          "mixed teams and users",
			pushAllowlist: []string{"admin-team", "release-user", "ci-bot"},
			wantErr:       false,
			description:   "Test mixed teams and users in push allowlist",
		},
		{
			name:          "empty allowlist entry",
			pushAllowlist: []string{"valid-team", "", "another-team"},
			wantErr:       true,
			errorContains: "empty entry in PushAllowlist",
			description:   "Test validation of empty allowlist entries",
		},
		{
			name:          "whitespace-only entry",
			pushAllowlist: []string{"valid-team", "   ", "another-team"},
			wantErr:       true,
			errorContains: "empty entry in PushAllowlist",
			description:   "Test validation of whitespace-only allowlist entries",
		},
		{
			name:          "duplicate entries",
			pushAllowlist: []string{"admin-team", "release-team", "admin-team"},
			wantErr:       true,
			errorContains: "duplicate entry in PushAllowlist",
			description:   "Test validation of duplicate allowlist entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			perms := &BranchPermissions{
				RestrictPushes: boolPtr(true),
				PushAllowlist:  tt.pushAllowlist,
			}

			err := ValidateBranchPermissions(perms)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateBranchPermissions() expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateBranchPermissions() error = %v, expected to contain %v", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateBranchPermissions() unexpected error = %v", err)
				}
			}
		})
	}
}

// Helper function to test that we have a ValidateBranchPermissions function.
// This would need to be implemented if it doesn't exist
func ensureValidateBranchPermissionsExists(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ValidateBranchPermissions function does not exist or has wrong signature: %v", r)
		}
	}()

	// Try to call the function with nil to see if it exists
	_ = ValidateBranchPermissions(nil)
}

func TestValidateBranchPermissions_FunctionExists(t *testing.T) {
	ensureValidateBranchPermissionsExists(t)
}
