---
id: task-25
title: Add push allowlist details import to import command
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

Enhance the import command to extract specific teams and users that are allowed to push to protected branches by querying GitHub branch protection restrictions API to populate the push_allowlist configuration

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Import command queries GitHub branch protection restrictions API for push access,Import extracts team names that have push access to protected branches,Import extracts individual usernames that have push access to protected branches,Generated YAML populates push_allowlist with teams and users when restrictions exist,Import correctly shows empty push_allowlist when no push restrictions are configured,Import handles repositories with multiple teams and users in push allowlist
<!-- AC:END -->
