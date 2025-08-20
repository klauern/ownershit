---
id: task-23
title: Add default labels import to import command
status: Done
assignee:
  - '@klauern'
created_date: '2025-08-20'
updated_date: '2025-08-20'
labels:
  - enhancement
  - github-api
  - import
dependencies:
  - task-22
---

## Description

Enhance the import command to extract repository labels with their colors, descriptions, and emojis from GitHub Labels API to populate the default_labels array in YAML output

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Import command extracts all repository labels using GitHub Labels API,Labels include name with color hex codes and descriptions,Generated YAML populates default_labels array with complete label configuration,Empty repositories show empty default_labels array as expected,Import handles repositories with 50+ labels correctly
<!-- AC:END -->

## Implementation Plan

1. Research GitHub v3 API Labels endpoint structure
2. Add getRepositoryLabels function to import.go using similar pattern to existing functions
3. Modify ImportRepositoryConfig to call getRepositoryLabels and populate DefaultLabels
4. Test with repositories containing different label configurations
5. Verify empty repositories return empty default_labels array

## Implementation Notes

### Approach Taken
- Added `getRepositoryLabels` function to `import.go` following the same pattern as other repository import functions
- Modified `ImportRepositoryConfig` to call the new labels function and populate the `DefaultLabels` field
- Used GitHub v3 REST API `Issues.ListLabels` endpoint with pagination support to handle repositories with many labels

### Features Implemented
- **Complete Label Import**: Extracts all repository labels with name, color, and description fields
- **Pagination Support**: Handles repositories with 50+ labels using GitHub's pagination (100 labels per page)
- **Error Handling**: Proper error handling and logging following existing patterns
- **Empty Repository Support**: Returns empty `default_labels` array for repositories without custom labels

### Technical Decisions and Trade-offs  
- **GitHub v3 REST API**: Used Issues.ListLabels endpoint which is simpler than GraphQL for this use case
- **Emoji Field**: GitHub API doesn't provide separate emoji data, so emoji field is populated as empty string (consistent with existing config structure)
- **Pagination**: Set page size to 100 labels per request to minimize API calls while staying within GitHub limits
- **Error Handling**: Used existing `NewGitHubAPIError` pattern for consistency with other import functions

### Modified Files
- `import.go`: Added `getRepositoryLabels` function and integrated it into `ImportRepositoryConfig`
