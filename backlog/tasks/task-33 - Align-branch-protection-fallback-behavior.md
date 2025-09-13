---
id: task-33
title: Align branch-protection fallback behavior
status: To Do
assignee:
  - '@agent'
created_date: '2025-09-12 19:30'
updated_date: '2025-09-12 19:30'
labels:
  - dx
dependencies: []
---

## Description

Handle existing rule case without noisy error; use typed error or internal fallback to REST cleanly.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Use sentinel error or internal fallback
- [ ] #2 Adjust caller logging
<!-- AC:END -->

## Implementation Plan

1. Introduce typed error or internal fallback\n2. Adjust caller handling\n3. Tests
