# End-to-End Testing Guide

This document describes how to run end-to-end (E2E) tests for the ownershit CLI tool against real GitHub repositories.

## Overview

The E2E tests validate the complete functionality of ownershit by making actual API calls to GitHub. These tests are designed to be safe and minimally invasive, with most tests being read-only operations.

## Prerequisites

### Required

- **GitHub Token**: A valid GitHub personal access token or GitHub App token
- **Go**: Go 1.19+ installed
- **Repository Access**: Access to a GitHub organization and repository for testing

### Recommended

- **GitHub CLI**: For easier authentication and user detection
- **Test Organization**: A dedicated test organization to avoid affecting production repositories

## Quick Start

### 1. Set up your GitHub token

```bash
export GITHUB_TOKEN=your_github_token_here
```

### 2. Configure test targets (optional)

```bash
export E2E_TEST_ORG=your-test-org
export E2E_TEST_REPO=your-test-repo
export E2E_TEST_USER=your-github-username
```

### 3. Run read-only tests

```bash
./scripts/run-e2e-tests.sh --readonly
```

### 4. Run all tests (including modifications)

```bash
./scripts/run-e2e-tests.sh --modifications
```

## Test Categories

### Read-Only Tests

These tests only read data from GitHub and make no modifications:

- **Rate Limit**: Tests GitHub API rate limiting
- **Token Validation**: Validates GitHub token format and access
- **Configuration Validation**: Tests YAML configuration parsing
- **CLI Configuration**: Tests command-line argument handling
- **Team Access**: Attempts to read organization teams (may fail without access)

### Modification Tests

These tests may make changes to repositories (requires explicit consent):

- **Repository Sync**: Tests repository settings synchronization
- **Branch Protection**: Tests branch protection rule application
- **Permission Updates**: Tests team permission modifications

## Running Tests

### Using the Test Script

The `scripts/run-e2e-tests.sh` script provides a convenient way to run E2E tests:

```bash
# Read-only tests only
./scripts/run-e2e-tests.sh --readonly

# All tests with custom organization
./scripts/run-e2e-tests.sh --org my-test-org --repo my-test-repo

# Skip building (if already built)
./scripts/run-e2e-tests.sh --skip-build

# Show help
./scripts/run-e2e-tests.sh --help
```

### Manual Test Execution

You can also run tests manually with Go:

```bash
# Enable E2E tests
export RUN_E2E_TESTS=true
export GITHUB_TOKEN=your_token

# Run specific test patterns
go test -v -run "TestE2E_RateLimit" ./...
go test -v -run "TestE2E_CLI" ./cmd/ownershit/...

# Run all E2E tests
go test -v -run "TestE2E" ./...
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GITHUB_TOKEN` | GitHub API token | - | ✅ |
| `RUN_E2E_TESTS` | Enable E2E tests | `false` | ✅ |
| `E2E_TEST_ORG` | Test organization | `test-org` | ❌ |
| `E2E_TEST_REPO` | Test repository | `test-repo` | ❌ |
| `E2E_TEST_USER` | GitHub username | auto-detect | ❌ |
| `E2E_ALLOW_MODIFICATIONS` | Allow modification tests | `false` | ❌ |

## Safety Features

### Skipping by Default

E2E tests are skipped by default unless explicitly enabled with `RUN_E2E_TESTS=true`.

### Read-Only First

Most tests are read-only operations that don't modify any GitHub resources.

### Explicit Consent

Modification tests require explicit consent via:

- Setting `E2E_ALLOW_MODIFICATIONS=true`, or
- Confirming when prompted by the test script

### Error Handling

Tests are designed to gracefully handle common error conditions:

- Missing repository access
- Insufficient permissions
- Non-existent organizations/repositories

## Test Configuration

### Basic Test Config

```yaml
organization: test-org
repositories:
  - name: test-repo
    description: "Test repository for E2E testing"
    wiki: true
    issues: true
    projects: false
team:
  - name: developers
    level: push
branches:
  require_code_owners: false
  require_approving_count: 1
  require_pull_request_reviews: true
```

### Testing Different Scenarios

You can create custom configuration files to test different scenarios:

```bash
# Test with custom config
export E2E_CONFIG_FILE=/path/to/custom-config.yaml
go test -v -run "TestE2E_CLI_ConfigValidation" ./cmd/ownershit/...
```

## Troubleshooting

### Common Issues

**"Skipping E2E tests"**

- Ensure `RUN_E2E_TESTS=true` is set
- Verify `GITHUB_TOKEN` is set and valid

**"Failed to get teams for org"**

- Organization may not exist or you may not have access
- This is expected and won't fail the tests

**"Repository operations failed"**

- Repository may not exist or you may not have write access
- Tests will log warnings but continue

**Token validation failed**

- Check your token format (should start with `ghp_`, `github_pat_`, or `ghs_`)
- Ensure token has appropriate scopes

### Debug Mode

Enable debug logging for more detailed output:

```bash
export OWNERSHIT_DEBUG=true
go test -v -run "TestE2E" ./...
```

## CI/CD Integration

For automated testing in CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run E2E Tests
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    RUN_E2E_TESTS: true
    E2E_TEST_ORG: my-test-org
  run: |
    ./scripts/run-e2e-tests.sh --readonly
```

## Best Practices

1. **Use Test Organizations**: Always use dedicated test organizations/repositories
1. **Start with Read-Only**: Run read-only tests first to validate setup
1. **Review Modifications**: Carefully review what modification tests will do
1. **Token Permissions**: Use tokens with minimal required permissions
1. **Regular Testing**: Run E2E tests as part of your development workflow

## Contributing

When adding new E2E tests:

1. Make them read-only when possible
1. Add appropriate skip conditions
1. Handle error cases gracefully
1. Document any special requirements
1. Test both success and failure scenarios
