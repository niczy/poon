# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is an internet-scale monorepo system with three main components:

1. **poon-server** (Go): gRPC server that manages the monorepo, supports patch merging, and provides file/directory access
2. **poon-git** (Go): Git-compatible server that enables partial checkout and sparse-checkout functionality by proxying to poon-server
3. **poon-cli** (Go): Command-line interface for direct interaction with the monorepo via gRPC

The system uses Protocol Buffers for communication between components, defined in **poon-proto**.

## Common Development Commands

### Building and Running
```bash
# Build all components
npm run build

# Run individual components in development
npm run dev:grpc-server    # Start gRPC server (Go)
npm run dev:git-server     # Start git-compatible server (Go)  
npm run dev:cli            # Run CLI tool (Go)

# Generate Protocol Buffer files
npm run proto:generate
```

## CLI Workflow

### Initialize Workspace
```bash
# Initialize a new poon workspace
poon start [workspace-name]

# Show workspace status
poon status
```

### Track Monorepo Directories
```bash
# Track one or more directories from the monorepo
poon track /src/frontend
poon track /src/backend /docs

# Use normal git workflow
git branch feature/my-change
git checkout feature/my-change
# Edit files...
git add .
git commit -m "My changes"
```

### Sync with Monorepo
```bash
# Push local changes back to monorepo as patches
poon push

# Sync with latest monorepo state
poon sync [--rebase]
```

### Testing and Linting
```bash
# Run all tests (npm + Go integration tests)
npm run test

# Run workflow integration tests  
cd poon-tests && go test -v ./...

# Run component-specific tests
cd poon-server && go test -v  # gRPC server tests
cd poon-git && go test -v     # HTTP server tests

# Run specific test suite
cd poon-tests && go test -v -run TestFullWorkflowIntegration

# Lint code
npm run lint

# Clean build artifacts
npm run clean
```

### Component-Specific Commands
```bash
# Go components (poon-server, poon-git, poon-cli)
cd poon-server && go run .
cd poon-git && go run .
cd poon-cli && go run . --help
```

## Project Structure

- `/poon-server/` - gRPC server implementing MonorepoService
- `/poon-cli/` - CLI client with workflow and legacy commands
- `/poon-git/` - Git-compatible HTTP server with sparse checkout support
- `/poon-proto/` - Protocol Buffer definitions and generated code
- `/poon-tests/` - Workflow integration tests (end-to-end testing)
- Each component has its own unit tests (e.g., `poon-server/server_test.go`)
- Root workspace manages Go modules and Node.js workspaces

## Key Implementation Details

### gRPC Service (poon-server)
- Implements MergePatch, ReadDirectory, ReadFile operations
- Configurable via PORT and REPO_ROOT environment variables
- Uses file system operations to serve monorepo content

### Git Compatibility (poon-git)
- Exposes Git HTTP protocol endpoints (/info/refs, /git-upload-pack)
- Provides REST API for directory listing and file access (/api/ls/, /api/cat/)
- Supports sparse checkout via /api/sparse-checkout endpoint
- HTTP server with JSON API responses

### CLI Interface (poon-cli)
- Built with Cobra framework
- Connects to gRPC server for all operations
- Workflow commands: start, track, push, sync, status
- Legacy commands: ls, cat, apply
- State management for tracked directories in `.poon/` directory

## Workflow Details

The CLI implements a git-based workflow for working with internet-scale monorepos:

1. **poon start** - Creates local git repo and connects to poon-git server
2. **poon track** - Downloads specific directories from monorepo via gRPC
3. **Normal git workflow** - Users work with familiar git commands (branch, commit, push)
4. **poon push** - Calculates diffs and sends patches to poon-server for merging
5. **poon sync** - Fetches latest monorepo state and merges with local changes

### State Management
- `.poon/config.json` - Workspace configuration
- `.poon/state.json` - File hashes and sync state for tracked paths
- Git integration with sparse-checkout for partial repository access

## Environment Configuration

- `PORT` - Server port (default: 50051 for gRPC, 3000 for git server)
- `GRPC_SERVER` - gRPC server address for git server and CLI
- `REPO_ROOT` - Repository root directory for poon-server