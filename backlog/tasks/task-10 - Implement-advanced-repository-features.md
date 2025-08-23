---
id: task-10
title: Implement advanced repository features
status: Done
assignee:
  - '@klauern'
created_date: '2025-07-10'
updated_date: '2025-08-22 23:30'
labels: []
dependencies: []
---

## Description

Add support for auto-delete head branches, discussion features, sponsorship settings, and repository environments

## Acceptance Criteria

- [x] Auto-delete head branches functional
- [x] Discussion features configurable
- [x] Sponsorship settings manageable
- [ ] Repository environments supported (deferred - requires additional research)
- [ ] Deployment protection rules implemented (deferred - complex feature scope)

## Implementation Plan

1. Extend Repository struct with new fields: DeleteBranchOnMerge, HasDiscussionsEnabled, HasSponsorshipsEnabled
2. Update SetRepository GraphQL mutation to include new fields
3. Update validation logic to handle new fields
4. Add tests for configuration validation
5. Update example configurations
6. Test with live GitHub API

## Implementation Notes

Successfully implemented advanced repository features:

## Completed Features:
1. **Auto-delete head branches** - Added DeleteBranchOnMerge field to Repository struct, implemented via REST API (GitHub's GraphQL doesn't support this field)
2. **Discussion features** - Added HasDiscussionsEnabled field, implemented via GraphQL updateRepository mutation 
3. **Sponsorship settings** - Added HasSponsorshipsEnabled field, implemented via GraphQL updateRepository mutation

## Implementation Details:
- Extended Repository struct in config.go with three new optional boolean fields
- Updated SetRepository GraphQL method to handle discussions and sponsorships settings
- Created new SetRepositoryAdvancedSettings REST API method for delete_branch_on_merge 
- Updated configuration parsing to handle nil values gracefully
- Modified MapPermissions function to call both GraphQL and REST API methods
- Added comprehensive tests for new functionality
- Updated example configuration with usage examples and documentation

## Technical Architecture:
- **GraphQL** used for hasDiscussionsEnabled and hasSponsorshipsEnabled (supported in UpdateRepositoryInput)
- **REST API** used for deleteBranchOnMerge (not available in GraphQL schema)
- Backward compatible - all new fields are optional (*bool)
- Proper error handling and logging for both success and failure cases

## Files Modified:
- config.go: Extended Repository struct and MapPermissions function
- github_v4.go: Updated SetRepository method for GraphQL fields  
- github_v3.go: Added SetRepositoryAdvancedSettings method for REST API fields
- github_v3_test.go: Added tests for new REST API method
- github_v4_test.go: Updated tests for modified SetRepository signature
- example-repositories.yaml: Added documentation and examples
