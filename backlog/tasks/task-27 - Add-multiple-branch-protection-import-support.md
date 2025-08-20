---
id: task-27
title: Add multiple branch protection import support
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

Enhance the import command to discover and extract protection rules for all protected branches in a repository, not just main/master branches, to support repositories that protect develop, staging, release, or other branches with different protection configurations

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Import command discovers all branches with protection rules using GitHub API,Import extracts protection rules for each protected branch individually,Generated YAML includes separate branch protection configuration for each protected branch,Import handles repositories with 10+ protected branches efficiently,Import correctly identifies branch-specific protection rule differences,Import maintains backward compatibility with single main/master branch protection,Import handles repositories with no protected branches gracefully
<!-- AC:END -->
