---
id: task-22
title: Add import sub-command to CLI
status: Done
assignee:
  - '@klauern'
created_date: '2025-08-20 02:36'
updated_date: '2025-08-20 02:45'
labels:
  - enhancement
  - cli
  - github-api
dependencies: []
---

## Description

Add an import sub-command that can query GitHub APIs to extract existing repository configuration data and output it in the repositories.yaml format. This will help users migrate existing repositories to ownershit management and identify missing API queries or data components.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Command accepts org/repo input parameter (e.g., 'ownershit import owner/repo')
- [x] #2 Successfully queries GitHub APIs to extract repository configuration data
- [x] #3 Outputs valid YAML that matches repositories.yaml schema structure
- [x] #4 Includes repository settings (wiki enabled/disabled, issues enabled/disabled, projects enabled/disabled)
- [x] #5 Includes team permissions data (admin, push, pull access levels)
- [x] #6 Includes branch protection rules and policies
- [x] #7 Handles GitHub API errors gracefully with meaningful error messages
- [x] #8 Command integrates with existing urfave/cli/v2 framework following project patterns
- [x] #9 Generated YAML can be used directly as input for existing ownershit sync command
<!-- AC:END -->

## Implementation Plan

1. Add import command to CLI structure in main.go following existing patterns
2. Create import function that accepts org/repo parameter
3. Implement GitHub API queries to extract:
   - Repository settings (wiki, issues, projects)
   - Team permissions (admin, push, pull levels)
   - Branch protection rules and policies
4. Create YAML output generation matching PermissionsSettings struct
5. Add comprehensive error handling for GitHub API failures
6. Test import command with real repository to validate output format

## Implementation Notes

## Implementation Notes

Successfully implemented the import sub-command with the following approach:

### Features Implemented:
1. **CLI Integration**: Added import command to main.go following existing urfave/cli patterns
2. **GitHub API Integration**: Created import.go with functions to extract repository data using both v3 REST and v4 GraphQL APIs
3. **Data Extraction**: Implemented extraction of:
   - Repository settings (wiki, issues, projects enabled/disabled)
   - Team permissions (admin, push, pull access levels) 
   - Branch protection rules and merge settings
   - Organization information
4. **YAML Output**: Generated valid YAML matching PermissionsSettings struct schema
5. **Error Handling**: Added comprehensive error handling for GitHub API failures with meaningful messages
6. **File Output**: Support for both stdout and file output via --output/-o flag

### Technical Decisions:
- Used separate configureImportClient function to avoid requiring repositories.yaml config file
- Leveraged existing GitHub client infrastructure and error handling patterns
- Applied consistent logging and debug output throughout
- Used existing permission level constants (Admin, Write, Read) for consistency
- Implemented proper branch protection detection for main/master branches

### Modified Files:
- : Added import command and CLI handling
- : New file with core import functionality
- Fixed linting issues and maintained code quality standards

### Testing:
- Successfully tested with real repositories (klauern/dotfiles, klauern/ownershit)
- Verified YAML output matches repositories.yaml schema
- Confirmed CLI flag parsing and file output functionality
- Validated integration with existing GitHub client and error handling
