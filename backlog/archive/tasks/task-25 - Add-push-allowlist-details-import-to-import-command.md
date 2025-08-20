---
id: task-25
title: Add push allowlist details import to import command
status: Done
assignee:
  - '@klauern'
created_date: '2025-08-20'
updated_date: '2025-08-20'
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
- [ ] #2 Import command queries GitHub branch protection restrictions API for push access,Import extracts team names that have push access to protected branches,Import extracts individual usernames that have push access to protected branches,Generated YAML populates push_allowlist with teams and users when restrictions exist,Import correctly shows empty push_allowlist when no push restrictions are configured,Import handles repositories with multiple teams and users in push allowlist
<!-- AC:END -->

## Implementation Plan

1. Research GitHub API for branch protection restrictions to understand push allowlist format
2. Enhance getBranchProtectionRules function to query restrictions API for push access details
3. Extract team names and usernames from restrictions API response 
4. Populate PushAllowlist field in BranchPermissions struct with extracted data
5. Add proper error handling for restrictions API calls
6. Test with repositories that have push restrictions configured
7. Test with repositories that have no push restrictions to ensure empty allowlist

## Implementation Notes

Successfully implemented push allowlist extraction functionality. Enhanced the convertBranchProtection function to extract team slugs and user logins from GitHub branch protection restrictions. Added comprehensive unit tests covering all scenarios including teams only, users only, mixed teams/users, and nil value handling. Verified end-to-end functionality with import command showing correct push_allowlist output. All tests pass and code maintains good quality standards.
