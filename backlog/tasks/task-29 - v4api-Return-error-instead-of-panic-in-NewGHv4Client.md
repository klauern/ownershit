---
id: task-29
title: 'v4api: Return error instead of panic in NewGHv4Client'
status: Done
assignee:
  - '@agent'
created_date: '2025-09-12 19:30'
updated_date: '2025-09-16 00:41'
labels:
  - tech-debt
dependencies: []
---

## Description

Change NewGHv4Client to return (*GitHubV4Client, error) and propagate token validation errors; update call sites and tests.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Change signature and return errors
- [x] #2 Update v4api and cmd callers
- [x] #3 Fix affected tests
<!-- AC:END -->


## Implementation Plan

1. Update signature and usage\n2. Fix tests and callers\n3. Verify build/tests


## Implementation Notes

Updated all call sites and tests, including token validation test to expect error rather than panic.

Validated with go test ./...
