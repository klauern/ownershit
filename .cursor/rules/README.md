# Cursor Rules

This directory contains consolidated Cursor rules for the ownershit project.

## Rule Organization

### Always Applied Rules (Context for Every Request)

- **[backlog-workflow.mdc](backlog-workflow.mdc)** - Task management using Backlog.md CLI
- **[development-workflow.mdc](development-workflow.mdc)** - Complete dev cycle: format → lint → test → commit
- **[project-structure.mdc](project-structure.mdc)** - Project layout and module organization
- **[security.mdc](security.mdc)** - Security practices and sensitive data handling

### Context-Aware Rules (Applied Based on File Type)

- **[go-style.mdc](go-style.mdc)** - Go coding conventions (applies to `*.go` files)
- **[testing.mdc](testing.mdc)** - Testing patterns and mocking (applies to `*_test.go` files)
- **[configuration.mdc](configuration.mdc)** - Config file format (applies to `*.yaml`, `config*.go`)
- **[api-integration.mdc](api-integration.mdc)** - GitHub API patterns (applies to `github*.go`, `v4api/**/*.go`)
- **[cli-implementation.mdc](cli-implementation.mdc)** - CLI command patterns (applies to `cmd/**/*.go`)

## Key Improvements from Consolidation

### Eliminated Duplications

**Before:** 10 rules with overlapping content
- `task-commands.mdc`, `commits.mdc` → **Merged into `development-workflow.mdc`**
- `cli-commands.mdc` → **Enhanced as `cli-implementation.mdc`**
- `github-api.mdc` → **Enhanced as `api-integration.mdc`**

**After:** 9 focused rules with clear separation of concerns

### Enhanced Content

1. **Critical API Patterns** - Added specific examples of common mistakes:
   - ❌ Using `Repositories.Get()` for topics (doesn't return topics!)
   - ✅ Using `ListAllTopics()` instead
   - Nil response checking patterns
   - Deterministic output (sorting maps)

2. **Security Best Practices** - Consolidated security guidance:
   - Debug-only HTTP logging with code examples
   - Nil safety patterns
   - Token security rules
   - Pre-release checklist

3. **Testing Patterns** - Better mock examples:
   - Set-based validation (order-independent)
   - gomock patterns with `DoAndReturn`
   - Table-driven test structure

4. **Development Workflow** - Single source of truth:
   - Complete dev cycle from code to PR
   - Conventional commits format
   - Git workflow rules
   - Task commands reference

## Quick Reference

### Most Common Rules

For **new features**:
1. [development-workflow.mdc](development-workflow.mdc) - Format → Lint → Test cycle
2. [cli-implementation.mdc](cli-implementation.mdc) - Add CLI command
3. [api-integration.mdc](api-integration.mdc) - GitHub API patterns
4. [configuration.mdc](configuration.mdc) - Config structure

For **fixing bugs**:
1. [go-style.mdc](go-style.mdc) - Code conventions
2. [testing.mdc](testing.mdc) - Test patterns
3. [api-integration.mdc](api-integration.mdc) - Common API pitfalls

For **task management**:
1. [backlog-workflow.mdc](backlog-workflow.mdc) - Using backlog CLI

## Maintenance

### Adding New Rules

1. Create `*.mdc` file in this directory
2. Add frontmatter metadata:
   ```yaml
   ---
   alwaysApply: true          # OR
   description: "Brief desc"  # OR
   globs: "*.ext,pattern"     # File pattern matching
   ---
   ```
3. Use `[filename](mdc:path/to/filename)` for file references
4. Update this README

### Updating Rules

When updating rules:
- Keep examples concise and practical
- Include both ✅ good and ❌ bad examples
- Reference actual project files with `mdc:` links
- Test with actual AI assistant to verify clarity

### Rule Philosophy

**Good Rules:**
- Actionable and specific
- Include code examples
- Show common mistakes
- Reference actual project files

**Bad Rules:**
- Vague guidance
- Theoretical examples
- Duplicate content from other rules
- No concrete examples
