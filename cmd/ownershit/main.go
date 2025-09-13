package main

import (
	"fmt"
	"os"
	"strings"

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

// main is the entry point for the ownershit CLI application.
// It configures logging, constructs the command-line interface with subcommands
// (init, branches, sync, archive, label, ratelimit, import, import-csv, permissions),
// and runs the app, terminating with a fatal error if execution fails.
func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Create a stub configuration file to get started",
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
			},
			{
				Name:      "sync",
				Usage:     "Synchronize branch, repo, owner and other configs on repositories",
				UsageText: "ownershit sync --config repositories.yaml",
				Before:    configureClient,
				Action:    syncCommand,
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
			},
			{
				Name:   "ratelimit",
				Usage:  "get ratelimit information for the GitHub GraphQL v4 API",
				Before: configureClient,
				Action: rateLimitCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						EnvVars: []string{"OWNERSHIT_DEBUG"},
						Usage:   "set output to debug logging",
						Value:   false,
					},
				},
			},
			{
				Name:      "import",
				Usage:     "Import repository configuration from GitHub and output as YAML",
				UsageText: "ownershit import [--output filename.yaml] owner/repo",
				Before:    configureImportClient,
				Action:    importCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "output file path (default: stdout)",
					},
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						EnvVars: []string{"OWNERSHIT_DEBUG"},
						Usage:   "set output to debug logging",
						Value:   false,
					},
				},
			},
			{
				Name:      "import-csv",
				Usage:     "Import repository configurations from multiple GitHub repositories and output as CSV",
				UsageText: "ownershit import-csv [flags] owner1/repo1 [owner2/repo2] ...",
				Before:    configureImportClient,
				Action:    importCSVCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "output CSV file path (default: stdout)",
					},
					&cli.BoolFlag{
						Name:    "append",
						Aliases: []string{"a"},
						Usage:   "append to existing CSV file instead of overwriting",
					},
					&cli.StringFlag{
						Name:    "batch-file",
						Aliases: []string{"f"},
						Usage:   "read repository list from file (one owner/repo per line)",
					},
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						EnvVars: []string{"OWNERSHIT_DEBUG"},
						Usage:   "enable debug logging",
						Value:   false,
					},
				},
			},
			{
				Name:   "permissions",
				Usage:  "Show required GitHub token permissions for ownershit operations",
				Action: permissionsCommand,
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

func configureImportClient(c *cli.Context) error {
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

	// Perform schema migration if needed
	if err := shit.MigrateConfigurationSchema(settings); err != nil {
		log.Err(err).
			Str("configPath", configPath).
			Str("operation", "migrateConfigSchema").
			Msg("configuration schema migration failed")
		return shit.NewConfigFileError(configPath, "migrate", "failed to migrate configuration schema", err)
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

// importCommand imports a repository's ownershit configuration from GitHub and writes it as YAML.
//
// importCommand expects a single CLI argument in the form `owner/repo`. It fetches the repository
// configuration via the GitHub client, marshals it to YAML, and either prints the YAML to stdout
// or writes it to the path specified by the `--output` flag (file created with mode 0600).
// It returns an error if the argument is missing or malformed, if the import or YAML marshaling
// fails, or if writing the output file fails.
func importCommand(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("expected exactly one argument: owner/repo")
	}

	repoPath := c.Args().Get(0)
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return fmt.Errorf("repository path must be in format owner/repo, got: %s", repoPath)
	}

	owner := parts[0]
	repo := parts[1]

	log.Info().
		Str("owner", owner).
		Str("repo", repo).
		Msg("importing repository configuration")

	config, err := shit.ImportRepositoryConfig(owner, repo, githubClient)
	if err != nil {
		return fmt.Errorf("failed to import repository configuration: %w", err)
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to YAML: %w", err)
	}

	outputPath := c.String("output")
	log.Debug().Str("outputPath", outputPath).Msg("checking output path")
	if outputPath != "" {
		log.Debug().Str("file", outputPath).Msg("writing to file")
		err = os.WriteFile(outputPath, yamlData, 0o600)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		log.Info().Str("file", outputPath).Msg("configuration exported")
	} else {
		log.Debug().Msg("writing to stdout")
		fmt.Print(string(yamlData))
	}

	return nil
}

// importCSVCommand parses a list of repositories and writes their configuration as CSV to stdout or a file.
//
// It accepts repository arguments in `owner/repo` form and/or a batch file path (flag `--batch-file`), and requires at least one repository.
// If `--output` is provided the command writes to that file; `--append` will open the file in append mode after validating header compatibility.
// The CSV generation is delegated to shit.ProcessRepositoriesCSV; if that returns a *shit.BatchProcessingError the command logs per-item errors and
// returns nil when at least one repository was processed successfully (partial success), otherwise an error is returned.
//
// Returns an error when repository parsing, output file opening, append-mode validation, or the CSV processing (with no successes) fails.
func importCSVCommand(c *cli.Context) error {
	// Parse repository list from args and/or batch file
	repos, err := shit.ParseRepositoryList(c.Args().Slice(), c.String("batch-file"))
	if err != nil {
		return fmt.Errorf("failed to parse repository list: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories specified. Use 'owner/repo' format or --batch-file")
	}

	// Setup output destination
	outputPath := c.String("output")
	appendMode := c.Bool("append")

	var output *os.File
	var shouldClose bool

	if outputPath != "" {
		// File output
		flag := os.O_CREATE | os.O_WRONLY
		if appendMode {
			flag |= os.O_APPEND
			// Validate header compatibility in append mode
			if err := shit.ValidateCSVAppendMode(outputPath); err != nil {
				return fmt.Errorf("append mode validation failed: %w", err)
			}
		} else {
			flag |= os.O_TRUNC
		}

		output, err = os.OpenFile(outputPath, flag, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		shouldClose = true
	} else {
		// Stdout output
		output = os.Stdout
	}

	if shouldClose {
		defer output.Close()
	}

	// Process repositories and generate CSV
	err = shit.ProcessRepositoriesCSV(repos, output, githubClient, !appendMode)
	if err != nil {
		// Check if it's a batch processing error with partial success
		if batchErr, ok := err.(*shit.BatchProcessingError); ok {
			log.Warn().
				Int("successful", batchErr.SuccessCount).
				Int("failed", batchErr.ErrorCount).
				Msg("CSV import completed with some failures")

			// Log detailed errors
			for _, errDetail := range batchErr.GetDetailedErrors() {
				log.Error().Msg(errDetail)
			}

			// Return success if we had any successful imports
			if batchErr.SuccessCount > 0 {
				return nil
			}
		}
		return fmt.Errorf("CSV import failed: %w", err)
	}

	if outputPath != "" {
		log.Info().
			Str("file", outputPath).
			Int("repositories", len(repos)).
			Msg("CSV export completed successfully")
	}

	return nil
}

// initCommand creates a stub configuration file at the path specified by the CLI `--config` flag.
// If a file already exists at that path it does nothing and informs the user. Otherwise it writes
// a template configuration (0600 permissions), prints next-step guidance to stdout, and returns
// any error encountered while writing the file.
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

	err := os.WriteFile(configPath, []byte(stubConfig), 0o600)
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

// getStubConfig returns a starter YAML configuration used to create a new ownershit config file.
// The returned string contains a complete example configuration (organization placeholder,
// branch protection defaults, team permissions, repository entries, and default labels) intended
// to be written as the initial config file and edited by the user.
func getStubConfig() string {
	return `# ownershit configuration file
# This file defines repository ownership, team permissions, and branch protection rules

# Configuration schema version
version: "1.0"

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

func permissionsCommand(c *cli.Context) error {
	fmt.Println("GitHub Token Permissions Required for ownershit")
	fmt.Println("=" + strings.Repeat("=", 48))
	fmt.Println()

	permissions := shit.GetRequiredTokenPermissions()

	fmt.Println("CLASSIC PERSONAL ACCESS TOKEN SCOPES:")
	fmt.Println("-------------------------------------")
	for _, scope := range permissions["classic_token_scopes"] {
		fmt.Printf("✓ %s\n", scope)
	}
	fmt.Println()

	fmt.Println("FINE-GRAINED PERSONAL ACCESS TOKEN PERMISSIONS:")
	fmt.Println("-----------------------------------------------")
	for _, perm := range permissions["fine_grained_permissions"] {
		if perm == "" {
			fmt.Println()
		} else {
			fmt.Println(perm)
		}
	}
	fmt.Println()

	fmt.Println("PERMISSION REQUIREMENTS BY OPERATION:")
	fmt.Println("------------------------------------")
	for _, op := range permissions["operations_requiring_permissions"] {
		fmt.Printf("• %s\n", op)
	}
	fmt.Println()

	fmt.Println("SETUP INSTRUCTIONS:")
	fmt.Println("------------------")
	fmt.Println("1. Create a token at: https://github.com/settings/tokens")
	fmt.Println("2. For classic tokens: Select the scopes listed above")
	fmt.Println("3. For fine-grained tokens: Configure repository and organization permissions as shown")
	fmt.Println("4. Set your token: export GITHUB_TOKEN=your_token_here")
	fmt.Println("5. Verify with: ownershit ratelimit")

	return nil
}
