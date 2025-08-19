# Project Overview: ownershit

## Purpose

The `ownershit` CLI tool is a comprehensive solution for managing GitHub repository ownership, permissions, and settings across organizations. It provides a unified interface for configuring repository settings, branch protection rules, team permissions, and security features through both GitHub's REST v3 and GraphQL v4 APIs.

## Key Features

### Current Capabilities

- **Team Permissions Management**: Configure admin/push/pull access control for teams
- **Advanced Branch Protection**: Status checks, admin enforcement, push restrictions
- **Repository Settings**: Wiki, issues, projects configuration
- **Label Management**: Sync default labels across repositories with emoji support
- **Repository Archiving**: Automated archiving based on configurable criteria
- **Merge Strategy Control**: Configure allowed merge types (merge commits, squash, rebase)

### Target Users

- **DevOps Engineers**: Managing repository settings at scale
- **Security Teams**: Enforcing security policies across repositories
- **Engineering Managers**: Standardizing team access and workflows
- **Platform Teams**: Automating GitHub organization management

## Technical Architecture

### Dual API Approach

The tool leverages both GitHub APIs to provide comprehensive functionality:

- **REST v3 API**: Team permissions, issue labels, merge strategies
- **GraphQL v4 API**: Repository settings, branch protection, archiving

This dual approach is necessary due to feature gaps in both APIs.

### Core Components

- **Configuration System**: YAML-based repository and permission definitions
- **CLI Interface**: User-friendly commands for common operations
- **Error Handling**: Comprehensive error reporting and recovery
- **Testing Framework**: Extensive unit and integration test coverage

## Current Status

### Completed Phases

- ✅ **Phase 1**: Basic functionality and linting fixes
- ✅ **Phase 2**: Security hardening and error handling
- 🔄 **Phase 3**: Enhanced test coverage (in progress)

### Success Metrics

- Test coverage: 72.1% (main package), targeting 90%+
- Security: Comprehensive token validation and vulnerability scanning
- Performance: \<2s for typical repository sync operations
- Reliability: Zero breaking changes to existing configurations

## Strategic Direction

### Immediate Priorities

1. Complete GraphQL API test coverage enhancement
1. Implement missing GitHub security features
1. Add comprehensive integration testing framework

### Long-term Vision

- Become the de facto tool for GitHub organization management
- Support all major GitHub repository and security features
- Provide enterprise-grade reliability and performance
- Enable GitOps-style repository management workflows
