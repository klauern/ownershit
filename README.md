# Ownershit

A comprehensive CLI tool for managing GitHub repository ownership, permissions, and branch protection rules across your organization.

## Features

- **Team Permissions**: Manage admin/push/pull access control for teams across repositories
- **Branch Protection**: Advanced branch protection rules with status checks, admin enforcement, and push restrictions
- **Repository Settings**: Configure wiki, issues, and projects settings
- **Label Management**: Sync default labels across repositories with emoji support
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

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Create a stub configuration file | `ownershit init` |
| `sync` | Synchronize all repository settings | `ownershit sync --config repositories.yaml` |
| `branches` | Update branch merge strategies | `ownershit branches` |
| `label` | Sync default labels across repositories | `ownershit label` |
| `ratelimit` | Check GitHub API rate limits | `ownershit ratelimit` |

### Archive Commands

| Command | Description | Example |
|---------|-------------|---------|
| `archive query` | Find repositories eligible for archiving | `ownershit archive query --username myuser --days 365` |
| `archive execute` | Archive selected repositories interactively | `ownershit archive execute --username myuser --stars 0` |

### Global Flags

| Flag | Description | Default | Environment Variable |
|------|-------------|---------|---------------------|
| `--config` | Configuration file path | `repositories.yaml` | - |
| `--debug, -d` | Enable debug logging | `false` | `OWNERSHIT_DEBUG` |

## Configuration

### Basic Configuration

Create a `repositories.yaml` file with your organization settings:

```yaml
# Your GitHub organization name (REQUIRED)
organization: your-org-name

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
    wiki: true
    issues: true
    projects: false
  - name: internal-tool
    wiki: false
    issues: true
    projects: true
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

## Architecture

### Core Components

- **GitHub REST v3 API**: Team permissions, issue labels, merge strategies
- **GitHub GraphQL v4 API**: Repository settings, branch protection, archiving
- **Dual API Approach**: Leverages strengths of both APIs for comprehensive coverage

### Configuration Structure

```
repositories.yaml
├── organization: string          # GitHub organization name
├── branches: BranchPermissions   # Branch protection rules
├── team: []TeamPermission        # Team access levels
├── repositories: []Repository    # Repository configurations
└── default_labels: []Label       # Default labels for all repos
```

### Key Files

- `cmd/ownershit/main.go` - CLI entry point and commands
- `config.go` - Configuration parsing and validation
- `github_v3.go` - REST API client implementation
- `github_v4.go` - GraphQL API client implementation
- `branch.go` - Branch protection management
- `archiving_v4.go` - Repository archiving functionality

## GitHub Token Setup

1. Create a Personal Access Token at: https://github.com/settings/tokens

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
