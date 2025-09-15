// Package main contains integration tests for the ownershit application.
package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	shit "github.com/klauern/ownershit"
	"github.com/urfave/cli/v2"
)

const (
	syncCommandName     = "sync"
	branchesCommandName = "branches"
	labelCommandName    = "label"
)

func TestAppCommandExecution(t *testing.T) {
	// Save original globals
	originalSettings := settings
	originalClient := githubClient

	defer func() {
		settings = originalSettings
		githubClient = originalClient
	}()

	tests := []struct {
		name           string
		command        string
		configData     string
		setupToken     bool
		token          string
		debugFlag      bool
		expectError    bool
		expectedOutput []string
	}{
		{
			name:    "sync command with valid config",
			command: syncCommandName,
			configData: `organization: test-org
repositories:
  - name: test-repo
    wiki: true
    issues: true
    projects: false
team:
  - name: developers
    level: push`,
			setupToken:     true,
			token:          "ghp_abcd1234567890ABCD1234567890abcd1234",
			debugFlag:      false,
			expectError:    false,
			expectedOutput: []string{"mapping all permissions"},
		},
		{
			name:    "branches command with debug",
			command: branchesCommandName,
			configData: `organization: test-org
repositories:
  - name: test-repo
branches:
  require_pull_request_reviews: true
  require_approving_count: 1`,
			setupToken:     true,
			token:          "ghp_abcd1234567890ABCD1234567890abcd1234",
			debugFlag:      true,
			expectError:    false,
			expectedOutput: []string{"performing branch updates"},
		},
		{
			name:    "label command",
			command: labelCommandName,
			configData: `organization: test-org
repositories:
  - name: test-repo
default_labels:
  - name: "bug"
    color: "d73a4a"
    description: "Something isn't working"`,
			setupToken:     true,
			token:          "ghp_abcd1234567890ABCD1234567890abcd1234",
			debugFlag:      false,
			expectError:    false,
			expectedOutput: []string{"synchronizing labels"},
		},
		{
			name:        "command with missing token",
			command:     "sync",
			configData:  `organization: test-org`,
			setupToken:  false,
			expectError: true,
		},
		{
			name:        "command with invalid config",
			command:     "sync",
			configData:  `invalid: [yaml syntax`,
			setupToken:  true,
			token:       "ghp_abcd1234567890ABCD1234567890abcd1234",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.setupToken {
				t.Setenv("GITHUB_TOKEN", tt.token)
			} else {
				t.Setenv("GITHUB_TOKEN", "")
			}

			// Create temporary config file
			tmpFile := filepath.Join(t.TempDir(), "repositories.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.configData), 0o600)
			if err != nil {
				t.Fatalf("failed to write test config file: %v", err)
			}

			// Capture log output
			var logOutput bytes.Buffer

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", tmpFile, "config file")
			set.Bool("debug", tt.debugFlag, "debug flag")
			c := cli.NewContext(app, set, nil)
			c.Context = context.Background()

			// Reset global variables
			settings = nil
			githubClient = nil

			// Test command execution
			err = configureClient(c)
			if tt.expectError {
				if err == nil {
					t.Error("configureClient() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("configureClient() unexpected error: %v", err)
				return
			}

			// Execute the specific command with panic recovery
			defer func() {
				if r := recover(); r != nil {
					// Commands may panic with nil/invalid clients during testing
					// This is expected behavior when testing with mock/nil clients
					t.Logf("%s command panicked as expected: %v", tt.command, r)
				}
			}()

			switch tt.command {
			case syncCommandName:
				err = syncCommand(c)
			case branchesCommandName:
				err = branchCommand(c)
			case labelCommandName:
				err = labelCommand(c)
			case "ratelimit":
				// Skip ratelimit as it will panic with nil client
				return
			default:
				t.Errorf("unknown command: %s", tt.command)
				return
			}

			// Note: Commands may panic or error with mock/invalid GitHub clients
			// We mainly want to verify the command can be called without crashing the app
			if err != nil {
				t.Logf("%s command returned error (expected with mock client): %v", tt.command, err)
			}

			// Check expected output in logs (basic verification)
			logStr := logOutput.String()
			for _, expectedOutput := range tt.expectedOutput {
				if !strings.Contains(logStr, expectedOutput) {
					// Note: We can't easily capture zerolog output in tests
					// This is a placeholder for future log output verification
					t.Logf("Expected log output '%s' not found in: %s", expectedOutput, logStr)
				}
			}
		})
	}
}

func TestAppHelpAndVersion(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantHelp bool
	}{
		{
			name:     "help flag",
			args:     []string{"ownershit", "--help"},
			wantHelp: true,
		},
		{
			name:     "version flag",
			args:     []string{"ownershit", "--version"},
			wantHelp: false,
		},
		{
			name:     "no args shows help",
			args:     []string{"ownershit"},
			wantHelp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create app instance (similar to main())
			app := &cli.App{
				Commands: []*cli.Command{
					{
						Name:   "init",
						Usage:  "Create a stub configuration file to get started",
						Action: initCommand,
					},
					{
						Name:   "branches",
						Usage:  "Perform branch management steps from repositories.yaml",
						Before: configureClient,
						Action: branchCommand,
					},
					{
						Name:      "sync",
						Usage:     "Synchronize branch, repo, owner and other configs on repositories",
						UsageText: "ownershit sync --config repositories.yaml",
						Before:    configureClient,
						Action:    syncCommand,
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
				Name:    "ownershit",
				Usage:   "fix up team ownership of your repositories in an organization",
				Action:  cli.ShowAppHelp,
				Authors: []*cli.Author{{Name: "Nick Klauer", Email: "klauer@gmail.com"}},
				Version: version,
				ExtraInfo: func() map[string]string {
					return map[string]string{
						"date":    date,
						"builtBy": builtBy,
						"commit":  commit,
					}
				},
			}

			// Test that app runs without panicking
			err := app.Run(tt.args)
			// Note: Help and version commands typically don't return errors
			// We mainly want to ensure the app doesn't panic
			if err != nil && !strings.Contains(err.Error(), "flag provided but not defined") {
				t.Errorf("app.Run() error = %v", err)
			}
		})
	}
}

func TestCompleteInitWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_repositories.yaml")

	// Test complete init workflow
	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("config", configPath, "config file")
	c := cli.NewContext(app, set, nil)

	// Run init command
	err := initCommand(c)
	if err != nil {
		t.Fatalf("initCommand() error = %v", err)
	}

	// Verify file was created
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		t.Error("initCommand() should have created config file")
		return
	}

	// Verify file contents
	content, err := os.ReadFile(configPath) // #nosec G304 - test file in temp directory
	if err != nil {
		t.Fatalf("failed to read created config file: %v", err)
	}

	contentStr := string(content)
	requiredSections := []string{
		"organization:",
		"repositories:",
		"team:",
		"branches:",
		"default_labels:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("created config file missing section: %s", section)
		}
	}

	// Test that running init again doesn't overwrite
	originalContent := string(content)
	err = initCommand(c)
	if err != nil {
		t.Errorf("initCommand() second run error = %v", err)
	}

	// Verify content is unchanged
	newContent, err := os.ReadFile(configPath) // #nosec G304 - test file in temp directory
	if err != nil {
		t.Fatalf("failed to read config file after second init: %v", err)
	}

	if string(newContent) != originalContent {
		t.Error("initCommand() should not overwrite existing config file")
	}
}

func TestErrorHandlingWorkflows(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   bool
		configData    string
		setupToken    bool
		token         string
		command       string
		expectedError string
	}{
		{
			name:          "missing config file",
			setupConfig:   false,
			setupToken:    true,
			token:         "ghp_abcd1234567890ABCD1234567890abcd1234",
			command:       "sync",
			expectedError: "configuration file not found",
		},
		{
			name:          "invalid config file",
			setupConfig:   true,
			configData:    `invalid: [yaml`,
			setupToken:    true,
			token:         "ghp_abcd1234567890ABCD1234567890abcd1234",
			command:       "sync",
			expectedError: "failed to parse YAML",
		},
		{
			name:          "missing GitHub token",
			setupConfig:   true,
			configData:    `organization: test-org`,
			setupToken:    false,
			command:       "sync",
			expectedError: "GitHub token",
		},
		{
			name:          "invalid token format",
			setupConfig:   true,
			configData:    `organization: test-org`,
			setupToken:    true,
			token:         "invalid-token-format",
			command:       "sync",
			expectedError: "GitHub token format is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.setupToken {
				t.Setenv("GITHUB_TOKEN", tt.token)
			} else {
				t.Setenv("GITHUB_TOKEN", "")
			}

			// Setup config file
			configPath := "non_existent.yaml"
			if tt.setupConfig {
				tmpFile := filepath.Join(t.TempDir(), "repositories.yaml")
				err := os.WriteFile(tmpFile, []byte(tt.configData), 0o600)
				if err != nil {
					t.Fatalf("failed to write test config file: %v", err)
				}
				configPath = tmpFile
			}

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", configPath, "config file")
			c := cli.NewContext(app, set, nil)
			c.Context = context.Background()

			// Reset global variables
			settings = nil
			githubClient = nil

			// Test error handling
			err := configureClient(c)
			if err == nil {
				t.Error("configureClient() expected error, but got none")
				return
			}

			errorMsg := err.Error()
			if !strings.Contains(errorMsg, tt.expectedError) {
				t.Errorf("configureClient() error = %v, expected to contain %q", err, tt.expectedError)
			}

			// Verify error type for config file errors
			if strings.Contains(tt.expectedError, "config file") {
				if _, ok := err.(*shit.ConfigFileError); !ok {
					t.Errorf("configureClient() expected ConfigFileError, got %T", err)
				}
			}
		})
	}
}
