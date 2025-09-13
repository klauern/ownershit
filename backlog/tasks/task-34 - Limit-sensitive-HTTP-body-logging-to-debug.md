---
id: task-34
title: Limit sensitive HTTP body logging to debug
status: To Do
assignee:
  - '@agent'
created_date: '2025-09-12 19:30'
updated_date: '2025-09-12 19:37'
labels:
  - security
dependencies: []
---

## Description

Avoid logging full response bodies at info/error; guard dumps under debug and consider truncation.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Wrap dumps with debug check
- [x] #2 Truncate if needed
<!-- AC:END -->


## Implementation Plan

1. Wrap dumps with debug check\n2. Truncate body logs\n3. Verify no sensitive logs


## Implementation Notes

Truncated logging via natural dump presence checks; maintained debug gating.
