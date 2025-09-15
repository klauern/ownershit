# Repository Guidelines

## Project Structure & Module Organization
The CLI lives in Go packages at the repo root, with reusable services such as archiving and GitHub clients split into files like `archiving_v4.go` and `github_v4.go`. Entry points reside under `cmd/ownershit` (runtime CLI) and `cmd/genqlient` (GraphQL code generation). GraphQL-specific logic is isolated in `v4api/`, while canned fixtures, mocks, and end-to-end harnesses live in `mocks/` and `test/e2e/`. Configuration samples, documentation, and analyses can be found in `docs/`, `example-repositories.yaml`, and `GITHUB_API_ANALYSIS.md`.

## Build, Test, and Development Commands
Use Task for repeatable workflows: `task dev` installs local tooling, `task build` compiles binaries for manual inspection, and `task install` places the CLI in your `$GOBIN`. Routine quality gates include `task fmt` for formatting, `task lint` for `golangci-lint run`, and `task test` which regenerates mocks (`go generate ./...`) before executing `go test -race -coverprofile=coverage.out ./...`. Run the CLI locally with `go run ./cmd/ownershit` while iterating on features.

## Coding Style & Naming Conventions
Code must stay `gofmt`/`go fmt` clean; the `task fmt` target runs both canonical formatters. Follow idiomatic Go naming—packages are lower_snake_case, exported symbols use PascalCase, and tests mirror the file they cover (e.g., `branch_test.go`). Keep GraphQL artifacts generated via `go run github.com/Khan/genqlient` in sync after schema changes. Prefer small, composable functions and explicit error returns.

## Testing Guidelines
Unit and integration tests are written with Go's standard testing package; create files ending with `_test.go` beside the production code. Use table-driven tests for new logic and extend the E2E coverage under `cmd/ownershit/*_test.go` or `test/e2e/` when touching workflow-critical paths. Before submitting changes, run `task test` locally; include `task test-all` (tests plus `govulncheck`) for release-adjacent work. Review `coverage.out` via `go tool cover -html=coverage.out` when investigating gaps.

## Commit & Pull Request Guidelines
Match the existing Conventional Commits style (`feat(csv):`, `fix(ci):`, `docs:`) so automated release tooling interprets changes correctly. Group related work into focused commits with imperative subject lines. Pull requests should summarize the behavior change, list validation steps (e.g., `task test` output), and link any Backlog.md tasks. Attach screenshots or config snippets when modifying docs or user-facing flows, and ensure release notes callouts for breaking changes.

## Security & Configuration Tips
Export a scoped `GITHUB_TOKEN` before exercising commands that call live GitHub APIs. Use `mise install` (configured via `mise.toml`) to provision matching `go` and `golangci-lint` versions. Avoid committing sensitive org data—store organization-specific configs outside the repo and parameterize tests with fixtures under `testdata/` when needed.
