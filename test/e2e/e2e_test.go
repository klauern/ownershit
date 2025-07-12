package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/klauern/ownershit"
	"github.com/klauern/ownershit/v4api"
)

// E2ETestConfig holds configuration for end-to-end tests
type E2ETestConfig struct {
	Organization string
	TestRepo     string
	Token        string
	UserLogin    string
}

// setupE2ETest prepares the environment for end-to-end testing
func setupE2ETest(t *testing.T) *E2ETestConfig {
	t.Helper()

	// Check if we're in CI or have explicit E2E testing enabled
	if os.Getenv("RUN_E2E_TESTS") != "true" && os.Getenv("CI") != "true" {
		t.Skip("Skipping E2E tests. Set RUN_E2E_TESTS=true to enable")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("Skipping E2E tests: GITHUB_TOKEN environment variable not set")
	}

	// Use environment variables or defaults for test configuration
	org := os.Getenv("E2E_TEST_ORG")
	if org == "" {
		org = "test-org" // Default test organization
	}

	repo := os.Getenv("E2E_TEST_REPO")
	if repo == "" {
		repo = "test-repo" // Default test repository
	}

	userLogin := os.Getenv("E2E_TEST_USER")
	if userLogin == "" {
		t.Skip("Skipping E2E tests: E2E_TEST_USER environment variable not set")
	}

	return &E2ETestConfig{
		Organization: org,
		TestRepo:     repo,
		Token:        token,
		UserLogin:    userLogin,
	}
}

func TestE2E_RateLimit(t *testing.T) {
	_ = setupE2ETest(t)
	
	// Test using the V4 API directly
	v4Client := v4api.NewGHv4Client()
	
	rateLimit, err := v4Client.GetRateLimit()
	if err != nil {
		t.Fatalf("Failed to get rate limit: %v", err)
	}

	t.Logf("Rate limit info: Limit=%d, Remaining=%d, Reset=%s", 
		rateLimit.RateLimit.Limit, rateLimit.RateLimit.Remaining, rateLimit.RateLimit.ResetAt.Format(time.RFC3339))

	// Validate rate limit response
	if rateLimit.RateLimit.Limit <= 0 {
		t.Error("Rate limit should be positive")
	}
	if rateLimit.RateLimit.Remaining < 0 {
		t.Error("Remaining requests should not be negative")
	}
	if rateLimit.RateLimit.ResetAt.IsZero() {
		t.Error("Reset time should be set")
	}
}

func TestE2E_GetTeams(t *testing.T) {
	config := setupE2ETest(t)
	
	v4Client := v4api.NewGHv4Client()
	
	teams, err := v4Client.GetTeams(config.Organization)
	if err != nil {
		// This might fail if the org doesn't exist or we don't have access
		t.Logf("Warning: Could not get teams for org %s: %v", config.Organization, err)
		return
	}

	t.Logf("Found %d teams in organization %s", len(teams), config.Organization)
	
	// Basic validation of team structure
	for _, team := range teams {
		if team.Node.Name == "" {
			t.Error("Team name should not be empty")
		}
		t.Logf("Team: %s (description: %s)", team.Node.Name, team.Node.Description)
	}
}

func TestE2E_RepositoryOperations(t *testing.T) {
	config := setupE2ETest(t)
	
	ctx := context.Background()
	client, err := ownershit.NewSecureGitHubClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create secure GitHub client: %v", err)
	}

	// Test basic repository access by trying a read-only operation
	t.Run("CheckRepositoryAccess", func(t *testing.T) {
		// Test error handling when a repo doesn't exist - this is safe and read-only
		err := client.SetBranchProtectionFallback(config.Organization, "non-existent-repo-12345", "main", nil)
		if err == nil {
			t.Error("Expected error when accessing non-existent repository")
		}
		t.Logf("Got expected error for non-existent repo: %v", err)
	})
}

func TestE2E_ConfigurationValidation(t *testing.T) {
	config := setupE2ETest(t)
	
	// Test various configuration scenarios
	tests := []struct {
		name        string
		settings    *ownershit.PermissionsSettings
		expectError bool
		description string
	}{
		{
			name: "valid minimal configuration",
			settings: &ownershit.PermissionsSettings{
				Organization: &config.Organization,
				Repositories: []*ownershit.Repository{
					{
						Name:     &config.TestRepo,
						Wiki:     boolPtr(true),
						Issues:   boolPtr(true),
						Projects: boolPtr(false),
					},
				},
				TeamPermissions: []*ownershit.Permissions{
					{
						Team:  stringPtr("developers"),
						Level: stringPtr("push"),
					},
				},
				BranchPermissions: ownershit.BranchPermissions{
					RequirePullRequestReviews: boolPtr(true),
					ApproverCount:            intPtr(1),
					RequireCodeOwners:        boolPtr(false),
				},
			},
			expectError: false,
			description: "Basic valid configuration should pass validation",
		},
		{
			name: "invalid approver count",
			settings: &ownershit.PermissionsSettings{
				Organization: &config.Organization,
				BranchPermissions: ownershit.BranchPermissions{
					RequirePullRequestReviews: boolPtr(true),
					ApproverCount:            intPtr(-1), // Invalid
				},
			},
			expectError: true,
			description: "Negative approver count should fail validation",
		},
		{
			name: "excessively high approver count",
			settings: &ownershit.PermissionsSettings{
				Organization: &config.Organization,
				BranchPermissions: ownershit.BranchPermissions{
					RequirePullRequestReviews: boolPtr(true),
					ApproverCount:            intPtr(150), // Too high
				},
			},
			expectError: true,
			description: "Excessively high approver count should fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)
			
			err := ownershit.ValidateBranchPermissions(&tt.settings.BranchPermissions)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateBranchPermissions() error = %v, expectError %v", err, tt.expectError)
			}

			if err != nil {
				t.Logf("Validation error (expected=%v): %v", tt.expectError, err)
			}
		})
	}
}

func TestE2E_TokenValidation(t *testing.T) {
	config := setupE2ETest(t)
	
	// Test token validation with real token
	err := ownershit.ValidateGitHubToken(config.Token)
	if err != nil {
		t.Errorf("Token validation failed for real token: %v", err)
	}

	// Test invalid token formats
	invalidTokens := []string{
		"invalid",
		"ghp_short",
		"not_a_token",
		"",
	}

	for _, token := range invalidTokens {
		err := ownershit.ValidateGitHubToken(token)
		if err == nil {
			t.Errorf("Expected validation to fail for invalid token: %s", token)
		}
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}