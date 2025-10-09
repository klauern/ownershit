# Scripts

This directory contains utility scripts for working with ownershit.

## generate-config.sh

Generates an ownershit configuration file from all repositories owned by a GitHub user.

### Usage

```bash
./scripts/generate-config.sh <github-username> [output-file]
```

### Examples

```bash
# Generate configuration for user 'klauern'
./scripts/generate-config.sh klauern

# Specify custom output file
./scripts/generate-config.sh klauern my-config.yaml
```

### Prerequisites

- **GITHUB_TOKEN**: Set your GitHub Personal Access Token as an environment variable
  ```bash
  export GITHUB_TOKEN=ghp_your_token_here
  ```

- **Go 1.21+**: Required to run the script

### What it does

1. Fetches all repositories you own from GitHub (not forked repos)
2. Analyzes repository settings to determine sensible defaults
3. Generates a YAML configuration file compatible with ownershit
4. Includes all repository metadata (description, homepage, private/archived status, etc.)

### Configuration Structure

The generated configuration includes:

- **version**: Schema version (1.0)
- **organization**: Your GitHub username
- **defaults**: Calculated defaults based on majority settings across your repos
  - `wiki`: Enable/disable wikis by default
  - `issues`: Enable/disable issues by default
  - `projects`: Enable/disable projects by default
  - `delete_branch_on_merge`: Auto-delete branches after merge
- **branches**: Basic branch protection settings (you should customize these)
- **team**: Empty array (add your team permissions)
- **repositories**: List of all your repositories with their specific settings
- **default_labels**: Empty array (add labels to sync across repos)
- **default_topics**: Empty array (add topics to apply to all repos)

### Next Steps After Generation

1. **Review the generated file**: Check that all repositories are listed correctly
2. **Add team permissions**: If working with an organization, add team access levels
3. **Configure branch protection**: Customize the `branches:` section with your security requirements
4. **Add labels**: Define standard labels in `default_labels:` to sync across repositories
5. **Add topics**: Define standard topics in `default_topics:` to categorize your repositories
6. **Test with dry-run**: 
   ```bash
   ownershit sync --config klauern-repositories.yaml --dry-run
   ```
7. **Apply changes**:
   ```bash
   ownershit sync --config klauern-repositories.yaml
   ```

## generate-user-config.go

The underlying Go program that does the actual work. You can also run it directly:

```bash
go run scripts/generate-user-config.go klauern klauern-repositories.yaml
```

## backfill-repo-features.py

Python script for backfilling repository features (pre-existing, not related to this feature).

## run-e2e-tests.sh

Runs end-to-end integration tests (pre-existing, not related to this feature).
