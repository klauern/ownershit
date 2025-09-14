package ownershit

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v66/github"
	"go.uber.org/mock/gomock"

	"github.com/klauern/ownershit/mocks"
)

func TestConvertBranchProtection(t *testing.T) {
	tests := []struct {
		name       string
		protection *github.Protection
		expected   *BranchPermissions
	}{
		{
			name:       "nil protection should return empty permissions with explicit false values",
			protection: nil,
			expected: &BranchPermissions{
				RequireStatusChecks:   github.Bool(false),
				RequireUpToDateBranch: github.Bool(false),
				RestrictPushes:        github.Bool(false),
			},
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
				RequireStatusChecks:           github.Bool(false),
				RequireUpToDateBranch:         github.Bool(false),
				RestrictPushes:                github.Bool(false),
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
				RequireStatusChecks:           github.Bool(false),
				RequireUpToDateBranch:         github.Bool(false),
				RestrictPushes:                github.Bool(false),
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
				RequireStatusChecks:           github.Bool(false),
				RequireUpToDateBranch:         github.Bool(false),
				RestrictPushes:                github.Bool(false),
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
				RestrictPushes:        github.Bool(false),
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
				RequireStatusChecks:   github.Bool(false),
				RequireUpToDateBranch: github.Bool(false),
				RestrictPushes:        github.Bool(true),
				PushAllowlist:         []string{"admin-team", "maintainers", "admin-user", "maintainer"},
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
				RequireStatusChecks:   github.Bool(false),
				RequireUpToDateBranch: github.Bool(false),
				RestrictPushes:        github.Bool(true),
				PushAllowlist:         []string{"core-team"},
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
				RequireStatusChecks:   github.Bool(false),
				RequireUpToDateBranch: github.Bool(false),
				RestrictPushes:        github.Bool(true),
				PushAllowlist:         []string{"admin"},
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
				RequireStatusChecks:   github.Bool(false),
				RequireUpToDateBranch: github.Bool(false),
				RestrictPushes:        github.Bool(true),
				PushAllowlist:         []string{"valid-team", "valid-user"},
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

func TestGetRepositoryDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepositoriesService(ctrl)

	client := &GitHubClient{
		Repositories: mockRepo,
		Context:      context.Background(),
	}

	tests := []struct {
		name           string
		owner          string
		repo           string
		mockRepoInfo   *github.Repository
		expectedResult *repositoryDetails
		expectError    bool
	}{
		{
			name:  "complete repository with all fields",
			owner: "testowner",
			repo:  "testrepo",
			mockRepoInfo: &github.Repository{
				HasWiki:       github.Bool(true),
				HasIssues:     github.Bool(true),
				HasProjects:   github.Bool(false),
				DefaultBranch: github.String("main"),
				Private:       github.Bool(false),
				Archived:      github.Bool(false),
				IsTemplate:    github.Bool(true),
				Description:   github.String("Test repository description"),
				Homepage:      github.String("https://example.com"),
			},
			expectedResult: &repositoryDetails{
				Wiki:          github.Bool(true),
				Issues:        github.Bool(true),
				Projects:      github.Bool(false),
				DefaultBranch: github.String("main"),
				Private:       github.Bool(false),
				Archived:      github.Bool(false),
				Template:      github.Bool(true),
				Description:   github.String("Test repository description"),
				Homepage:      github.String("https://example.com"),
			},
			expectError: false,
		},
		{
			name:  "private archived repository",
			owner: "testowner",
			repo:  "archived-repo",
			mockRepoInfo: &github.Repository{
				HasWiki:       github.Bool(false),
				HasIssues:     github.Bool(false),
				HasProjects:   github.Bool(false),
				DefaultBranch: github.String("master"),
				Private:       github.Bool(true),
				Archived:      github.Bool(true),
				IsTemplate:    github.Bool(false),
				Description:   github.String("An archived private repository"),
				Homepage:      nil, // No homepage URL
			},
			expectedResult: &repositoryDetails{
				Wiki:          github.Bool(false),
				Issues:        github.Bool(false),
				Projects:      github.Bool(false),
				DefaultBranch: github.String("master"),
				Private:       github.Bool(true),
				Archived:      github.Bool(true),
				Template:      github.Bool(false),
				Description:   github.String("An archived private repository"),
				Homepage:      nil,
			},
			expectError: false,
		},
		{
			name:  "repository with nil optional fields",
			owner: "testowner",
			repo:  "minimal-repo",
			mockRepoInfo: &github.Repository{
				HasWiki:       github.Bool(true),
				HasIssues:     github.Bool(true),
				HasProjects:   github.Bool(true),
				DefaultBranch: github.String("develop"),
				Private:       github.Bool(false),
				Archived:      github.Bool(false),
				IsTemplate:    github.Bool(false),
				Description:   nil, // No description
				Homepage:      nil, // No homepage
			},
			expectedResult: &repositoryDetails{
				Wiki:          github.Bool(true),
				Issues:        github.Bool(true),
				Projects:      github.Bool(true),
				DefaultBranch: github.String("develop"),
				Private:       github.Bool(false),
				Archived:      github.Bool(false),
				Template:      github.Bool(false),
				Description:   nil,
				Homepage:      nil,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				Get(gomock.Any(), tt.owner, tt.repo).
				Return(tt.mockRepoInfo, nil, nil).
				Times(1)

			result, err := getRepositoryDetails(client, tt.owner, tt.repo)

			if tt.expectError && err == nil {
				t.Errorf("expected an error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil && tt.expectedResult != nil {
				t.Errorf("expected result but got nil")
				return
			}
			if result != nil && tt.expectedResult == nil {
				t.Errorf("expected nil but got result")
				return
			}

			// Compare all fields
			if !compareBoolPointers(result.Wiki, tt.expectedResult.Wiki) {
				t.Errorf("Wiki mismatch: got %v, want %v",
					getBoolPointerValue(result.Wiki),
					getBoolPointerValue(tt.expectedResult.Wiki))
			}

			if !compareBoolPointers(result.Issues, tt.expectedResult.Issues) {
				t.Errorf("Issues mismatch: got %v, want %v",
					getBoolPointerValue(result.Issues),
					getBoolPointerValue(tt.expectedResult.Issues))
			}

			if !compareBoolPointers(result.Projects, tt.expectedResult.Projects) {
				t.Errorf("Projects mismatch: got %v, want %v",
					getBoolPointerValue(result.Projects),
					getBoolPointerValue(tt.expectedResult.Projects))
			}

			if !compareStringPointers(result.DefaultBranch, tt.expectedResult.DefaultBranch) {
				t.Errorf("DefaultBranch mismatch: got %v, want %v",
					getStringPointerValue(result.DefaultBranch),
					getStringPointerValue(tt.expectedResult.DefaultBranch))
			}

			if !compareBoolPointers(result.Private, tt.expectedResult.Private) {
				t.Errorf("Private mismatch: got %v, want %v",
					getBoolPointerValue(result.Private),
					getBoolPointerValue(tt.expectedResult.Private))
			}

			if !compareBoolPointers(result.Archived, tt.expectedResult.Archived) {
				t.Errorf("Archived mismatch: got %v, want %v",
					getBoolPointerValue(result.Archived),
					getBoolPointerValue(tt.expectedResult.Archived))
			}

			if !compareBoolPointers(result.Template, tt.expectedResult.Template) {
				t.Errorf("Template mismatch: got %v, want %v",
					getBoolPointerValue(result.Template),
					getBoolPointerValue(tt.expectedResult.Template))
			}

			if !compareStringPointers(result.Description, tt.expectedResult.Description) {
				t.Errorf("Description mismatch: got %v, want %v",
					getStringPointerValue(result.Description),
					getStringPointerValue(tt.expectedResult.Description))
			}

			if !compareStringPointers(result.Homepage, tt.expectedResult.Homepage) {
				t.Errorf("Homepage mismatch: got %v, want %v",
					getStringPointerValue(result.Homepage),
					getStringPointerValue(tt.expectedResult.Homepage))
			}
		})
	}
}

func compareStringPointers(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func getStringPointerValue(ptr *string) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func TestImportRepositoryConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepositoriesService(ctrl)
	mockIssues := mocks.NewMockIssuesService(ctrl)

	client := &GitHubClient{
		Repositories: mockRepo,
		Issues:       mockIssues,
		Context:      context.Background(),
	}

	// Mock repository details
	mockRepoInfo := &github.Repository{
		HasWiki:          github.Bool(true),
		HasIssues:        github.Bool(true),
		HasProjects:      github.Bool(false),
		DefaultBranch:    github.String("main"),
		Private:          github.Bool(false),
		Archived:         github.Bool(false),
		IsTemplate:       github.Bool(true),
		Description:      github.String("Test repository description"),
		Homepage:         github.String("https://example.com"),
		AllowMergeCommit: github.Bool(true),
		AllowSquashMerge: github.Bool(true),
		AllowRebaseMerge: github.Bool(false),
	}

	// Set up mock expectations
	// For getRepositoryDetails
	mockRepo.EXPECT().
		Get(gomock.Any(), "testowner", "testrepo").
		Return(mockRepoInfo, nil, nil).
		Times(1)

	// For getTeamPermissions
	mockRepo.EXPECT().
		ListTeams(gomock.Any(), "testowner", "testrepo", nil).
		Return([]*github.Team{}, nil, nil).
		Times(1)

	// For getBranchProtectionRules (trying main and master branches)
	mockRepo.EXPECT().
		GetBranchProtection(gomock.Any(), "testowner", "testrepo", "main").
		Return(nil, nil, fmt.Errorf("no protection found")).
		Times(1)

	mockRepo.EXPECT().
		GetBranchProtection(gomock.Any(), "testowner", "testrepo", "master").
		Return(nil, nil, fmt.Errorf("no protection found")).
		Times(1)

	// For getRepositoryLabels
	mockIssues.EXPECT().
		ListLabels(gomock.Any(), "testowner", "testrepo", gomock.Any()).
		Return([]*github.Label{}, &github.Response{}, nil).
		Times(1)

	// Execute the function
	config, err := ImportRepositoryConfig("testowner", "testrepo", client, true)

	// Verify results
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config but got nil")
	}

	if len(config.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(config.Repositories))
	}

	repo := config.Repositories[0]

	// Check all new fields are properly set
	if repo.DefaultBranch == nil || *repo.DefaultBranch != "main" {
		t.Errorf("expected DefaultBranch to be 'main', got %v", getStringPointerValue(repo.DefaultBranch))
	}

	if repo.Private == nil || *repo.Private != false {
		t.Errorf("expected Private to be false, got %v", getBoolPointerValue(repo.Private))
	}

	if repo.Archived == nil || *repo.Archived != false {
		t.Errorf("expected Archived to be false, got %v", getBoolPointerValue(repo.Archived))
	}

	if repo.Template == nil || *repo.Template != true {
		t.Errorf("expected Template to be true, got %v", getBoolPointerValue(repo.Template))
	}

	if repo.Description == nil || *repo.Description != "Test repository description" {
		t.Errorf("expected Description to be 'Test repository description', got %v", getStringPointerValue(repo.Description))
	}

	if repo.Homepage == nil || *repo.Homepage != "https://example.com" {
		t.Errorf("expected Homepage to be 'https://example.com', got %v", getStringPointerValue(repo.Homepage))
	}
}
