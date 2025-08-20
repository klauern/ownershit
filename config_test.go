package ownershit

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/shurcooL/githubv4"
	"go.uber.org/mock/gomock"
)

var ErrDummyConfigError = errors.New("dummy error")

// Helper function for creating int pointers in tests.
func intPtr(i int) *int {
	return &i
}

func TestValidateBranchPermissions(t *testing.T) {
	tests := []struct {
		name    string
		perms   *BranchPermissions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil permissions should error",
			perms:   nil,
			wantErr: true,
			errMsg:  "branch permissions cannot be nil",
		},
		{
			name: "valid configuration",
			perms: &BranchPermissions{
				RequireCodeOwners:         boolPtr(true),
				ApproverCount:             intPtr(2),
				RequirePullRequestReviews: boolPtr(true),
				RequireStatusChecks:       boolPtr(true),
				StatusChecks:              []string{"ci/build", "ci/test"},
				RequireUpToDateBranch:     boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "negative approver count should error",
			perms: &BranchPermissions{
				ApproverCount: intPtr(-1),
			},
			wantErr: true,
			errMsg:  "approver count cannot be negative",
		},
		{
			name: "require status checks with empty list should error",
			perms: &BranchPermissions{
				RequireStatusChecks: boolPtr(true),
				StatusChecks:        []string{},
			},
			wantErr: true,
			errMsg:  "RequireStatusChecks is enabled but no status checks specified",
		},
		{
			name: "empty status check name should error",
			perms: &BranchPermissions{
				RequireStatusChecks: boolPtr(true),
				StatusChecks:        []string{"ci/build", "", "ci/test"},
			},
			wantErr: true,
			errMsg:  "empty status check name",
		},
		{
			name: "require approving reviews when pull request reviews disabled should error",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(false),
				ApproverCount:             intPtr(1),
			},
			wantErr: true,
			errMsg:  "cannot require approving reviews when require_pull_request_reviews is false",
		},
		{
			name: "require code owners when pull request reviews disabled should error",
			perms: &BranchPermissions{
				RequirePullRequestReviews: boolPtr(false),
				RequireCodeOwners:         boolPtr(true),
			},
			wantErr: true,
			errMsg:  "cannot require code owner reviews when require_pull_request_reviews is false",
		},
		{
			name: "require up-to-date branch without status checks should error",
			perms: &BranchPermissions{
				RequireStatusChecks:   boolPtr(false),
				RequireUpToDateBranch: boolPtr(true),
			},
			wantErr: true,
			errMsg:  "RequireUpToDateBranch requires RequireStatusChecks to be enabled",
		},
		{
			name: "restrict pushes with empty allowlist should error",
			perms: &BranchPermissions{
				RestrictPushes: boolPtr(true),
				PushAllowlist:  []string{},
			},
			wantErr: true,
			errMsg:  "RestrictPushes is enabled but no users/teams specified in PushAllowlist",
		},
		{
			name: "empty push allowlist entry should error",
			perms: &BranchPermissions{
				RestrictPushes: boolPtr(true),
				PushAllowlist:  []string{"team1", "", "team2"},
			},
			wantErr: true,
			errMsg:  "empty entry in PushAllowlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchPermissions(tt.perms)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateBranchPermissions() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func generateDefaultPermissionsSettings() *PermissionsSettings {
	return &PermissionsSettings{
		BranchPermissions: BranchPermissions{
			AllowSquashMerge: boolPtr(false),
			AllowMergeCommit: boolPtr(false),
			AllowRebaseMerge: boolPtr(false),
		},
		Organization: stringPtr("klauern"),
		Repositories: []*Repository{
			{
				Name:     stringPtr("test"),
				Issues:   boolPtr(false),
				Projects: boolPtr(false),
				Wiki:     boolPtr(false),
			},
		},
		TeamPermissions: []*Permissions{
			{
				Team:  stringPtr("klauern"),
				Level: stringPtr(string(Admin)),
			},
		},
	}
}

func TestMapPermissions(t *testing.T) {
	mocks := setupMocks(t)

	mocks.teamMock.EXPECT().AddTeamRepoBySlug(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).AnyTimes().Return(defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().
		Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(ErrDummyConfigError)

	mocks.repoMock.
		EXPECT().
		Edit(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		Return(nil, defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil).
		Do(func(ctx context.Context, query *GetRepoQuery, things map[string]interface{}) {
			query.Repository.ID = githubv4.ID("12345")
			query.Repository.HasIssuesEnabled = false
			query.Repository.HasProjectsEnabled = false
			query.Repository.HasWikiEnabled = false
		})

	type args struct {
		settings *PermissionsSettings
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test nothing",
			args: args{
				settings: generateDefaultPermissionsSettings(),
				client:   mocks.client,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MapPermissions(tt.args.settings, tt.args.client)
		})
	}
}

func TestUpdateBranchMergeStrategies(t *testing.T) {
	mocks := setupMocks(t)

	mocks.teamMock.EXPECT().AddTeamRepoBySlug(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).AnyTimes().Return(defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().
		Mutate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(ErrDummyConfigError)

	mocks.repoMock.
		EXPECT().
		Edit(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any()).
		Return(nil, defaultGoodResponse, nil)

	mocks.graphMock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil).
		Do(func(ctx context.Context, query *GetRepoQuery, things map[string]interface{}) {
			query.Repository.ID = githubv4.ID("12345")
			query.Repository.HasIssuesEnabled = false
			query.Repository.HasProjectsEnabled = false
			query.Repository.HasWikiEnabled = false
		})

	type args struct {
		settings *PermissionsSettings
		client   *GitHubClient
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default run",
			args: args{
				settings: generateDefaultPermissionsSettings(),
				client:   mocks.client,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateBranchMergeStrategies(tt.args.settings, tt.args.client)
		})
	}
}

// func TestSyncLabels(t *testing.T) {
// 	mocks := setupMocks(t)
// 	mocks.issuesMock.EXPECT().ListLabels(gomock.Any(), gomock.Any(),
// 		gomock.Any()).Return()
// }

var gitHubTokenValidationTests = []struct {
	name     string
	token    string
	wantErr  error
	wantType string
}{
	{
		name:     "empty token",
		token:    "",
		wantErr:  ErrTokenEmpty,
		wantType: "empty",
	},
	{
		name:     "whitespace only token",
		token:    "   \t\n  ",
		wantErr:  ErrTokenEmpty,
		wantType: "whitespace",
	},
	{
		name:     "valid classic personal access token",
		token:    "ghp_abcd1234567890ABCD1234567890abcd1234",
		wantErr:  nil,
		wantType: "classic",
	},
	{
		name:     "valid fine-grained personal access token",
		token:    "github_pat_abcd1234567890ABCD1234567890abcd1234567890ABCD1234567890abcd12",
		wantErr:  nil,
		wantType: "fine-grained",
	},
	{
		name:     "valid GitHub App token",
		token:    "ghs_abcd1234567890ABCD1234567890abcd1234",
		wantErr:  nil,
		wantType: "app",
	},
	{
		name:     "valid OAuth token",
		token:    "gho_abcd1234567890ABCD1234567890abcd1234",
		wantErr:  nil,
		wantType: "oauth",
	},
	{
		name:     "valid refresh token",
		token:    "ghr_abcd1234567890ABCD1234567890abcd1234",
		wantErr:  nil,
		wantType: "refresh",
	},
	{
		name:     "valid SAML token",
		token:    "ghu_abcd1234567890ABCD1234567890abcd1234",
		wantErr:  nil,
		wantType: "saml",
	},
	{
		name:     "invalid classic token - too short",
		token:    "ghp_abcd1234567890ABCD1234567890abcd123",
		wantErr:  ErrTokenInvalid,
		wantType: "invalid",
	},
	{
		name:     "invalid classic token - too long",
		token:    "ghp_abcd1234567890ABCD1234567890abcd12345",
		wantErr:  ErrTokenInvalid,
		wantType: "invalid",
	},
	{
		name:     "invalid token - wrong prefix",
		token:    "abc_1234567890123456789012345678901234567890",
		wantErr:  ErrTokenInvalid,
		wantType: "invalid",
	},
	{
		name:     "invalid token - no prefix",
		token:    "1234567890123456789012345678901234567890",
		wantErr:  ErrTokenInvalid,
		wantType: "invalid",
	},
	{
		name:     "invalid fine-grained token - too short",
		token:    "github_pat_123456789012345678901",
		wantErr:  ErrTokenInvalid,
		wantType: "invalid",
	},
}

func TestValidateGitHubToken(t *testing.T) {
	for _, tt := range gitHubTokenValidationTests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitHubToken(tt.token)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateGitHubToken() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateGitHubToken() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidateGitHubToken() error = %v, wantErr nil", err)
			}
		})
	}
}

func TestGetValidatedGitHubToken(t *testing.T) {
	// Save original env var
	original := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if original == "" {
			os.Unsetenv("GITHUB_TOKEN")
		} else {
			os.Setenv("GITHUB_TOKEN", original)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		unsetEnv bool
		wantErr  bool
		errType  error
	}{
		{
			name:     "environment variable not set",
			unsetEnv: true,
			wantErr:  true,
			errType:  ErrTokenNotFound,
		},
		{
			name:     "empty environment variable",
			envValue: "",
			wantErr:  true,
			errType:  ErrTokenNotFound,
		},
		{
			name:     "invalid token format",
			envValue: "invalid_token_format",
			wantErr:  true,
			errType:  ErrTokenInvalid,
		},
		{
			name:     "valid classic token",
			envValue: "ghp_abcd1234567890ABCD1234567890abcd1234",
			wantErr:  false,
		},
		{
			name:     "valid fine-grained token",
			envValue: "github_pat_abcd1234567890ABCD1234567890abcd1234567890ABCD1234567890abcd12",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unsetEnv {
				os.Unsetenv("GITHUB_TOKEN")
			} else {
				os.Setenv("GITHUB_TOKEN", tt.envValue)
			}

			token, err := GetValidatedGitHubToken()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetValidatedGitHubToken() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("GetValidatedGitHubToken() error = %v, want error type %v", err, tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("GetValidatedGitHubToken() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if token != tt.envValue {
					t.Errorf("GetValidatedGitHubToken() token = %v, want %v", token, tt.envValue)
				}
			}
		})
	}
}

func TestValidateSchemaVersion(t *testing.T) {
	tests := []struct {
		name    string
		version *string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil version should be valid (defaults to legacy)",
			version: nil,
			wantErr: false,
		},
		{
			name:    "empty version should be valid",
			version: stringPtr(""),
			wantErr: false,
		},
		{
			name:    "current version should be valid",
			version: stringPtr(CurrentSchemaVersion),
			wantErr: false,
		},
		{
			name:    "legacy version should be valid",
			version: stringPtr(LegacySchemaVersion),
			wantErr: false,
		},
		{
			name:    "unsupported version should error",
			version: stringPtr("2.0"),
			wantErr: true,
			errMsg:  "unsupported schema version",
		},
		{
			name:    "invalid version format should error",
			version: stringPtr("invalid"),
			wantErr: true,
			errMsg:  "unsupported schema version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchemaVersion(tt.version)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateSchemaVersion() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateSchemaVersion() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("ValidateSchemaVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigrateConfigurationSchema(t *testing.T) {
	tests := []struct {
		name            string
		settings        *PermissionsSettings
		wantErr         bool
		errMsg          string
		expectedVersion string
	}{
		{
			name:     "nil settings should error",
			settings: nil,
			wantErr:  true,
			errMsg:   "settings cannot be nil for migration",
		},
		{
			name: "current version needs no migration",
			settings: &PermissionsSettings{
				Version:      stringPtr(CurrentSchemaVersion),
				Organization: stringPtr("test-org"),
			},
			wantErr:         false,
			expectedVersion: CurrentSchemaVersion,
		},
		{
			name: "legacy version should migrate to current",
			settings: &PermissionsSettings{
				Version:      stringPtr(LegacySchemaVersion),
				Organization: stringPtr("test-org"),
			},
			wantErr:         false,
			expectedVersion: CurrentSchemaVersion,
		},
		{
			name: "nil version should migrate to current",
			settings: &PermissionsSettings{
				Version:      nil,
				Organization: stringPtr("test-org"),
			},
			wantErr:         false,
			expectedVersion: CurrentSchemaVersion,
		},
		{
			name: "unsupported version should error",
			settings: &PermissionsSettings{
				Version:      stringPtr("2.0"),
				Organization: stringPtr("test-org"),
			},
			wantErr: true,
			errMsg:  "cannot migrate unsupported schema version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MigrateConfigurationSchema(tt.settings)

			if tt.wantErr {
				if err == nil {
					t.Errorf("MigrateConfigurationSchema() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("MigrateConfigurationSchema() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("MigrateConfigurationSchema() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.settings != nil && tt.settings.Version != nil {
					if *tt.settings.Version != tt.expectedVersion {
						t.Errorf("MigrateConfigurationSchema() version = %v, want %v", *tt.settings.Version, tt.expectedVersion)
					}
				}
			}
		})
	}
}

func TestGetSchemaVersion(t *testing.T) {
	tests := []struct {
		name     string
		settings *PermissionsSettings
		expected string
	}{
		{
			name:     "nil settings should return legacy version",
			settings: nil,
			expected: LegacySchemaVersion,
		},
		{
			name: "nil version should return legacy version",
			settings: &PermissionsSettings{
				Version: nil,
			},
			expected: LegacySchemaVersion,
		},
		{
			name: "empty version should return legacy version",
			settings: &PermissionsSettings{
				Version: stringPtr(""),
			},
			expected: LegacySchemaVersion,
		},
		{
			name: "current version should return current version",
			settings: &PermissionsSettings{
				Version: stringPtr(CurrentSchemaVersion),
			},
			expected: CurrentSchemaVersion,
		},
		{
			name: "legacy version should return legacy version",
			settings: &PermissionsSettings{
				Version: stringPtr(LegacySchemaVersion),
			},
			expected: LegacySchemaVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSchemaVersion(tt.settings)
			if result != tt.expected {
				t.Errorf("GetSchemaVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsCurrentSchemaVersion(t *testing.T) {
	tests := []struct {
		name     string
		settings *PermissionsSettings
		expected bool
	}{
		{
			name:     "nil settings should return false",
			settings: nil,
			expected: false,
		},
		{
			name: "current version should return true",
			settings: &PermissionsSettings{
				Version: stringPtr(CurrentSchemaVersion),
			},
			expected: true,
		},
		{
			name: "legacy version should return false",
			settings: &PermissionsSettings{
				Version: stringPtr(LegacySchemaVersion),
			},
			expected: false,
		},
		{
			name: "nil version should return false",
			settings: &PermissionsSettings{
				Version: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCurrentSchemaVersion(tt.settings)
			if result != tt.expected {
				t.Errorf("IsCurrentSchemaVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}
