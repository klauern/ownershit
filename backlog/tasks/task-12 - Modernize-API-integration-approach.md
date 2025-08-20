---
id: task-12
title: Modernize API integration approach
status: Done
assignee: []
created_date: '2025-07-10'
updated_date: '2025-08-18 03:13'
labels: []
dependencies: []
---

## Description

Update to latest GitHub API features including fine-grained tokens, REST API v4 monitoring, and latest GraphQL schema

## Acceptance Criteria

- [x] Fine-grained token support implemented
- [x] REST API v4 compatibility evaluated
- [x] GraphQL schema updated to latest
- [x] API modernization opportunities identified
- [x] Token permission optimization complete

## Implementation Notes

**API Modernization Completed:**

1. **REST API v4 Evaluation**: Clarified that GitHub has REST API v3 and GraphQL API v4 (not sequential versions). Current go-github v66 is appropriate for REST v3. Identified upgrade path to v74 with breaking changes documented.

2. **Token Permission Optimization**: 
   - Enhanced `ValidateGitHubToken()` with comprehensive token format validation
   - Added `GetRequiredTokenPermissions()` function providing detailed permission guidance for classic and fine-grained tokens  
   - Created new `permissions` CLI command (`ownershit permissions`) to display required scopes and setup instructions
   - Added comprehensive test coverage for permission functions

3. **API Modernization Opportunities Identified**:
   - Upgrade path from go-github v66 → v74 (with careful attention to breaking changes)
   - Enhanced token validation patterns for all GitHub token types (classic, fine-grained, app, OAuth, etc.)
   - Better user guidance for token setup and troubleshooting

4. **Files Modified**:
   - `config.go`: Added `GetRequiredTokenPermissions()` function with detailed permission mappings
   - `cmd/ownershit/main.go`: Added `permissions` command and `permissionsCommand()` function
   - `config_permissions_test.go`: Comprehensive tests for permission functions  
   - `cmd/ownershit/permissions_test.go`: CLI command tests

Fine-grained token support was already implemented. GraphQL client remains on latest schema via genqlient generation.
