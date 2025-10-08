# Quick Start: Generate Configuration from Your Repositories

This guide shows you how to automatically generate an `ownershit` configuration file from all of your existing GitHub repositories.

## Overview

Instead of manually creating a configuration file, you can use the provided script to:
1. Fetch all repositories you own from GitHub
2. Analyze their settings to determine sensible defaults
3. Generate a complete YAML configuration file ready for use

## Prerequisites

1. **GitHub Personal Access Token**
   - Create a token at: https://github.com/settings/tokens
   - Required scopes:
     - `repo` - Full repository access
     - `read:org` - Read organization data (if applicable)

2. **Go 1.21+** installed on your system

## Step-by-Step Guide

### 1. Set Your GitHub Token

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

### 2. Run the Generation Script

```bash
# Generate configuration for your username
./scripts/generate-config.sh klauern

# This creates: klauern-repositories.yaml
```

Or specify a custom output filename:

```bash
./scripts/generate-config.sh klauern my-config.yaml
```

### 3. Review the Generated File

The script creates a configuration file with:

```yaml
version: "1.0"
organization: klauern

# Defaults calculated from your repositories
defaults:
  wiki: false              # Based on majority of repos
  issues: true             # Based on majority of repos
  projects: false          # Based on majority of repos
  delete_branch_on_merge: true

# Basic branch protection (customize this!)
branches:
  require_pull_request_reviews: false
  require_approving_count: 0
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: true

# Empty - add your teams here
team: []

# All your repositories with their specific settings
repositories:
  - name: repo-1
    description: "Repository description"
    private: true
    # Only includes settings that differ from defaults
    
  - name: repo-2
    archived: true
    # etc...

# Empty - add your standard labels
default_labels: []

# Empty - add your standard topics
default_topics: []
```

### 4. Customize the Configuration

#### Add Team Permissions (if using organizations)

```yaml
team:
  - name: developers
    level: push
  - name: maintainers
    level: admin
  - name: contractors
    level: pull
```

#### Configure Branch Protection

```yaml
branches:
  require_pull_request_reviews: true
  require_approving_count: 2
  require_code_owners: true
  
  # Status checks
  require_status_checks: true
  status_checks:
    - "ci/build"
    - "ci/test"
    - "ci/lint"
  require_up_to_date_branch: true
  
  # Advanced settings
  enforce_admins: true
  require_linear_history: true
  allow_force_pushes: false
```

#### Add Default Labels

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
```

#### Add Default Topics

```yaml
default_topics:
  - "golang"
  - "cli"
  - "automation"
```

### 5. Test with Dry Run

**Always test first!** Dry run shows what would happen without making changes:

```bash
ownershit sync --config klauern-repositories.yaml --dry-run
```

Review the output to ensure it matches your expectations.

### 6. Apply the Configuration

Once you're satisfied with the dry run output:

```bash
ownershit sync --config klauern-repositories.yaml
```

## Understanding the Generation Logic

### Defaults Calculation

The script analyzes all your repositories and sets defaults based on majority (>50%):

- If more than half your repos have wikis enabled → `defaults.wiki: true`
- If more than half have issues enabled → `defaults.issues: true`
- Same logic for projects and delete_branch_on_merge

### Repository-Specific Settings

Repositories only include settings that differ from the calculated defaults. For example:

```yaml
defaults:
  wiki: false
  issues: true

repositories:
  - name: repo-with-defaults
    # Inherits: wiki=false, issues=true
    
  - name: repo-with-custom-settings
    wiki: true     # Override: different from default
    issues: false  # Override: different from default
```

This keeps the configuration clean and focused on differences.

### Metadata Included

The script captures important repository metadata:
- `description` - Repository description
- `homepage` - Project homepage URL
- `private` - Whether repo is private (only if true)
- `archived` - Whether repo is archived (only if true)
- `template` - Whether repo is a template (only if true)
- `default_branch` - Default branch (only if not "main")
- `discussions_enabled` - GitHub discussions enabled (only if true)

## Advanced Usage

### Direct Go Script Execution

You can also run the Go script directly:

```bash
go run scripts/generate-user-config.go klauern klauern-repositories.yaml
```

### Comparing with Example

Compare your generated config with the example:

```bash
diff klauern-repositories.yaml example-repositories.yaml
```

### Updating Configuration

If you add new repositories on GitHub, regenerate and merge:

1. Generate fresh config: `./scripts/generate-config.sh klauern new-config.yaml`
2. Review differences: `diff klauern-repositories.yaml new-config.yaml`
3. Merge new repositories into your existing config
4. Test with dry run
5. Apply changes

## Troubleshooting

### "GITHUB_TOKEN environment variable not set"

Set your token:
```bash
export GITHUB_TOKEN=ghp_your_token_here
```

### "failed to fetch repositories: 401"

Your token is invalid or expired. Create a new token at https://github.com/settings/tokens

### "failed to fetch repositories: 403"

Your token lacks required scopes. Ensure it has:
- `repo` - Full repository access
- `read:org` - Read organization data

### Empty repositories list

The script only fetches repositories you **own**, not:
- Forked repositories
- Repositories you have access to but don't own
- Organization repositories (use organization name instead of username)

## Next Steps

1. ✅ Generate configuration
2. ✅ Review and customize
3. ✅ Test with dry run
4. ✅ Apply changes
5. Consider using the generated config as a starting point for:
   - Setting up CI/CD workflows
   - Enforcing security policies
   - Standardizing across repositories
   - Onboarding new repositories

## More Information

- [Main README](README.md) - Full documentation
- [Scripts README](scripts/README.md) - Detailed script documentation
- [Example Configuration](example-repositories.yaml) - Full example with all features

## Questions?

- Check existing issues: https://github.com/klauern/ownershit/issues
- Open a new issue for bugs or feature requests
