---
id: task-24
title: Add advanced branch protection settings import
status: To Do
assignee: []
created_date: '2025-08-20 15:03'
labels:
  - enhancement
  - github-api
  - import
  - branch-protection
dependencies:
  - task-22
---

## Description

Enhance the import command to extract complete branch protection configuration including require_conversation_resolution, require_linear_history, allow_force_pushes, and allow_deletions settings that are currently missing from the import output

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Import command extracts require_conversation_resolution setting from GitHub branch protection API,Import command extracts require_linear_history setting from branch protection rules,Import command extracts allow_force_pushes setting from branch protection configuration,Import command extracts allow_deletions setting from branch protection rules,Generated YAML includes all advanced branch protection settings when they exist,Import handles branches with no protection rules without errors
<!-- AC:END -->
