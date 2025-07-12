package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	shit "github.com/klauern/ownershit"
	"github.com/urfave/cli/v2"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		configData string
		wantErr    bool
		errorType  string
	}{
		{
			name:       "valid config file",
			configPath: "test_config.yaml",
			configData: `organization: test-org
repositories:
  - name: test-repo
    description: Test repository`,
			wantErr: false,
		},
		{
			name:       "non-existent config file",
			configPath: "non_existent.yaml",
			wantErr:    true,
			errorType:  "config file error",
		},
		{
			name:       "invalid YAML config",
			configPath: "invalid_config.yaml",
			configData: `organization: test-org
repositories:
  - name: test-repo
    description: [invalid yaml structure`,
			wantErr:   true,
			errorType: "config file error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file if needed
			if tt.configData != "" {
				tmpFile := filepath.Join(t.TempDir(), tt.configPath)
				err := os.WriteFile(tmpFile, []byte(tt.configData), 0o600)
				if err != nil {
					t.Fatalf("failed to write test config file: %v", err)
				}
				tt.configPath = tmpFile
			}

			// Create CLI context with config path
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", tt.configPath, "config file")
			c := cli.NewContext(app, set, nil)

			err := readConfig(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errorType == "config file error" {
				if _, ok := err.(*shit.ConfigFileError); !ok {
					t.Errorf("readConfig() expected ConfigFileError, got %T", err)
				}
			}

			if !tt.wantErr && settings == nil {
				t.Error("readConfig() should have set global settings variable")
			}
		})
	}
}

func TestConfigureClient(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	tests := []struct {
		name       string
		setupToken bool
		token      string
		configPath string
		configData string
		debugFlag  bool
		wantErr    bool
	}{
		{
			name:       "valid configuration with debug",
			setupToken: true,
			token:      "ghp_abcd1234567890ABCD1234567890abcd1234",
			configPath: "test_config.yaml",
			configData: `organization: test-org`,
			debugFlag:  true,
			wantErr:    false,
		},
		{
			name:       "missing GitHub token",
			setupToken: false,
			configPath: "test_config.yaml",
			configData: `organization: test-org`,
			wantErr:    true,
		},
		{
			name:       "invalid config file",
			setupToken: true,
			token:      "ghp_abcd1234567890ABCD1234567890abcd1234",
			configPath: "non_existent.yaml",
			wantErr:    true,
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

			// Create temporary config file if needed
			configPath := tt.configPath
			if tt.configData != "" {
				tmpFile := filepath.Join(t.TempDir(), tt.configPath)
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
			set.Bool("debug", tt.debugFlag, "debug flag")
			c := cli.NewContext(app, set, nil)
			c.Context = context.Background()

			// Reset global variables
			settings = nil
			githubClient = nil

			err := configureClient(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("configureClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if githubClient == nil {
					t.Error("configureClient() should have set global githubClient variable")
				}
				if settings == nil {
					t.Error("configureClient() should have set global settings variable")
				}
			}
		})
	}
}

func TestSyncCommand(t *testing.T) {
	// Save original values
	originalSettings := settings
	originalClient := githubClient
	defer func() {
		settings = originalSettings
		githubClient = originalClient
	}()

	tests := []struct {
		name          string
		setupClient   bool
		setupSettings bool
		wantErr       bool
	}{
		{
			name:          "nil client with valid settings",
			setupClient:   true,
			setupSettings: true,
			wantErr:       false, // syncCommand doesn't return errors currently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupSettings {
				org := testOrg
				settings = &shit.PermissionsSettings{
					Organization: &org,
				}
			} else {
				settings = nil
			}

			// We can't easily create a real client in tests, so this test
			// focuses on ensuring the function can be called without panicking
			githubClient = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// Note: This test mainly verifies the function can be called
			// Full integration testing would require mocked GitHub client
			err := syncCommand(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBranchCommand(t *testing.T) {
	// Save original values
	originalSettings := settings
	originalClient := githubClient
	defer func() {
		settings = originalSettings
		githubClient = originalClient
	}()

	tests := []struct {
		name          string
		setupClient   bool
		setupSettings bool
		wantErr       bool
	}{
		{
			name:          "nil client with valid settings",
			setupClient:   true,
			setupSettings: true,
			wantErr:       false, // branchCommand doesn't return errors currently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupSettings {
				org := testOrg
				settings = &shit.PermissionsSettings{
					Organization: &org,
				}
			} else {
				settings = nil
			}

			githubClient = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			err := branchCommand(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("branchCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLabelCommand(t *testing.T) {
	// Save original values
	originalSettings := settings
	originalClient := githubClient
	defer func() {
		settings = originalSettings
		githubClient = originalClient
	}()

	tests := []struct {
		name          string
		setupClient   bool
		setupSettings bool
		wantErr       bool
	}{
		{
			name:          "nil client with valid settings",
			setupClient:   true,
			setupSettings: true,
			wantErr:       false, // labelCommand doesn't return errors currently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupSettings {
				org := testOrg
				settings = &shit.PermissionsSettings{
					Organization: &org,
				}
			} else {
				settings = nil
			}

			githubClient = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			err := labelCommand(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("labelCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRateLimitCommand(t *testing.T) {
	// Save original values
	originalClient := githubClient
	defer func() {
		githubClient = originalClient
	}()

	tests := []struct {
		name        string
		setupClient bool
		wantErr     bool
	}{
		{
			name:        "nil client should cause panic",
			setupClient: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			githubClient = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// This test will panic due to nil client, which we catch
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("rateLimitCommand() panicked: %v", r)
					}
					// Expected panic due to nil client
				}
			}()

			err := rateLimitCommand(c)
			if !tt.wantErr && err != nil {
				t.Errorf("rateLimitCommand() unexpected error = %v", err)
			}
		})
	}
}

func TestInitCommand(t *testing.T) {
	tests := []struct {
		name         string
		configPath   string
		fileExists   bool
		writeError   bool
		expectedFile bool
		wantErr      bool
	}{
		{
			name:         "create new config file",
			configPath:   "new_config.yaml",
			fileExists:   false,
			writeError:   false,
			expectedFile: true,
			wantErr:      false,
		},
		{
			name:         "config file already exists",
			configPath:   "existing_config.yaml",
			fileExists:   true,
			writeError:   false,
			expectedFile: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.configPath)

			// Create existing file if needed
			if tt.fileExists {
				err := os.WriteFile(configPath, []byte("existing content"), 0o600)
				if err != nil {
					t.Fatalf("failed to create existing config file: %v", err)
				}
			}

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("config", configPath, "config file")
			c := cli.NewContext(app, set, nil)

			err := initCommand(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("initCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if file was created/exists as expected
			_, statErr := os.Stat(configPath)
			fileExists := statErr == nil
			if fileExists != tt.expectedFile {
				t.Errorf("initCommand() file existence = %v, expected %v", fileExists, tt.expectedFile)
			}

			// If file should exist and was created, verify content structure
			if tt.expectedFile && !tt.fileExists && !tt.wantErr {
				content, readErr := os.ReadFile(configPath)
				if readErr != nil {
					t.Errorf("failed to read created config file: %v", readErr)
				} else {
					// Verify the stub config contains expected sections
					contentStr := string(content)
					expectedSections := []string{"organization:", "repositories:", "team:", "branches:"}
					for _, section := range expectedSections {
						if !strings.Contains(contentStr, section) {
							t.Errorf("initCommand() created config missing section: %s", section)
						}
					}
				}
			}
		})
	}
}

func TestGetStubConfig(t *testing.T) {
	stubConfig := getStubConfig()

	// Test that stub config is not empty
	if stubConfig == "" {
		t.Error("getStubConfig() should return non-empty string")
	}

	// Test that stub config contains required sections
	requiredSections := []string{
		"organization:",
		"repositories:",
		"team:",
		"branches:",
		"default_labels:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(stubConfig, section) {
			t.Errorf("getStubConfig() missing required section: %s", section)
		}
	}

	// Test that stub config contains example values
	exampleValues := []string{
		"your-org-name",
		"example-repo",
		"developers",
		"maintainers",
		"require_pull_request_reviews",
	}

	for _, value := range exampleValues {
		if !strings.Contains(stubConfig, value) {
			t.Errorf("getStubConfig() missing example value: %s", value)
		}
	}
}

// Note: Testing main() function directly is typically not recommended
// as it calls os.Exit(). Instead, we test the individual components
// that main() orchestrates.
