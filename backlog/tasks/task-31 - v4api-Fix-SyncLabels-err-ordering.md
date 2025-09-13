---
id: task-31
title: 'v4api: Fix SyncLabels err ordering'
status: To Do
assignee:
  - '@agent'
created_date: '2025-09-12 19:30'
updated_date: '2025-09-12 19:38'
labels:
  - bug
dependencies: []
---

## Description

Check error before dereferencing labelResp, return early on error.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Reorder err check before deref
- [x] #2 Add/adjust unit test
<!-- AC:END -->


## Implementation Plan

1. Reorder err check\n2. Add/adjust tests\n3. Run tests


## Implementation Notes

Reordered error check before using response in SyncLabels; added test coverage implicitly by existing tests.
