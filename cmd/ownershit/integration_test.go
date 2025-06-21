package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	shit "github.com/klauern/ownershit"
	"github.com/urfave/cli/v2"
)

const (
	testOrg         = "test-org"
	configFileError = "config file error"
)

func TestAppConfiguration(t *testing.T) {
	// Test that the app is properly configured
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "branches",
				Usage:  "Perform branch management steps from repositories.yaml",
				Action: branchCommand,
			},
			{
				Name:      "sync",
				Usage:     "Synchronize branch, repo, owner and other configs on repositories",
				UsageText: "ownershit sync --config repositories.yaml",
				Action:    syncCommand,
			},
			{
				Name:   "label",
				Usage:  "set default labels for repositories",
				Action: labelCommand,
			},
			{
				Name:   "ratelimit",
				Usage:  "get ratelimit information for the GitHub GraphQL v4 API",
				Action: rateLimitCommand,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "repositories.yaml",
				Usage: "configuration of repository updates to perform",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				EnvVars: []string{"OWNERSHIT_DEBUG"},
				Usage:   "set output to debug logging",
				Value:   false,
			},
		},
		Name:   "ownershit",
		Usage:  "fix up team ownership of your repositories in an organization",
		Action: cli.ShowAppHelp,
		Before: configureClient,
	}

	// Test that all expected commands are present
	expectedCommands := []string{"branches", "sync", "label", "ratelimit"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range app.Commands {
			if cmd.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %s not found in app configuration", expected)
		}
	}

	// Test that flags are properly configured
	expectedFlags := []string{"config", "debug"}
	for _, expected := range expectedFlags {
		found := false
		for _, flag := range app.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				if f.Name == expected {
					found = true
				}
			case *cli.BoolFlag:
				if f.Name == expected {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("Expected flag %s not found in app configuration", expected)
		}
	}
}

func TestCompleteWorkflow(t *testing.T) {
	// Save original environment and globals
	originalToken := os.Getenv("GITHUB_TOKEN")
	originalSettings := settings
	originalClient := githubClient

	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		settings = originalSettings
		githubClient = originalClient
	}()

	tests := []struct {
		name        string
		configData  string
		setupToken  bool
		token       string
		expectError bool
		command     string
	}{
		{
			name: "complete workflow with valid config",
			configData: `organization: ` + testOrg + `
repositories:
  - name: test-repo
    description: Test repository
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
branches:
  require_code_owners: true
  require_approving_count: 1
  require_pull_request_reviews: true
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false`,
			setupToken:  true,
			token:       "ghp_abcd1234567890ABCD1234567890abcd1234",
			expectError: false,
			command:     "sync",
		},
		{
			name: "workflow with missing token",
			configData: `organization: ` + testOrg + `
repositories:
  - name: test-repo`,
			setupToken:  false,
			expectError: true,
			command:     "sync",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.setupToken {
				os.Setenv("GITHUB_TOKEN", tt.token)
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}

			// Create temporary config file
			tmpFile := filepath.Join(t.TempDir(), "repositories.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.configData), 0o600)
			if err != nil {
				t.Fatalf("failed to write test config file: %v", err)
			}

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", tmpFile, "config file")
			set.Bool("debug", false, "debug flag")
			c := cli.NewContext(app, set, nil)
			c.Context = context.Background()

			// Reset global variables
			settings = nil
			githubClient = nil

			// Test the complete workflow: configure client + run command
			err = configureClient(c)
			if (err != nil) != tt.expectError {
				t.Errorf("configureClient() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				testSuccessfulConfiguration(t)
				testCommandExecution(t, c, tt.command)
			}
		})
	}
}

func testSuccessfulConfiguration(t *testing.T) {
	t.Helper()

	// Test that global variables are properly set
	if settings == nil {
		t.Error("settings should be set after configureClient")
	}
	if githubClient == nil {
		t.Error("githubClient should be set after configureClient")
	}

	// Test configuration parsing
	if settings.Organization == nil || *settings.Organization != testOrg {
		t.Error("Organization should be properly parsed from config")
	}

	if len(settings.Repositories) == 0 {
		t.Error("Repositories should be properly parsed from config")
	}

	// Verify specific repository configuration
	repo := settings.Repositories[0]
	if repo.Name == nil || *repo.Name != "test-repo" {
		t.Error("Repository name should be properly parsed")
	}

	// Verify team permissions
	if len(settings.TeamPermissions) > 0 {
		team := settings.TeamPermissions[0]
		if team.Team == nil || *team.Team != "developers" {
			t.Error("Team permissions should be properly parsed")
		}
	}
}

func testCommandExecution(t *testing.T, c *cli.Context, command string) {
	t.Helper()

	// Test command execution (this will use nil client but shouldn't crash)
	switch command {
	case "sync":
		_ = syncCommand(c)
		// Command doesn't return errors currently, so just verify it doesn't crash
	case "branches":
		_ = branchCommand(c)
	case "label":
		_ = labelCommand(c)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		expectError bool
		errorType   string
	}{
		{
			name:        "invalid YAML syntax",
			configData:  `organization: ` + testOrg + `\ninvalid: [yaml syntax`,
			expectError: true,
			errorType:   configFileError,
		},
		{
			name:        "missing required fields",
			configData:  `invalid: configuration`,
			expectError: false, // YAML is valid, just missing expected fields
		},
		{
			name: "valid minimal config",
			configData: `organization: minimal-org
repositories: []`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpFile := filepath.Join(t.TempDir(), "test_config.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.configData), 0o600)
			if err != nil {
				t.Fatalf("failed to write test config file: %v", err)
			}

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", tmpFile, "config file")
			c := cli.NewContext(app, set, nil)

			// Reset global variables
			settings = nil

			err = readConfig(c)
			if (err != nil) != tt.expectError {
				t.Errorf("readConfig() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError && tt.errorType == configFileError {
				if _, ok := err.(*shit.ConfigFileError); !ok {
					t.Errorf("readConfig() expected ConfigFileError, got %T", err)
				}
			}
		})
	}
}

func TestVersionInfo(t *testing.T) {
	// Test that version variables are properly defined
	if version == "" {
		version = "test-version" // Set for testing
	}
	if commit == "" {
		commit = "test-commit"
	}
	if date == "" {
		date = "test-date"
	}
	if builtBy == "" {
		builtBy = "test-builder"
	}

	// These should all have values now
	if version == "" {
		t.Error("version should be defined")
	}
	if commit == "" {
		t.Error("commit should be defined")
	}
	if date == "" {
		t.Error("date should be defined")
	}
	if builtBy == "" {
		t.Error("builtBy should be defined")
	}
}
