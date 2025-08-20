---
id: task-24
title: Add advanced branch protection settings import
status: Done
assignee:
  - '@klauern'
created_date: '2025-08-20 15:03'
updated_date: '2025-08-20 15:21'
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
- [x] #1 Import command extracts require_conversation_resolution setting from GitHub branch protection API
- [x] #2 Import command extracts require_linear_history setting from branch protection rules
- [x] #3 Import command extracts allow_force_pushes setting from branch protection configuration
- [x] #4 Import command extracts allow_deletions setting from branch protection rules
- [x] #5 Generated YAML includes all advanced branch protection settings when they exist
- [x] #6 Import handles branches with no protection rules without errors
<!-- AC:END -->

## Implementation Plan

1. Research GitHub Protection struct fields for advanced settings\n2. Check if github.com/google/go-github/v66 exposes the required fields\n3. Update convertBranchProtection function to extract the four missing settings\n4. Add tests to verify the new fields are correctly imported\n5. Test import command with repositories that have these settings configured

## Implementation Notes

Successfully enhanced the import command to extract all four advanced branch protection settings:

- Updated convertBranchProtection() function in import.go to extract require_conversation_resolution, require_linear_history, allow_force_pushes, and allow_deletions from the GitHub Protection struct
- Discovered that github.com/google/go-github/v66 already exposes these fields as nested structs with simple Enabled boolean fields
- Added comprehensive test coverage in import_test.go with 5 test cases covering all scenarios including nil protection, enabled/disabled settings, and mixed configurations
- Verified that generated YAML output includes all advanced settings even when no protection is configured (fields show as null)
- Fixed comment formatting to comply with linter requirements

Modified files:
- import.go: Enhanced convertBranchProtection() function (lines 233-250)
- import_test.go: Added comprehensive test suite for the conversion function

All acceptance criteria met and tests passing.
