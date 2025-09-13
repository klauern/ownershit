---
id: task-32
title: 'Config: nil-safety in MapPermissions logging and SetRepository'
status: To Do
assignee:
  - '@agent'
created_date: '2025-09-12 19:30'
updated_date: '2025-09-12 19:37'
labels:
  - reliability
dependencies: []
---

## Description

Check repo feature pointers for nil before deref/logging; pass only non-nil to SetRepository or default sensibly.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Nil-check wiki/issues/projects in logs
- [x] #2 Adjust SetRepository call
<!-- AC:END -->


## Implementation Plan

1. Add nil checks in logging\n2. Adjust SetRepository call\n3. Run tests

## Implementation Notes

Added nil checks around repo feature flags when logging; adjusted SetRepository to set only non-nil fields (avoids deref panics).
