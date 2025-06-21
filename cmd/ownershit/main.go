package main

import (
	"fmt"
	"os"

	shit "github.com/klauern/ownershit"
	"github.com/klauern/ownershit/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	settings     *shit.PermissionsSettings
	githubClient *shit.GitHubClient
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Create a stub configuration file to get started",
				Before: func(c *cli.Context) error {
					if c.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					return nil
				},
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
			{
				Name:        "archive",
				Usage:       "Archive repositories",
				Before:      configureClient,
				Subcommands: cmd.ArchiveSubcommands,
			},
			{
				Name:   "label",
				Usage:  "set default labels for repositories",
				Before: configureClient,
				Action: labelCommand,
			},
			{
				Name:   "ratelimit",
				Usage:  "get ratelimit information for the GitHub GraphQL v4 API",
				Before: configureClient,
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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("running app")
	}
}

func configureClient(c *cli.Context) error {
	err := readConfig(c)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	if c.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	client, err := shit.NewSecureGitHubClient(c.Context)
	if err != nil {
		log.Err(err).
			Str("operation", "initializeGitHubClient").
			Msg("GitHub client initialization failed")
		return fmt.Errorf("failed to initialize GitHub client: %w", err)
	}
	githubClient = client
	return nil
}

func readConfig(c *cli.Context) error {
	configPath := c.String("config")
	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Err(err).
				Str("configPath", configPath).
				Str("operation", "readConfigFile").
				Msg("config file not found")
			fmt.Printf("❌ Configuration file '%s' not found.\n\n", configPath)
			fmt.Printf("To get started, run:\n")
			fmt.Printf("  ownershit init\n\n")
			fmt.Printf("This will create a stub configuration file with examples and comments.\n")
			return shit.NewConfigFileError(configPath, "read", "configuration file not found - run 'ownershit init' to create one", err)
		}
		log.Err(err).
			Str("configPath", configPath).
			Str("operation", "readConfigFile").
			Msg("config file error")
		return shit.NewConfigFileError(configPath, "read", "failed to read configuration file", err)
	}

	settings = &shit.PermissionsSettings{}
	if err := yaml.Unmarshal(file, settings); err != nil {
		log.Err(err).
			Str("configPath", configPath).
			Str("operation", "parseConfigFile").
			Msg("YAML unmarshal error with config file")
		return shit.NewConfigFileError(configPath, "parse", "failed to parse YAML configuration", err)
	}

	return nil
}

func syncCommand(c *cli.Context) error {
	log.Info().Msg("mapping all permissions for repositories")
	shit.MapPermissions(settings, githubClient)
	return nil
}

func branchCommand(c *cli.Context) error {
	log.Info().Msg("performing branch updates on repositories")
	shit.UpdateBranchMergeStrategies(settings, githubClient)
	return nil
}

func labelCommand(c *cli.Context) error {
	log.Info().Msg("synchronizing labels on repositories")
	shit.SyncLabels(settings, githubClient)
	return nil
}

func rateLimitCommand(c *cli.Context) error {
	log.Info().Msg("getting ratelimit information")
	githubClient.GetRateLimit()
	return nil
}

func initCommand(c *cli.Context) error {
	configPath := c.String("config")
	
	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		log.Warn().Str("configPath", configPath).Msg("configuration file already exists")
		fmt.Printf("Configuration file '%s' already exists.\n", configPath)
		fmt.Printf("Remove it first if you want to create a new one.\n")
		return nil
	}
	
	stubConfig := getStubConfig()
	
	err := os.WriteFile(configPath, []byte(stubConfig), 0644)
	if err != nil {
		log.Err(err).Str("configPath", configPath).Msg("failed to write configuration file")
		return fmt.Errorf("failed to create configuration file: %w", err)
	}
	
	log.Info().Str("configPath", configPath).Msg("configuration file created")
	fmt.Printf("✅ Created configuration file: %s\n\n", configPath)
	fmt.Printf("Next steps:\n")
	fmt.Printf("1. Edit %s and update the 'organization' field with your GitHub organization\n", configPath)
	fmt.Printf("2. Add your repositories to the 'repositories' section\n")
	fmt.Printf("3. Configure team permissions as needed\n")
	fmt.Printf("4. Set up your GITHUB_TOKEN environment variable\n")
	fmt.Printf("5. Run 'ownershit sync' to apply the configuration\n\n")
	fmt.Printf("For help, run: ownershit --help\n")
	
	return nil
}

func getStubConfig() string {
	return `# ownershit configuration file
# This file defines repository ownership, team permissions, and branch protection rules

# Your GitHub organization name (REQUIRED - update this!)
organization: your-org-name

# Branch protection rules applied to all repositories
branches:
  # Pull request requirements
  require_pull_request_reviews: true
  require_approving_count: 1
  require_code_owners: false
  
  # Merge strategy controls
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false
  
  # Status checks (uncomment and configure as needed)
  # require_status_checks: true
  # status_checks:
  #   - "ci/build"
  #   - "ci/test"
  # require_up_to_date_branch: true
  
  # Advanced protection (uncomment as needed)
  # enforce_admins: true
  # restrict_pushes: false
  # push_allowlist: []
  # require_conversation_resolution: true
  # require_linear_history: false
  # allow_force_pushes: false
  # allow_deletions: false

# Team permissions for repositories
team:
  - name: developers
    level: push
  - name: maintainers
    level: admin
  # Available levels: admin, push, pull

# Repository configurations
repositories:
  - name: example-repo
    wiki: true
    issues: true
    projects: false
  # Add more repositories here

# Default labels applied to all repositories (optional)
default_labels:
  - name: "bug"
    color: "d73a4a"
    emoji: "🐛"
    description: "Something isn't working"
  - name: "enhancement"
    color: "a2eeef"
    emoji: "✨"
    description: "New feature or request"
  - name: "documentation"
    color: "0075ca"
    emoji: "📚"
    description: "Improvements or additions to documentation"

# Environment setup:
# export GITHUB_TOKEN=your_github_token_here
`
}
