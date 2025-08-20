package ownershit

import (
	"testing"

	"github.com/google/go-github/v66/github"
)

func TestConvertBranchProtection(t *testing.T) {
	tests := []struct {
		name       string
		protection *github.Protection
		expected   *BranchPermissions
	}{
		{
			name:       "nil protection should return empty permissions",
			protection: nil,
			expected:   &BranchPermissions{},
		},
		{
			name: "advanced settings enabled",
			protection: &github.Protection{
				RequiredConversationResolution: &github.RequiredConversationResolution{Enabled: true},
				RequireLinearHistory:           &github.RequireLinearHistory{Enabled: true},
				AllowForcePushes:               &github.AllowForcePushes{Enabled: true},
				AllowDeletions:                 &github.AllowDeletions{Enabled: true},
			},
			expected: &BranchPermissions{
				RequireConversationResolution: github.Bool(true),
				RequireLinearHistory:          github.Bool(true),
				AllowForcePushes:              github.Bool(true),
				AllowDeletions:                github.Bool(true),
			},
		},
		{
			name: "advanced settings disabled",
			protection: &github.Protection{
				RequiredConversationResolution: &github.RequiredConversationResolution{Enabled: false},
				RequireLinearHistory:           &github.RequireLinearHistory{Enabled: false},
				AllowForcePushes:               &github.AllowForcePushes{Enabled: false},
				AllowDeletions:                 &github.AllowDeletions{Enabled: false},
			},
			expected: &BranchPermissions{
				RequireConversationResolution: github.Bool(false),
				RequireLinearHistory:          github.Bool(false),
				AllowForcePushes:              github.Bool(false),
				AllowDeletions:                github.Bool(false),
			},
		},
		{
			name: "mixed settings with some advanced rules",
			protection: &github.Protection{
				RequiredPullRequestReviews: &github.PullRequestReviewsEnforcement{
					RequiredApprovingReviewCount: 2,
					RequireCodeOwnerReviews:      true,
				},
				EnforceAdmins:                  &github.AdminEnforcement{Enabled: true},
				RequiredConversationResolution: &github.RequiredConversationResolution{Enabled: true},
				AllowForcePushes:               &github.AllowForcePushes{Enabled: false},
			},
			expected: &BranchPermissions{
				RequirePullRequestReviews:     github.Bool(true),
				ApproverCount:                 github.Int(2),
				RequireCodeOwners:             github.Bool(true),
				EnforceAdmins:                 github.Bool(true),
				RequireConversationResolution: github.Bool(true),
				AllowForcePushes:              github.Bool(false),
			},
		},
		{
			name: "protection with status checks and advanced settings",
			protection: &github.Protection{
				RequiredStatusChecks: &github.RequiredStatusChecks{
					Strict:   true,
					Contexts: &[]string{"ci/build", "ci/test"},
				},
				RequireLinearHistory: &github.RequireLinearHistory{Enabled: true},
				AllowDeletions:       &github.AllowDeletions{Enabled: false},
			},
			expected: &BranchPermissions{
				RequireStatusChecks:   github.Bool(true),
				RequireUpToDateBranch: github.Bool(true),
				StatusChecks:          []string{"ci/build", "ci/test"},
				RequireLinearHistory:  github.Bool(true),
				AllowDeletions:        github.Bool(false),
			},
		},
		{
			name: "protection with push restrictions - teams and users",
			protection: &github.Protection{
				Restrictions: &github.BranchRestrictions{
					Teams: []*github.Team{
						{Slug: github.String("admin-team")},
						{Slug: github.String("maintainers")},
					},
					Users: []*github.User{
						{Login: github.String("admin-user")},
						{Login: github.String("maintainer")},
					},
				},
			},
			expected: &BranchPermissions{
				RestrictPushes: github.Bool(true),
				PushAllowlist:  []string{"admin-team", "maintainers", "admin-user", "maintainer"},
			},
		},
		{
			name: "protection with push restrictions - teams only",
			protection: &github.Protection{
				Restrictions: &github.BranchRestrictions{
					Teams: []*github.Team{
						{Slug: github.String("core-team")},
					},
					Users: []*github.User{},
				},
			},
			expected: &BranchPermissions{
				RestrictPushes: github.Bool(true),
				PushAllowlist:  []string{"core-team"},
			},
		},
		{
			name: "protection with push restrictions - users only",
			protection: &github.Protection{
				Restrictions: &github.BranchRestrictions{
					Teams: []*github.Team{},
					Users: []*github.User{
						{Login: github.String("admin")},
					},
				},
			},
			expected: &BranchPermissions{
				RestrictPushes: github.Bool(true),
				PushAllowlist:  []string{"admin"},
			},
		},
		{
			name: "protection with push restrictions - nil slug/login handled gracefully",
			protection: &github.Protection{
				Restrictions: &github.BranchRestrictions{
					Teams: []*github.Team{
						{Slug: github.String("valid-team")},
						{Slug: nil}, // This should be skipped
					},
					Users: []*github.User{
						{Login: github.String("valid-user")},
						{Login: nil}, // This should be skipped
					},
				},
			},
			expected: &BranchPermissions{
				RestrictPushes: github.Bool(true),
				PushAllowlist:  []string{"valid-team", "valid-user"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertBranchProtection(tt.protection)

			// Check advanced settings specifically
			if !compareBoolPointers(result.RequireConversationResolution, tt.expected.RequireConversationResolution) {
				t.Errorf("RequireConversationResolution mismatch: got %v, want %v",
					getBoolPointerValue(result.RequireConversationResolution),
					getBoolPointerValue(tt.expected.RequireConversationResolution))
			}

			if !compareBoolPointers(result.RequireLinearHistory, tt.expected.RequireLinearHistory) {
				t.Errorf("RequireLinearHistory mismatch: got %v, want %v",
					getBoolPointerValue(result.RequireLinearHistory),
					getBoolPointerValue(tt.expected.RequireLinearHistory))
			}

			if !compareBoolPointers(result.AllowForcePushes, tt.expected.AllowForcePushes) {
				t.Errorf("AllowForcePushes mismatch: got %v, want %v",
					getBoolPointerValue(result.AllowForcePushes),
					getBoolPointerValue(tt.expected.AllowForcePushes))
			}

			if !compareBoolPointers(result.AllowDeletions, tt.expected.AllowDeletions) {
				t.Errorf("AllowDeletions mismatch: got %v, want %v",
					getBoolPointerValue(result.AllowDeletions),
					getBoolPointerValue(tt.expected.AllowDeletions))
			}

			// Check other settings that were already working
			if !compareBoolPointers(result.RequirePullRequestReviews, tt.expected.RequirePullRequestReviews) {
				t.Errorf("RequirePullRequestReviews mismatch: got %v, want %v",
					getBoolPointerValue(result.RequirePullRequestReviews),
					getBoolPointerValue(tt.expected.RequirePullRequestReviews))
			}

			if !compareIntPointers(result.ApproverCount, tt.expected.ApproverCount) {
				t.Errorf("ApproverCount mismatch: got %v, want %v",
					getIntPointerValue(result.ApproverCount),
					getIntPointerValue(tt.expected.ApproverCount))
			}

			if !compareBoolPointers(result.EnforceAdmins, tt.expected.EnforceAdmins) {
				t.Errorf("EnforceAdmins mismatch: got %v, want %v",
					getBoolPointerValue(result.EnforceAdmins),
					getBoolPointerValue(tt.expected.EnforceAdmins))
			}

			// Check status checks
			if len(result.StatusChecks) != len(tt.expected.StatusChecks) {
				t.Errorf("StatusChecks length mismatch: got %d, want %d", len(result.StatusChecks), len(tt.expected.StatusChecks))
			} else {
				for i, check := range result.StatusChecks {
					if check != tt.expected.StatusChecks[i] {
						t.Errorf("StatusChecks[%d] mismatch: got %s, want %s", i, check, tt.expected.StatusChecks[i])
					}
				}
			}

			// Check push restrictions
			if !compareBoolPointers(result.RestrictPushes, tt.expected.RestrictPushes) {
				t.Errorf("RestrictPushes mismatch: got %v, want %v",
					getBoolPointerValue(result.RestrictPushes),
					getBoolPointerValue(tt.expected.RestrictPushes))
			}

			// Check push allowlist
			if len(result.PushAllowlist) != len(tt.expected.PushAllowlist) {
				t.Errorf("PushAllowlist length mismatch: got %d, want %d", len(result.PushAllowlist), len(tt.expected.PushAllowlist))
			} else {
				for i, actor := range result.PushAllowlist {
					if actor != tt.expected.PushAllowlist[i] {
						t.Errorf("PushAllowlist[%d] mismatch: got %s, want %s", i, actor, tt.expected.PushAllowlist[i])
					}
				}
			}
		})
	}
}

// Helper functions for comparing pointers
func compareBoolPointers(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func compareIntPointers(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func getBoolPointerValue(ptr *bool) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func getIntPointerValue(ptr *int) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}
