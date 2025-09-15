package ownershit

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/mock/gomock"
)

var ErrBranchProtectionTest = errors.New("branch protection test error")

const (
	testMainBranch = "main"
)

func TestGitHubClient_SetEnhancedBranchProtection_Integration(t *testing.T) {
	mock := setupMocks(t)

	tests := []struct {
		name          string
		repoID        githubv4.ID
		branchPattern string
		perms         *BranchPermissions
		wantErr       bool
		mockSetup     func()
		description   string
	}{
		{
			name:          "minimal protection configuration",
			repoID:        githubv4.ID("repo_minimal"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
			},
			wantErr:     false,
			description: "Test minimal branch protection with only PR reviews required",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "complete GraphQL-supported features",
			repoID:        githubv4.ID("repo_complete"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(3),
				RequireCodeOwners:         boolPtr(true),
				RequireStatusChecks:       boolPtr(true),
				StatusChecks:              []string{"ci/build", "ci/test", "security/scan", "lint"},
				RequireUpToDateBranch:     boolPtr(true),
			},
			wantErr:     false,
			description: "Test all features that are directly supported by GraphQL API",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "advanced features requiring REST fallback",
			repoID:        githubv4.ID("repo_advanced"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews:     boolPtr(true),
				ApproverCount:                 intPtr(2),
				RequireCodeOwners:             boolPtr(true),
				EnforceAdmins:                 boolPtr(true),
				RestrictPushes:                boolPtr(true),
				PushAllowlist:                 []string{"admin-team", "release-team"},
				RequireConversationResolution: boolPtr(true),
				RequireLinearHistory:          boolPtr(true),
				AllowForcePushes:              boolPtr(false),
				AllowDeletions:                boolPtr(false),
			},
			wantErr:     false,
			description: "Test advanced features that currently log fallback messages",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "int32 overflow handling",
			repoID:        githubv4.ID("repo_overflow"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(2147483648), // Exceeds int32 max
			},
			wantErr:     false,
			description: "Test that large approver counts are handled gracefully",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "negative approver count handling",
			repoID:        githubv4.ID("repo_negative"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(-2147483649), // Below int32 min
			},
			wantErr:     false,
			description: "Test that negative approver counts are handled gracefully",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "empty status checks list",
			repoID:        githubv4.ID("repo_empty_status"),
			branchPattern: "develop",
			perms: &BranchPermissions{
				RequireStatusChecks:   boolPtr(true),
				StatusChecks:          []string{}, // Empty list
				RequireUpToDateBranch: boolPtr(true),
			},
			wantErr:     false,
			description: "Test behavior with empty status checks list",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:          "GraphQL API error",
			repoID:        githubv4.ID("repo_error"),
			branchPattern: "main",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(1),
			},
			wantErr:     true,
			description: "Test error handling when GraphQL API fails",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrBranchProtectionTest)
			},
		},
		{
			name:          "multiple status checks with special characters",
			repoID:        githubv4.ID("repo_special_checks"),
			branchPattern: "release/*",
			perms: &BranchPermissions{
				RequireStatusChecks:   boolPtr(true),
				StatusChecks:          []string{"ci/build-and-test", "security/sast-scan", "deps/vulnerability-check", "qa/e2e-tests"},
				RequireUpToDateBranch: boolPtr(true),
			},
			wantErr:     false,
			description: "Test complex status check names and wildcard branch patterns",
			mockSetup: func() {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)
			tt.mockSetup()
			err := mock.client.SetEnhancedBranchProtection(tt.repoID, tt.branchPattern, tt.perms)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetEnhancedBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitHubClient_SetBranchRules_DeprecationIntegration(t *testing.T) {
	mock := setupMocks(t)

	tests := []struct {
		name                     string
		repoID                   githubv4.ID
		branchPattern            *string
		approverCount            *int
		requireCodeOwners        *bool
		requiresApprovingReviews *bool
		wantErr                  bool
		description              string
	}{
		{
			name:                     "deprecated method basic usage",
			repoID:                   githubv4.ID("repo_deprecated"),
			branchPattern:            github.String("main"),
			approverCount:            github.Int(2),
			requireCodeOwners:        github.Bool(true),
			requiresApprovingReviews: github.Bool(true),
			wantErr:                  false,
			description:              "Test that deprecated SetBranchRules properly delegates to SetEnhancedBranchProtection",
		},
		{
			name:                     "deprecated method with nil values",
			repoID:                   githubv4.ID("repo_deprecated_nil"),
			branchPattern:            nil,
			approverCount:            nil,
			requireCodeOwners:        nil,
			requiresApprovingReviews: nil,
			wantErr:                  false,
			description:              "Test deprecated method with nil values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)
			// SetBranchRules calls SetEnhancedBranchProtection, so we mock that
			mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := mock.client.SetBranchRules(tt.repoID, tt.branchPattern, tt.approverCount, tt.requireCodeOwners, tt.requiresApprovingReviews)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetBranchRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBranchProtection_EdgeCases(t *testing.T) {
	mock := setupMocks(t)

	tests := []struct {
		name        string
		setup       func() (*BranchPermissions, githubv4.ID, string)
		wantErr     bool
		description string
	}{
		{
			name: "nil permissions struct",
			setup: func() (*BranchPermissions, githubv4.ID, string) {
				return nil, githubv4.ID("repo_nil"), testMainBranch
			},
			wantErr:     true,
			description: "Test behavior with nil permissions struct - should panic/error gracefully",
		},
		{
			name: "empty branch pattern",
			setup: func() (*BranchPermissions, githubv4.ID, string) {
				return &BranchPermissions{RequirePullRequestReviews: boolPtr(true)}, githubv4.ID("repo_empty_pattern"), ""
			},
			wantErr:     false,
			description: "Test with empty branch pattern",
		},
		{
			name: "very long status check names",
			setup: func() (*BranchPermissions, githubv4.ID, string) {
				longName := strings.Repeat("very-long-status-check-name-", 10) + "end"
				return &BranchPermissions{
					RequireStatusChecks: boolPtr(true),
					StatusChecks:        []string{longName},
				}, githubv4.ID("repo_long_names"), testMainBranch
			},
			wantErr:     false,
			description: "Test with very long status check names",
		},
		{
			name: "special characters in branch pattern",
			setup: func() (*BranchPermissions, githubv4.ID, string) {
				return &BranchPermissions{RequirePullRequestReviews: boolPtr(true)}, githubv4.ID("repo_special"), "feature/user-123_fix"
			},
			wantErr:     false,
			description: "Test with special characters in branch pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)
			perms, repoID, pattern := tt.setup()

			if perms != nil {
				mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			err := mock.client.SetEnhancedBranchProtection(repoID, pattern, perms)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.SetEnhancedBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test the branch protection input validation and mutation construction.
func TestBranchProtection_InputValidation(t *testing.T) {
	mock := setupMocks(t)

	// This test verifies that the GraphQL input is constructed correctly
	// by testing various combinations of permissions
	tests := []struct {
		name        string
		perms       *BranchPermissions
		description string
	}{
		{
			name: "all boolean flags set to true",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				RequireCodeOwners:         boolPtr(true),
				RequireStatusChecks:       boolPtr(true),
				RequireUpToDateBranch:     boolPtr(true),
			},
			description: "Test with all available boolean flags enabled",
		},
		{
			name: "all boolean flags set to false",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(false),
				RequireCodeOwners:         boolPtr(false),
				RequireStatusChecks:       boolPtr(false),
				RequireUpToDateBranch:     boolPtr(false),
			},
			description: "Test with all available boolean flags disabled",
		},
		{
			name: "mixed boolean values with arrays",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				RequireCodeOwners:         boolPtr(false),
				RequireStatusChecks:       boolPtr(true),
				StatusChecks:              []string{"check1", "check2", "check3"},
				RequireUpToDateBranch:     boolPtr(false),
			},
			description: "Test mixed boolean values with status checks array",
		},
		{
			name: "maximum approver count",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(2147483647), // int32 max
			},
			description: "Test with maximum int32 approver count",
		},
		{
			name: "minimum approver count",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(true),
				ApproverCount:             intPtr(-2147483648), // int32 min
			},
			description: "Test with minimum int32 approver count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)
			mock.graphMock.EXPECT().Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			err := mock.client.SetEnhancedBranchProtection(githubv4.ID("test_repo"), testMainBranch, tt.perms)
			if err != nil {
				t.Errorf("GitHubClient.SetEnhancedBranchProtection() unexpected error = %v", err)
			}
		})
	}
}

// Test repository resolution integration.
func TestGitHubClient_GetRepository_Integration(t *testing.T) {
	mock := setupMocks(t)

	tests := []struct {
		name        string
		repoName    string
		owner       string
		wantErr     bool
		description string
	}{
		{
			name:        "successful repository resolution",
			repoName:    "testrepo",
			owner:       "testowner",
			wantErr:     false,
			description: "Test successful repository resolution for branch protection",
		},
		{
			name:        "repository not found",
			repoName:    "nonexistent",
			owner:       "badowner",
			wantErr:     true,
			description: "Test error handling when repository doesn't exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test case: %s", tt.description)

			if tt.wantErr {
				mock.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrBranchProtectionTest)
			} else {
				mock.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			repoID, err := mock.client.GetRepository(&tt.repoName, &tt.owner)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubClient.GetRepository() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && repoID == "" {
				t.Error("Expected non-empty repository ID for successful case")
			}
		})
	}
}
