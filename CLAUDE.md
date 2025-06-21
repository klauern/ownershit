# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

This project uses [Task](https://taskfile.dev) for task management. Run tasks with `task <task-name>`:

### Development Setup
- `task dev` - Set up development environment (go mod tidy, install mockgen)

### Building and Installing  
- `task build` - Build binaries locally
- `task install` - Install binary locally

### Code Quality
- `task fmt` - Format code with go fmt and gofmt
- `task lint` - Run golangci-lint
- `task test` - Run tests with race detection and coverage
- `task test-cover` - Run tests and show coverage in browser
- `task mocks` - Generate mocks for testing (runs `go generate ./...`)

### GraphQL Client Generation
- `task gql:download-schema` - Download GitHub's GraphQL schema  
- `task gql:generate-client` - Generate GraphQL client code

### Manual Testing
- `task test-query` - Run a test query with hardcoded username

## Architecture

This is a Go CLI tool for managing GitHub repository ownership and permissions. Key architectural components:

### Core Structure
- **Main CLI**: `cmd/ownershit/main.go` - CLI entry point with commands for sync, branches, and archive operations
- **Configuration**: `config.go` - Defines repository, team, and branch permission structures from YAML config
- **GitHub Integration**: Dual API approach using both GitHub REST v3 and GraphQL v4 APIs

### Key Components
- **GitHub V3 API**: `github_v3.go` - REST API operations for repository management
- **GitHub V4 API**: `github_v4.go` + `v4api/` - GraphQL operations via generated client
- **Generated GraphQL Client**: `v4api/generated.go` - Auto-generated from schema using genqlient
- **Archiving**: `archiving_v4.go` - Repository archiving functionality via GraphQL
- **Branch Management**: `branch.go` - Branch protection and permissions

### Configuration Structure
The tool expects a `repositories.yaml` file defining:
- Organization settings
- Team permissions (admin/push/pull levels)  
- Repository configurations (wiki, issues, projects settings)
- Branch protection rules

### Testing
- Extensive test coverage with `*_test.go` files
- Mock generation using go.uber.org/mock for external dependencies
- Mocks stored in `mocks/` directories

### Dependencies
- CLI framework: `github.com/urfave/cli/v2`
- GitHub APIs: `github.com/google/go-github/v66` (REST), `github.com/shurcooL/githubv4` (GraphQL)
- GraphQL client generation: `github.com/Khan/genqlient`
- Configuration: `gopkg.in/yaml.v3`
- Logging: `github.com/rs/zerolog`