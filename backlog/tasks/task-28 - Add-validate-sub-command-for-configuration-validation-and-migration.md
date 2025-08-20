---
id: task-28
title: Add validate sub-command for configuration validation and migration
status: To Do
assignee: []
created_date: '2025-08-20'
updated_date: '2025-08-20'
labels: []
dependencies:
  - task-14
---

## Description

Create a dedicated validate sub-command that allows users to validate configuration files, check migration status, perform dry-run migrations, and explicitly migrate configs without performing GitHub operations

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Validate command validates configuration without GitHub operations
- [ ] #2 Command shows current schema version and migration status
- [ ] #3 Dry-run mode shows what migration changes would be made
- [ ] #4 Option to write migrated config to file or update in-place
- [ ] #5 Clear error messages for validation failures
- [ ] #6 Help documentation for validate command usage
<!-- AC:END -->
