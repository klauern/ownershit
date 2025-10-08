# Ownershit

A comprehensive CLI tool for managing GitHub repository ownership, permissions, and branch protection rules across your organization.

## Features

- **Team Permissions**: Manage admin/push/pull access control for teams across repositories
- **Branch Protection**: Advanced branch protection rules with status checks, admin enforcement, and push restrictions
- **Repository Settings**: Configure wiki, issues, and projects settings
- **Label Management**: Sync default labels across repositories with emoji support
- **Topic Management**: Mass-assign repository topics/tags additively or by replacement
- **Repository Archiving**: Find and archive inactive repositories based on configurable criteria
- **Merge Strategy Control**: Configure allowed merge types (merge commits, squash, rebase)

## Installation

### Using Go Install

```bash
go install github.com/klauern/ownershit/cmd/ownershit@latest
```

### Using Task (Development)

```bash
# Clone the repository
git clone https://github.com/klauern/ownershit.git
cd ownershit

# Set up development environment
task dev

# Build and install locally
task install
```

## Quick Start

1. Create a configuration file:

   ```bash
   ownershit init
   ```

1. Edit `repositories.yaml` with your organization details

1. Set up your GitHub token:

   ```bash
   export GITHUB_TOKEN=your_github_token_here
   ```

1. Synchronize your repositories:

   ```bash
   ownershit sync
   ```

## Commands

### Core Commands

| Command       | Description                             | Example                                            |
| ------------- | --------------------------------------- | -------------------------------------------------- |
| `init`        | Create a stub configuration file          | `ownershit init`                                   |
| `sync`        | Synchronize all repository settings     | `ownershit sync --config repositories.yaml`         |
| `branches`    | Update branch merge strategies          | `ownershit branches`                               |
| `label`       | Sync default labels across repositories | `ownershit label`                                  |
| `topics`      | Sync repository topics/tags             | `ownershit topics --additive=true`                 |
| `import`      | Import repository configuration as YAML  | `ownershit import owner/repo --output config.yaml`  |
| `permissions` | Show required GitHub token permissions  | `ownershit permissions`                            |
| `ratelimit`   | Check GitHub API rate limits            | `ownershit ratelimit`                              |

### Archive Commands

| Command           | Description                                 | Example                                                 |
| ----------------- | ------------------------------------------- | ------------------------------------------------------- |
| `archive query`   | Find repositories eligible for archiving    | `ownershit archive query --username myuser --days 365`  |
| `archive execute` | Archive selected repositories interactively | `ownershit archive execute --username myuser --stars 0` |

### Import/Export Commands

| Command      | Description                                    | Example                                                           |
| ------------ | ---------------------------------------------- | ----------------------------------------------------------------- |
| `import-csv` | Import multiple repositories and export as CSV | `ownershit import-csv owner/repo1 owner/repo2 --output repos.csv` |

### Global Flags

| Flag          | Description             | Default             | Environment Variable |
| ------------- | ----------------------- | ------------------- | -------------------- |
| `--config`    | Configuration file path | `repositories.yaml` | -                    |
| `--debug, -d` | Enable debug logging    | `false`             | `OWNERSHIT_DEBUG`    |

## Configuration

### Basic Configuration

Create a `repositories.yaml` file with your organization settings:

```yaml
# Your GitHub organization name (REQUIRED)
organization: your-org-name

# Global defaults for repository features (optional)
# These apply to all repositories unless explicitly overridden
default_wiki: false      # Disable wikis by default
default_issues: true     # Enable issues by default
default_projects: false  # Disable projects by default

# Team permissions for repositories
team:
  - name: developers
    level: push
  - name: maintainers
    level: admin
  - name: security-team
    level: admin

# Repository configurations
repositories:
  - name: my-app
    wiki: true           # Override: enable wiki for this repo
    # issues: inherits default (true)
    # projects: inherits default (false)
  - name: internal-tool
    # wiki: inherits default (false)
    # issues: inherits default (true)
    projects: true       # Override: enable projects for this repo
```

### Advanced Branch Protection

Configure comprehensive branch protection rules:

```yaml
branches:
  # Basic protection
  require_pull_request_reviews: true
  require_approving_count: 2
  require_code_owners: true

  # Merge strategy controls
  allow_merge_commit: false
  allow_squash_merge: true
  allow_rebase_merge: true

  # Status checks
  require_status_checks: true
  status_checks:
    - "ci/build"
    - "ci/test"
    - "security/scan"
  require_up_to_date_branch: true

  # Advanced protection
  enforce_admins: true
  restrict_pushes: true
  push_allowlist:
    - "admin-team"
    - "deploy-team"
  require_conversation_resolution: true
  require_linear_history: true
  allow_force_pushes: false
  allow_deletions: false
```

### Label Management

Define default labels for all repositories:

```yaml
default_labels:
  - name: "bug"
    color: "d73a4a"
    emoji: "🐛"
    description: "Something isn't working"
  - name: "enhancement"
    color: "a2eeef"
    emoji: "✨"
    description: "New feature or request"
  - name: "security"
    color: "ff6b6b"
    emoji: "🔒"
    description: "Security-related issue"
```

### Topic Management

Define default topics to apply across repositories:

```yaml
default_topics:
  - "golang"
  - "cli"
  - "github-management"
  - "internal-tool"
```

### Repository Feature Defaults

Configure global defaults for repository features. These defaults apply to all repositories unless explicitly overridden at the repository level.

#### Available Default Settings

- `default_wiki` - Enable/disable wikis for all repositories
- `default_issues` - Enable/disable issue tracking for all repositories
- `default_projects` - Enable/disable GitHub projects for all repositories

#### How It Works

1. **Global Defaults**: Set default values at the top level of your configuration
2. **Per-Repository Overrides**: Specify values at the repository level to override defaults
3. **Inheritance**: If a repository doesn't specify a value, it inherits from the default
4. **Nil Behavior**: If no default is set and no repository value is provided, no change is made

#### Example Configuration

```yaml
organization: my-org

# Set defaults - most repos don't need wikis or projects
default_wiki: false
default_issues: true
default_projects: false

repositories:
  # Inherits all defaults (wiki: false, issues: true, projects: false)
  - name: simple-tool

  # Override wiki only - enable for documentation
  - name: main-app
    wiki: true
    # issues: inherits default (true)
    # projects: inherits default (false)

  # Override projects only - needs project board
  - name: team-planning
    # wiki: inherits default (false)
    # issues: inherits default (true)
    projects: true

  # Override all defaults
  - name: special-repo
    wiki: true
    issues: false
    projects: true
```

#### Migration from Explicit Settings

If you have an existing configuration with explicit settings on every repository, you can migrate to using defaults:

**Before** (explicit settings everywhere):
```yaml
repositories:
  - name: repo1
    wiki: false
    issues: true
    projects: false
  - name: repo2
    wiki: false
    issues: true
    projects: false
  - name: repo3
    wiki: false
    issues: true
    projects: false
```

**After** (using defaults):
```yaml
default_wiki: false
default_issues: true
default_projects: false

repositories:
  - name: repo1
  - name: repo2
  - name: repo3
```

This feature is fully backward compatible - existing configurations continue to work unchanged.

## Development

This project uses [Task](https://taskfile.dev) for task management.

### Development Setup

```bash
# Set up development environment
task dev

# Install development dependencies
# - go mod tidy
# - install mockgen
# - install govulncheck
# - install bump tool
```

### Building and Testing

```bash
# Build binaries
task build

# Run tests with coverage
task test

# View test coverage in browser
task test-cover

# Run tests with security checks
task test-all
```

### Code Quality

```bash
# Format code
task fmt

# Run linter
task lint

# Generate mocks for testing
task mocks

# Run security vulnerability check
task security
```

### GraphQL Client Management

```bash
# Download latest GitHub GraphQL schema
task gql:download-schema

# Generate GraphQL client code
task gql:generate-client
```

### Releases

```bash
# Show release options
task release

# Create patch release (v0.6.0 → v0.6.1)
task release:patch

# Create minor release (v0.6.0 → v0.7.0)
task release:minor

# Create major release (v0.6.0 → v1.0.0)
task release:major

# Create release candidate (v0.6.0 → v0.7.0-rc1)
task release:rc
```

## Utilities

### Backfill Repository Features

The `scripts/backfill-repo-features.py` script helps migrate existing configurations by detecting actual feature usage and updating your YAML file:

```bash
# Requires uv or Python 3.11+ with PyGithub and PyYAML
export GITHUB_TOKEN=your_token
uv run scripts/backfill-repo-features.py repositories.yaml
```

This script:
- Checks actual wiki/issues/projects usage for each repository
- Adds explicit settings where they differ from defaults
- Removes redundant explicit settings that match defaults
- Creates a `.backup` file before making changes

## Architecture

### Core Components

- **GitHub REST v3 API**: Team permissions, issue labels, merge strategies
- **GitHub GraphQL v4 API**: Repository settings, branch protection, archiving
- **Dual API Approach**: Leverages strengths of both APIs for comprehensive coverage

### Configuration Structure

```
repositories.yaml
├── organization: string          # GitHub organization name
├── default_wiki: bool            # Global default (optional)
├── default_issues: bool          # Global default (optional)
├── default_projects: bool        # Global default (optional)
├── branches: BranchPermissions   # Branch protection rules
├── team: []TeamPermission        # Team access levels
├── repositories: []Repository    # Repository configurations
├── default_labels: []Label       # Default labels for all repos
└── default_topics: []string      # Default topics for all repos
```

### Key Files

- `cmd/ownershit/main.go` - CLI entry point and commands
- `config.go` - Configuration parsing and validation
- `github_v3.go` - REST API client implementation
- `github_v4.go` - GraphQL API client implementation
- `branch.go` - Branch protection management
- `archiving_v4.go` - Repository archiving functionality

## GitHub Token Setup

1. Create a Personal Access Token at: <https://github.com/settings/tokens>

1. Required scopes:

   - `repo` - Full repository access
   - `admin:org` - Organization admin access (for team management)

1. Set the token as an environment variable:

   ```bash
   export GITHUB_TOKEN=your_token_here
   ```

## Examples

### Complete Organization Setup

```bash
# Initialize configuration
ownershit init

# Edit repositories.yaml with your settings

# Apply all configurations
ownershit sync --config repositories.yaml --debug
```

### Archive Inactive Repositories

```bash
# Find repositories inactive for 365+ days with 0 stars
ownershit archive query --username myuser --days 365 --stars 0

# Interactively select and archive repositories
ownershit archive execute --username myuser --days 365 --stars 0
```

### Update Branch Protection Only

```bash
# Apply only branch merge strategy changes
ownershit branches --config repositories.yaml
```

### Sync Labels Across Repositories

```bash
# Update labels on all configured repositories
ownershit label --config repositories.yaml
```

### Sync Topics Across Repositories

```bash
# Additively merge topics (default) - preserves existing topics
ownershit topics --config repositories.yaml

# Replace all topics with configured ones
ownershit topics --config repositories.yaml --additive=false
```

### Import Repository Configuration

```bash
# Import single repository configuration as YAML
ownershit import myorg/myrepo --output repo-config.yaml

# Import repository to stdout and preview settings
ownershit import myorg/myrepo
```

### Bulk Export Repository Data as CSV

```bash
# Export multiple repositories to CSV file
ownershit import-csv myorg/repo1 myorg/repo2 myorg/repo3 --output repositories.csv

# Export from batch file with repository list
ownershit import-csv --batch-file repo-list.txt --output export.csv

# Append to existing CSV file
ownershit import-csv myorg/new-repo --output existing-data.csv --append
```

## Troubleshooting

### Common Issues

**Configuration file not found**

```bash
# Create a new configuration file
ownershit init
```

**GitHub API rate limits**

```bash
# Check current rate limit status
ownershit ratelimit --debug
```

**Permission errors**

```bash
# Verify token has required scopes:
# - repo (full repository access)
# - admin:org (organization admin access)
```

**Debug mode**

```bash
# Enable verbose logging
ownershit sync --debug
# or
export OWNERSHIT_DEBUG=true
ownershit sync
```

### Getting Help

- Run `ownershit --help` for command help
- Run `ownershit <command> --help` for command-specific help
- Check the [example configuration](testdata/enhanced-branch-protection-example.yaml) for advanced features
- Enable debug mode (`--debug`) for detailed logging

## Contributing

1. Fork the repository
1. Create a feature branch (`git checkout -b feature/amazing-feature`)
1. Make your changes
1. Run tests (`task test-all`)
1. Commit your changes (`git commit -m 'Add amazing feature'`)
1. Push to the branch (`git push origin feature/amazing-feature`)
1. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Authors

- **Nick Klauer** - *Initial work* - [klauern](https://github.com/klauern)
