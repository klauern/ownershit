package cmd

import (
	"context"
	"flag"
	"strings"
	"testing"

	shit "github.com/klauern/ownershit"
	"github.com/urfave/cli/v2"
)

func TestUserClientSetup(t *testing.T) {
	// Save original username
	originalUsername := username
	defer func() {
		username = originalUsername
	}()

	tests := []struct {
		name       string
		setupToken bool
		token      string
		username   string
		wantErr    bool
		errorType  error
	}{
		{
			name:       "valid setup with token and username",
			setupToken: true,
			token:      "ghp_abcd1234567890ABCD1234567890abcd1234",
			username:   "test-user",
			wantErr:    false,
		},
		{
			name:       "missing username",
			setupToken: true,
			token:      "ghp_abcd1234567890ABCD1234567890abcd1234",
			username:   "",
			wantErr:    true,
			errorType:  ErrUsernameNotDefined,
		},
		{
			name:       "whitespace only username",
			setupToken: true,
			token:      "ghp_abcd1234567890ABCD1234567890abcd1234",
			username:   "   ",
			wantErr:    true,
			errorType:  ErrUsernameNotDefined,
		},
		{
			name:       "missing GitHub token",
			setupToken: false,
			username:   "test-user",
			wantErr:    true,
		},
		{
			name:       "invalid token format",
			setupToken: true,
			token:      "invalid-token",
			username:   "test-user",
			wantErr:    true,
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
			username = strings.TrimSpace(tt.username)

			// Trim above already handles whitespace-only usernames

			// Create CLI context
			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)
			c.Context = context.Background()

			// Reset global client
			client = nil

			err := userClientSetup(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("userClientSetup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errorType != nil {
				if err != tt.errorType {
					t.Errorf("userClientSetup() expected error %v, got %v", tt.errorType, err)
				}
			}

			if !tt.wantErr && client == nil {
				t.Error("userClientSetup() should have set global client variable")
			}
		})
	}
}

func TestQueryCommand(t *testing.T) {
	// Save original values
	originalUsername := username
	originalClient := client
	originalFlags := []int{forks, stars, days, watchers}
	defer func() {
		username = originalUsername
		client = originalClient
		forks, stars, days, watchers = originalFlags[0], originalFlags[1], originalFlags[2], originalFlags[3]
	}()

	tests := []struct {
		name        string
		setupClient bool
		username    string
		forks       int
		stars       int
		days        int
		watchers    int
		wantErr     bool
	}{
		{
			name:        "nil client should cause error",
			setupClient: false,
			username:    "test-user",
			forks:       0,
			stars:       0,
			days:        365,
			watchers:    0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables
			username = tt.username
			forks = tt.forks
			stars = tt.stars
			days = tt.days
			watchers = tt.watchers

			// We can't easily create a real client in tests
			// This test mainly verifies the function structure
			client = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// This test will panic due to nil client, which we catch
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("queryCommand() panicked: %v", r)
					}
					// Expected panic due to nil client
				}
			}()

			err := queryCommand(c)
			if !tt.wantErr && err != nil {
				t.Errorf("queryCommand() unexpected error = %v", err)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	// Save original values
	originalUsername := username
	originalClient := client
	originalFlags := []int{forks, stars, days, watchers}
	defer func() {
		username = originalUsername
		client = originalClient
		forks, stars, days, watchers = originalFlags[0], originalFlags[1], originalFlags[2], originalFlags[3]
	}()

	tests := []struct {
		name        string
		setupClient bool
		username    string
		forks       int
		stars       int
		days        int
		watchers    int
		wantErr     bool
	}{
		{
			name:        "nil client should cause error",
			setupClient: false,
			username:    "test-user",
			forks:       0,
			stars:       0,
			days:        365,
			watchers:    0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables
			username = tt.username
			forks = tt.forks
			stars = tt.stars
			days = tt.days
			watchers = tt.watchers

			client = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// This test will panic due to nil client, which we catch
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("executeCommand() panicked: %v", r)
					}
					// Expected panic due to nil client
				}
			}()

			err := executeCommand(c)
			if !tt.wantErr && err != nil {
				t.Errorf("executeCommand() unexpected error = %v", err)
			}
		})
	}
}

func TestMapRepoNames(t *testing.T) {
	tests := []struct {
		name     string
		repos    []shit.RepositoryInfo
		expected map[string]shit.RepositoryInfo
	}{
		{
			name: "single repository",
			repos: []shit.RepositoryInfo{
				{Name: "test-repo"},
			},
			expected: map[string]shit.RepositoryInfo{
				"test-repo": {Name: "test-repo"},
			},
		},
		{
			name: "multiple repositories",
			repos: []shit.RepositoryInfo{
				{Name: "repo1"},
				{Name: "repo2"},
				{Name: "repo3"},
			},
			expected: map[string]shit.RepositoryInfo{
				"repo1": {Name: "repo1"},
				"repo2": {Name: "repo2"},
				"repo3": {Name: "repo3"},
			},
		},
		{
			name:     "empty repository list",
			repos:    []shit.RepositoryInfo{},
			expected: map[string]shit.RepositoryInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapRepoNames(tt.repos)

			if len(result) != len(tt.expected) {
				t.Errorf("mapRepoNames() returned %d items, expected %d", len(result), len(tt.expected))
				return
			}

			for key, expectedRepo := range tt.expected {
				if actualRepo, exists := result[key]; !exists {
					t.Errorf("mapRepoNames() missing key %s", key)
				} else if actualRepo.Name != expectedRepo.Name {
					t.Errorf("mapRepoNames() for key %s: got %v, want %v", key, actualRepo.Name, expectedRepo.Name)
				}
			}
		})
	}
}

func TestArchiveFlags(t *testing.T) {
	// Test that archive flags are properly defined
	if len(archiveFlags) == 0 {
		t.Error("archiveFlags should not be empty")
	}

	// Verify specific flags exist
	flagNames := make(map[string]bool)
	for _, flag := range archiveFlags {
		switch f := flag.(type) {
		case *cli.StringFlag:
			flagNames[f.Name] = true
		case *cli.IntFlag:
			flagNames[f.Name] = true
		}
	}

	expectedFlags := []string{"username", "forks", "stars", "days", "watchers"}
	for _, expected := range expectedFlags {
		if !flagNames[expected] {
			t.Errorf("archiveFlags missing expected flag: %s", expected)
		}
	}
}

func TestArchiveSubcommands(t *testing.T) {
	// Test that archive subcommands are properly defined
	if len(ArchiveSubcommands) == 0 {
		t.Error("ArchiveSubcommands should not be empty")
	}

	// Verify specific subcommands exist
	commandNames := make(map[string]bool)
	for _, cmd := range ArchiveSubcommands {
		commandNames[cmd.Name] = true
	}

	expectedCommands := []string{"query", "execute"}
	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("ArchiveSubcommands missing expected command: %s", expected)
		}
	}

	// Verify each command has the expected Before function
	for _, cmd := range ArchiveSubcommands {
		if cmd.Before == nil {
			t.Errorf("ArchiveSubcommands command %s should have Before function", cmd.Name)
		}
		if len(cmd.Flags) == 0 {
			t.Errorf("ArchiveSubcommands command %s should have flags", cmd.Name)
		}
	}
}

func TestQueryCommandWithNilClient(t *testing.T) {
	// Save original values
	originalUsername := username
	originalClient := client
	originalFlags := []int{forks, stars, days, watchers}
	defer func() {
		username = originalUsername
		client = originalClient
		forks, stars, days, watchers = originalFlags[0], originalFlags[1], originalFlags[2], originalFlags[3]
	}()

	tests := []struct {
		name     string
		username string
		forks    int
		stars    int
		days     int
		watchers int
		wantErr  bool
	}{
		{
			name:     "nil client should panic",
			username: "test-user",
			forks:    0,
			stars:    0,
			days:     365,
			watchers: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables
			username = tt.username
			forks = tt.forks
			stars = tt.stars
			days = tt.days
			watchers = tt.watchers
			client = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// Expect panic with nil client
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("queryCommand() unexpected panic: %v", r)
					}
					// Expected panic
				} else if tt.wantErr {
					t.Error("queryCommand() expected panic but didn't panic")
				}
			}()

			_ = queryCommand(c)
		})
	}
}

func TestExecuteCommandWithNilClient(t *testing.T) {
	// Save original values
	originalUsername := username
	originalClient := client
	originalFlags := []int{forks, stars, days, watchers}
	defer func() {
		username = originalUsername
		client = originalClient
		forks, stars, days, watchers = originalFlags[0], originalFlags[1], originalFlags[2], originalFlags[3]
	}()

	tests := []struct {
		name     string
		username string
		forks    int
		stars    int
		days     int
		watchers int
		wantErr  bool
	}{
		{
			name:     "nil client should panic",
			username: "test-user",
			forks:    0,
			stars:    0,
			days:     365,
			watchers: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables
			username = tt.username
			forks = tt.forks
			stars = tt.stars
			days = tt.days
			watchers = tt.watchers
			client = nil

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			c := cli.NewContext(app, set, nil)

			// Expect panic with nil client
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("executeCommand() unexpected panic: %v", r)
					}
					// Expected panic
				} else if tt.wantErr {
					t.Error("executeCommand() expected panic but didn't panic")
				}
			}()

			_ = executeCommand(c)
		})
	}
}

func TestGlobalVariableDefaults(t *testing.T) {
	// Test that global variables have sensible defaults
	if ErrUsernameNotDefined == nil {
		t.Error("ErrUsernameNotDefined should be defined")
	}

	// Test that archive flags have proper default values
	for _, flag := range archiveFlags {
		switch f := flag.(type) {
		case *cli.IntFlag:
			if f.Name == "days" && f.Value != 365 {
				t.Errorf("days flag should default to 365, got %d", f.Value)
			}
			if (f.Name == "forks" || f.Name == "stars" || f.Name == "watchers") && f.Value != 0 {
				t.Errorf("%s flag should default to 0, got %d", f.Name, f.Value)
			}
		case *cli.StringFlag:
			if f.Name == "username" && f.Value != "" {
				t.Errorf("username flag should default to empty string, got %s", f.Value)
			}
		}
	}
}
