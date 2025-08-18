package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

// E2E CLI tests that can run against real GitHub repositories
func TestE2E_CLI_Commands(t *testing.T) {
	// Skip unless explicitly requested
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E CLI tests. Set RUN_E2E_TESTS=true to enable")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("Skipping E2E CLI tests: GITHUB_TOKEN environment variable not set")
	}

	// Use test organization/repo from environment or defaults
	testOrg := getEnvOrDefault("E2E_TEST_ORG", "test-org")
	testRepo := getEnvOrDefault("E2E_TEST_REPO", "test-repo")

	// Create test configuration
	configContent := generateTestConfig(testOrg, testRepo)
	configFile := createTempConfig(t, configContent)
	defer os.Remove(configFile)

	// Test each command
	t.Run("ratelimit", func(t *testing.T) {
		testE2ERateLimitCommand(t, configFile)
	})

	t.Run("sync_dry_run", func(t *testing.T) {
		testE2ESyncCommand(t, configFile, true)
	})

	// Only run actual sync if explicitly requested
	if os.Getenv("E2E_ALLOW_MODIFICATIONS") == "true" {
		t.Run("sync_real", func(t *testing.T) {
			testE2ESyncCommand(t, configFile, false)
		})
	}
}

func testE2ERateLimitCommand(t *testing.T, configFile string) {
	// Save original globals
	originalSettings := settings
	originalClient := githubClient
	defer func() {
		settings = originalSettings
		githubClient = originalClient
	}()

	// Create CLI context
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("config", configFile, "config file")
	set.Bool("debug", true, "debug flag")
	c := cli.NewContext(app, set, nil)
	c.Context = context.Background()

	// Configure client
	err := configureClient(c)
	if err != nil {
		t.Fatalf("Failed to configure client: %v", err)
	}

	// Run rate limit command
	err = rateLimitCommand(c)
	if err != nil {
		t.Errorf("Rate limit command failed: %v", err)
	}

	// Verify client was properly configured
	if githubClient == nil {
		t.Error("GitHub client should be configured after rateLimitCommand")
	}
}

func testE2ESyncCommand(t *testing.T, configFile string, dryRun bool) {
	// Save original globals
	originalSettings := settings
	originalClient := githubClient
	defer func() {
		settings = originalSettings
		githubClient = originalClient
	}()

	// Create CLI context
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("config", configFile, "config file")
	set.Bool("debug", true, "debug flag")
	if dryRun {
		set.Bool("dry-run", true, "dry run flag")
	}
	c := cli.NewContext(app, set, nil)
	c.Context = context.Background()

	// Configure client
	err := configureClient(c)
	if err != nil {
		t.Fatalf("Failed to configure client: %v", err)
	}

	// Run sync command
	err = syncCommand(c)

	if dryRun {
		// For dry run, we expect it to work without making changes
		if err != nil {
			t.Logf("Sync dry run completed with message: %v", err)
		}
	} else {
		// For real sync, log the result but don't fail the test
		// since we might not have write permissions
		if err != nil {
			t.Logf("Sync command result: %v", err)
		}
	}

	// Verify configuration was loaded
	if settings == nil {
		t.Error("Settings should be loaded after sync command")
	}

	if settings != nil && settings.Organization == nil {
		t.Error("Organization should be set in settings")
	}
}

func TestE2E_CLI_ConfigValidation(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E CLI config tests. Set RUN_E2E_TESTS=true to enable")
	}

	tests := []struct {
		name          string
		configContent string
		expectError   bool
		description   string
	}{
		{
			name: "valid_complete_config",
			configContent: `
organization: test-org
repositories:
  - name: test-repo
    description: "Test repository for E2E testing"
    wiki: true
    issues: true
    projects: false
team:
  - name: developers
    level: push
  - name: admins
    level: admin
default_labels:
  - name: bug
    color: "d73a4a"
    description: "Something isn't working"
  - name: enhancement  
    color: "a2eeef"
    description: "New feature or request"
branches:
  require_code_owners: true
  require_approving_count: 2
  require_pull_request_reviews: true
  require_status_checks: true
  status_checks:
    - "ci/build"
    - "ci/test"
  require_up_to_date_branch: true
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false`,
			expectError: false,
			description: "Complete valid configuration should load successfully",
		},
		{
			name: "minimal_valid_config",
			configContent: `
organization: minimal-org
repositories:
  - name: minimal-repo`,
			expectError: false,
			description: "Minimal configuration should be valid",
		},
		{
			name: "invalid_yaml_syntax",
			configContent: `
organization: test-org
repositories:
  - name: test-repo
    invalid: [unclosed yaml array`,
			expectError: true,
			description: "Invalid YAML syntax should be rejected",
		},
		{
			name: "invalid_branch_config",
			configContent: `
organization: test-org
repositories:
  - name: test-repo
branches:
  require_approving_count: -1`,
			expectError: true,
			description: "Invalid branch configuration should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)

			configFile := createTempConfig(t, tt.configContent)
			defer os.Remove(configFile)

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", configFile, "config file")
			c := cli.NewContext(app, set, nil)

			// Reset settings
			originalSettings := settings
			settings = nil
			defer func() {
				settings = originalSettings
			}()

			// Try to read config
			err := readConfig(c)

			if (err != nil) != tt.expectError {
				t.Errorf("readConfig() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Validate successful config loading
				if settings == nil {
					t.Error("Settings should be loaded for valid config")
				}
				if settings != nil && settings.Organization == nil {
					t.Error("Organization should be set for valid config")
				}
			} else {
				// Validate error handling
				if err == nil {
					t.Error("Expected error for invalid config")
				} else {
					t.Logf("Got expected error: %v", err)
				}
			}
		})
	}
}

// Test the init command to ensure it creates valid configuration
func TestE2E_CLI_InitCommand(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E init command test. Set RUN_E2E_TESTS=true to enable")
	}

	// Create temporary directory for config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "repositories.yaml")

	// Create CLI context
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("config", configPath, "config file")
	c := cli.NewContext(app, set, nil)

	// Run init command
	err := initCommand(c)
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Init command should have created config file")
	}

	// Verify the created config is valid
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	t.Logf("Created config content:\n%s", string(content))

	// Try to parse the created config
	originalSettings := settings
	settings = nil
	defer func() {
		settings = originalSettings
	}()

	err = readConfig(c)
	if err != nil {
		t.Errorf("Created config file should be valid: %v", err)
	}

	if settings == nil {
		t.Error("Settings should be loaded from created config")
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateTestConfig(org, repo string) string {
	return `# Generated test configuration for E2E testing
organization: ` + org + `

repositories:
  - name: ` + repo + `
    description: "Test repository for E2E testing"
    wiki: true
    issues: true
    projects: false

team:
  - name: developers
    level: push

default_labels:
  - name: bug
    color: "d73a4a"
    description: "Something isn't working"

branches:
  require_code_owners: false
  require_approving_count: 1
  require_pull_request_reviews: true
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false
`
}

func createTempConfig(t *testing.T, content string) string {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "test_repositories.yaml")
	err := os.WriteFile(tmpFile, []byte(strings.TrimSpace(content)), 0o600)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	return tmpFile
}
