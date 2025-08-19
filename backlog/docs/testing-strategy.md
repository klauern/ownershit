# Testing Strategy

## Overview

The ownershit project employs a comprehensive testing strategy to ensure reliability, security, and maintainability of the GitHub repository management tool.

## Current Test Coverage

### Package-Level Coverage

- **Main Package**: 72.1% coverage
- **CLI Commands (`cmd/`)**: 80.6% coverage (recently improved)
- **GraphQL API (`v4api/`)**: 27.0% coverage (target: 80%+)
- **Overall Target**: 90%+ across all packages

### Test Types

#### Unit Tests

- **Mock-based testing** for all external API interactions
- **Configuration validation** tests for YAML parsing
- **Error handling verification** for all error scenarios
- **Business logic testing** for core functionality

#### Integration Tests

- **Live API testing** with dedicated test repositories
- **End-to-end workflow testing** for complete command flows
- **Configuration compatibility** testing across versions
- **Error scenario testing** for API failures

#### Security Tests

- **Token validation** testing for GitHub authentication
- **Permission boundary** testing for access controls
- **Vulnerability scanning** integration with govulncheck
- **Secret detection** in test configurations

## Testing Infrastructure

### Mock Framework

- **Library**: `go.uber.org/mock` (migrated from deprecated golang/mock)
- **Generation**: Automated via `go generate` commands
- **Coverage**: All external dependencies mocked

### Test Organization

```
project/
├── *_test.go           # Unit tests alongside source
├── mocks/              # Generated mock implementations
├── testdata/           # Test configuration files
└── integration_test.go # Integration test suites
```

### CI/CD Integration

- **Automated testing** on all pull requests
- **Coverage reporting** with threshold enforcement
- **Security scanning** integrated into test pipeline
- **Performance regression** testing for critical paths

## Test Development Guidelines

### Writing Unit Tests

1. **Mock external dependencies** (GitHub APIs, file system)
1. **Test error conditions** alongside happy paths
1. **Validate configuration parsing** for all YAML structures
1. **Assert proper logging** and error messages

### Writing Integration Tests

1. **Use dedicated test repositories** for safe API testing
1. **Clean up test artifacts** after test execution
1. **Test against real GitHub API** for compatibility
1. **Validate end-to-end workflows** including error recovery

### Testing Best Practices

- **Atomic tests**: Each test should test one specific behavior
- **Descriptive names**: Test names should clearly describe the scenario
- **Setup/teardown**: Proper test isolation and cleanup
- **Data-driven tests**: Use table-driven tests for multiple scenarios

## Quality Gates

### Coverage Requirements

- **Minimum**: 80% coverage for new code
- **Target**: 90% coverage across all packages
- **Critical paths**: 100% coverage for security and error handling

### Performance Benchmarks

- **API operations**: \<500ms for single repository operations
- **Bulk operations**: \<2s for typical organization sync
- **Memory usage**: \<100MB for large organization management

### Security Requirements

- **No secrets in tests**: All test tokens must be mock or expired
- **Permission validation**: All access control paths tested
- **Vulnerability scanning**: No high/critical vulnerabilities allowed

## Future Enhancements

### Planned Improvements

1. **Property-based testing** for configuration validation
1. **Chaos engineering** for resilience testing
1. **Performance profiling** integration
1. **Mutation testing** for test quality validation

### Tooling Enhancements

1. **Test data generation** utilities
1. **GitHub API mocking** improvements
1. **Coverage visualization** tools
1. **Automated test maintenance** workflows
