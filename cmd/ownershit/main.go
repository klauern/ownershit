package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// Sentinel errors used for argument validation and CLI input checks.
var (
	ErrExpectedOneArgument     = errors.New("expected exactly one argument: owner/repo")
	ErrInvalidRepoPathFormat   = errors.New("repository path must be in format owner/repo")
	ErrNoRepositoriesSpecified = errors.New("no repositories specified. Use 'owner/repo' format or --batch-file")
	ErrConfigPathIsDirectory   = errors.New("configuration path is a directory")
)

// main is the entry point for the ownershit CLI application.
// It configures logging, constructs the command-line interface with subcommands
// (init, branches, sync, archive, label, ratelimit, import, import-csv, permissions),
// and runs the app, terminating with a fatal error if execution fails.
func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := &cli.App{
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			return nil
		},
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Value: "repositories.yaml",
						Usage: "configuration of repository updates to perform",
					},
				},
			},
			{
				Name:      "sync",
				Usage:     "Synchronize branch, repo, owner and other configs on repositories",
				UsageText: "ownershit sync --config repositories.yaml [--dry-run]",
				Before:    configureClient,
				Action:    syncCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Value: "repositories.yaml",
						Usage: "configuration of repository updates to perform",
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "preview changes without applying them",
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
				},
			},
			{
				Name:   "topics",
				Usage:  "set repository topics/tags",
				Before: configureClient,
				Action: topicsCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Value: "repositories.yaml",
						Usage: "configuration of repository updates to perform",
					},
					&cli.BoolFlag{
						Name:  "additive",
						Value: true,
						Usage: "add topics to existing ones (false replaces all topics)",
					},
				},
			},
			{
				Name:   "ratelimit",
				Usage:  "get ratelimit information for the GitHub GraphQL v4 API",
				Before: configureImportClient,
				Action: rateLimitCommand,
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

// configureClient sets up the GitHub client and configuration for CLI commands.
func configureClient(c *cli.Context) error {
	if err := readConfig(c); err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	return newGitHubClient(c)
}

// configureImportClient sets up the GitHub client specifically for import operations.
func configureImportClient(c *cli.Context) error { return newGitHubClient(c) }

// newGitHubClient sets debug level and initializes githubClient; shared by configure* helpers.
func newGitHubClient(c *cli.Context) error {
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

// readConfig reads and validates the configuration file specified in the CLI context.
func readConfig(c *cli.Context) error {
	configPath := c.String("config")
	file, err := os.ReadFile(configPath) // #nosec G304 - config path validated by CLI
	if err != nil {
		if os.IsNotExist(err) {
			log.Err(err).
				Str("configPath", configPath).
				Str("operation", "readConfigFile").
				Msg("config file not found")
			fmt.Fprintf(os.Stderr, "❌ Configuration file '%s' not found.\n\n", configPath)
			fmt.Fprintln(os.Stderr, "To get started, run:")
			fmt.Fprintln(os.Stderr, "  ownershit init")
			fmt.Fprintln(os.Stderr, "This will create a stub configuration file with examples and comments.")
			return shit.NewConfigFileError(configPath, "read", "configuration file not found - run 'ownershit init' to create one", err)
		}
		log.Err(err).
			Str("configPath", configPath).
			Str("operation", "readConfigFile").
			Msg("config file error")
		return shit.NewConfigFileError(configPath, "read", "failed to read configuration file", err)
	}

	settings = &shit.PermissionsSettings{}
	dec := yaml.NewDecoder(bytes.NewReader(file))
	dec.KnownFields(true)
	if err := dec.Decode(settings); err != nil {
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

// syncCommand synchronizes repository settings based on the configuration file.
func syncCommand(c *cli.Context) error {
	dryRun := c.Bool("dry-run")
	if dryRun {
		log.Info().Msg("DRY RUN MODE - No changes will be applied")
	}
	log.Info().Msg("mapping all permissions for repositories")
	shit.MapPermissions(settings, githubClient, dryRun)
	return nil
}

// branchCommand updates branch merge strategies for repositories.
func branchCommand(c *cli.Context) error {
	log.Info().Msg("performing branch updates on repositories")
	shit.UpdateBranchMergeStrategies(settings, githubClient)
	return nil
}

// labelCommand synchronizes labels across repositories.
func labelCommand(c *cli.Context) error {
	log.Info().Msg("synchronizing labels on repositories")
	shit.SyncLabels(settings, githubClient)
	return nil
}

// topicsCommand synchronizes topics across repositories.
func topicsCommand(c *cli.Context) error {
	additive := c.Bool("additive")
	log.Info().
		Bool("additive", additive).
		Msg("synchronizing topics on repositories")
	shit.SyncTopics(settings, githubClient, additive)
	return nil
}

// rateLimitCommand displays GitHub API rate limit information.
func rateLimitCommand(c *cli.Context) error {
	log.Info().Msg("getting ratelimit information")
	rl, err := githubClient.GetRateLimit()
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to retrieve GraphQL rate limit; verify network connectivity and GITHUB_TOKEN permissions")
		return fmt.Errorf("get ratelimit: %w", err)
	}
	log.Info().
		Str("login", rl.Login).
		Int("limit", rl.Limit).
		Int("cost", rl.Cost).
		Int("remaining", rl.Remaining).
		Time("resetAt", rl.ResetAt).
		Msg("rate limit")
	return nil
}

// importCommand imports a repository's ownershit configuration from GitHub and writes it as YAML.
// It expects a single CLI argument in the form `owner/repo`. It fetches the repository
// configuration via the GitHub client, marshals it to YAML, and either prints the YAML to stdout
// or writes it to the path specified by the `--output` flag (file created with mode 0600).
// It returns an error if the argument is missing or malformed, if the import or YAML marshaling
// fails, or if writing the output file fails.
func importCommand(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("%w (got %d arguments)", ErrExpectedOneArgument, c.NArg())
	}

	repoPath := c.Args().Get(0)
	owner, repo, ok := strings.Cut(repoPath, "/")
	if !ok {
		return fmt.Errorf("%w: %s", ErrInvalidRepoPathFormat, repoPath)
	}
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return fmt.Errorf("%w: %s", ErrInvalidRepoPathFormat, repoPath)
	}

	log.Info().
		Str("owner", owner).
		Str("repo", repo).
		Msg("importing repository configuration")

	// Strict error handling for team permissions during YAML import
	config, err := shit.ImportRepositoryConfig(owner, repo, githubClient, false)
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
		// Ensure directory exists for output file
		if dir := filepath.Dir(outputPath); dir != "." {
			if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
				return fmt.Errorf("failed to create output directory %q: %w", dir, mkErr)
			}
		}
		log.Debug().Str("file", outputPath).Msg("writing to file")
		err = os.WriteFile(outputPath, yamlData, 0o600)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		log.Info().Str("file", outputPath).Msg("configuration exported")
	} else {
		log.Debug().Msg("writing to stdout")
		if _, werr := fmt.Fprintln(os.Stdout, string(yamlData)); werr != nil {
			return fmt.Errorf("failed to write YAML to stdout: %w", werr)
		}
	}

	return nil
}

// importCSVCommand parses repos and writes their configuration as CSV to stdout or file.
//
// It accepts repositories in `owner/repo` form and/or a batch file path (`--batch-file`).
// If `--output` is provided, the command writes to that file; `--append` opens the file in
// append mode after validating header compatibility. The CSV generation is delegated to
// shit.ProcessRepositoriesCSV; when it returns a *shit.BatchProcessingError the command logs
// per-item errors and returns success if at least one repository processed successfully.
//
// Returns an error on repository parsing, opening output, append-mode validation, or when
// all repositories fail.
func importCSVCommand(c *cli.Context) error {
	// Parse repository list from args and/or batch file
	repos, err := shit.ParseRepositoryList(c.Args().Slice(), c.String("batch-file"))
	if err != nil {
		return fmt.Errorf("failed to parse repository list: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("%w", ErrNoRepositoriesSpecified)
	}

	// Setup output destination
	outputPath := c.String("output")
	appendMode := c.Bool("append")

	var output *os.File
	var shouldClose bool
	fileHasContent := false
	if outputPath != "" {
		if stat, statErr := os.Stat(outputPath); statErr == nil {
			fileHasContent = stat.Size() > 0
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("failed to stat output file: %w", statErr)
		}
	}

	if outputPath != "" {
		// File output
		flag := os.O_CREATE | os.O_WRONLY
		if appendMode {
			flag |= os.O_APPEND
			// Validate header compatibility only when appending to a non-empty file
			if fileHasContent {
				if vErr := shit.ValidateCSVAppendMode(outputPath); vErr != nil {
					return fmt.Errorf("append mode validation failed: %w", vErr)
				}
			}
		} else {
			flag |= os.O_TRUNC
		}

		output, err = os.OpenFile(outputPath, flag, 0o600) // #nosec G304 - output path validated by CLI
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		shouldClose = true
	} else {
		// Stdout output
		output = os.Stdout
	}

	var closeErr error

	// Determine whether to write the CSV header
	writeHeader := true
	if outputPath == "" {
		// stdout: always write header; append has no meaning here
		if appendMode {
			log.Warn().Msg("--append is ignored when writing to stdout")
		}
	} else {
		// file output
		if appendMode && fileHasContent {
			writeHeader = false
		}
	}

	// Process repositories and generate CSV
	err = shit.ProcessRepositoriesCSV(repos, output, githubClient, writeHeader)

	// Close output immediately if needed
	if shouldClose {
		if closeErr = output.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close output file")
		}
	}

	if err != nil {
		// Check if it's a batch processing error with partial success
		if batchErr, ok := err.(*shit.BatchProcessingError); ok {
			log.Warn().
				Int("successful", batchErr.SuccessCount).
				Int("failed", batchErr.ErrorCount).
				Str("file", outputPath).
				Msg("CSV export completed with some failures")

			// Log detailed errors with structured fields
			for _, repoErr := range batchErr.Errors {
				evt := log.Error()
				if repoErr.Owner != "" {
					evt = evt.Str("owner", repoErr.Owner)
				}
				if repoErr.Repository != "" {
					evt = evt.Str("repo", repoErr.Repository)
				}
				evt.Err(repoErr.Error).Msg("failed to process repository")
			}

			// Return success if we had any successful exports
			if batchErr.SuccessCount > 0 {
				// Check for close error after successful processing
				if closeErr != nil {
					return fmt.Errorf("CSV export succeeded but failed to close output: %w", closeErr)
				}
				return nil
			}
		}
		if closeErr != nil {
			return fmt.Errorf("CSV export failed: %w (also failed to close output: %w)", err, closeErr)
		}
		return fmt.Errorf("CSV export failed: %w", err)
	}

	if outputPath != "" {
		log.Info().
			Str("file", outputPath).
			Int("repositories", len(repos)).
			Msg("CSV export completed successfully")
	}

	// Check for close error after successful processing
	if closeErr != nil {
		return fmt.Errorf("CSV export succeeded but failed to close output: %w", closeErr)
	}
	return nil
}

// initCommand creates a stub configuration file at the path specified by the CLI `--config` flag.
// If a file already exists at that path it does nothing and informs the user. Otherwise it writes
// a template configuration (0600 permissions), prints next-step guidance to stdout, and returns
// any error encountered while writing the file.
func initCommand(c *cli.Context) error {
	configPath := c.String("config")

	// Ensure parent directory exists
	dir := filepath.Dir(configPath)
	if dir != "." {
		if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
			return fmt.Errorf("failed to create config directory %q: %w", dir, mkErr)
		}
	}

	// Check if config file already exists or path is a directory
	if fi, err := os.Stat(configPath); err == nil {
		if fi.IsDir() {
			return fmt.Errorf("%w: %s", ErrConfigPathIsDirectory, configPath)
		}
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

// permissionsCommand displays the required GitHub token permissions.
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
