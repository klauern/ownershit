---
id: 002-configuration-format
title: Decision Record - Configuration Format Choice
type: decision
status: Accepted
date: '2025-07-12'
deciders: Development Team
---

# Decision Record: Configuration Format Choice

**Date**: 2025-07-12  
**Status**: Accepted  
**Deciders**: Development Team

## Context

The ownershit tool requires a configuration format for defining repository settings, team permissions, and organizational policies. We needed to choose a format that balances:

- **Human readability** for manual editing
- **Machine parsability** for programmatic generation
- **Version control friendliness** for GitOps workflows
- **Schema validation** capabilities

## Decision

We will use **YAML** as the primary configuration format for `repositories.yaml` files.

### Configuration Structure

```yaml
# Top-level organization settings
organization: example-org

# Global branch protection rules
branches:
  require_pull_request_reviews: true
  require_approving_count: 2
  enforce_admins: true

# Team permission definitions
team:
  - name: developers
    level: push
  - name: security
    level: admin

# Repository-specific settings
repositories:
  - name: web-app
    wiki: true
    issues: true
    projects: false
  - name: api-service
    wiki: false
    issues: true
    projects: true

# Default labels for all repositories
default_labels:
  - name: "bug"
    color: "d73a4a"
    emoji: "🐛"
    description: "Something isn't working"
```

## Consequences

### Positive

- **Human Readable**: Easy to read and edit manually
- **Git Friendly**: Clean diffs, good merge behavior
- **Comments Support**: Can include documentation within config
- **Ecosystem Support**: Wide tooling support for validation and editing
- **Hierarchical Structure**: Natural representation of nested settings

### Negative

- **Indentation Sensitive**: Whitespace errors can break configuration
- **Type Ambiguity**: Strings vs numbers can be ambiguous
- **Large File Issues**: Can become unwieldy for very large organizations

### Neutral

- **Schema Validation**: Requires additional tooling but widely available
- **Templating**: Can integrate with Go templates or other templating systems

## Implementation Details

### Parsing Strategy

- Use `gopkg.in/yaml.v3` for robust YAML parsing
- Implement comprehensive validation with clear error messages
- Support environment variable substitution for sensitive values

### Validation Approach

```go
type Config struct {
    Organization string `yaml:"organization" validate:"required"`
    Team         []TeamPermission `yaml:"team" validate:"dive"`
    Repositories []Repository `yaml:"repositories" validate:"dive"`
}
```

### Error Handling

- Line number reporting for YAML syntax errors
- Field-level validation with context
- Suggestions for common configuration mistakes

## Alternatives Considered

### JSON

- **Pros**: Strict parsing, excellent tooling support
- **Cons**: No comments, less human-readable, harder manual editing
- **Decision**: Rejected due to poor human readability

### TOML

- **Pros**: Simple syntax, good for flat configurations
- **Cons**: Poor support for complex nested structures
- **Decision**: Rejected due to limited nesting capabilities

### HCL (HashiCorp Configuration Language)

- **Pros**: Terraform-like syntax, good for infrastructure
- **Cons**: Less familiar to developers, limited ecosystem
- **Decision**: Rejected due to learning curve and limited adoption

### Multiple Formats

- **Pros**: Maximum flexibility for users
- **Cons**: Increased complexity, testing burden, documentation overhead
- **Decision**: Rejected due to maintenance complexity

## Schema Evolution Strategy

### Versioning Approach

- Consider adding `version: v1` field for future schema changes
- Maintain backward compatibility for at least 2 major versions
- Provide migration utilities for breaking changes

### Extension Points

- Support for `extensions:` or `custom:` sections for future features
- Environment-specific overrides via includes or overlays
- Plugin system for custom validation rules

## Validation Strategy

### Static Validation

- Schema validation against defined structs
- Cross-reference validation (team names, repository existence)
- Security validation (no secrets in plain text)

### Runtime Validation

- GitHub API validation (organization exists, teams exist)
- Permission validation (user has required access)
- Dry-run mode for safe validation

## Related Decisions

- **Secret Management**: Environment variables and external secret stores
- **GitOps Integration**: Configuration should be version-controlled
- **Multi-Environment**: Consider environment-specific overlays in future
