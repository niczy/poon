# Poon Workflow Integration Tests

This directory contains Go integration tests focused on end-to-end workflow validation across all Poon components.

## Test Structure

```
poon-tests/
├── go.mod                  # Go module for tests
├── cli_test.go            # CLI command and error handling tests
├── workflow_test.go       # End-to-end workflow integration tests
├── testutil/              # Test utilities and helpers
│   ├── server.go          # Test server management
│   ├── cli.go             # CLI testing utilities
│   └── workflow.go        # Workflow-specific test helpers
├── monorepo/              # Sample monorepo content (legacy)
└── README.md              # This file
```

## Component-Specific Tests

- **poon-server tests**: Located in `../poon-server/server_test.go`
- **poon-git tests**: Located in `../poon-git/server_test.go`
- **Workflow integration tests**: Located in this directory

## Running Tests

### Prerequisites
- Go 1.21 or later
- All Poon components built and available

### Run All Tests
```bash
cd poon-tests
go test -v ./...
```

### Run Specific Test Suites
```bash
# End-to-end workflow tests
go test -v -run TestFullWorkflowIntegration

# Multi-workspace tests
go test -v -run TestMultiWorkspaceIntegration

# CLI command tests
go test -v -run TestCLIWorkflow

# Error handling and recovery tests
go test -v -run TestWorkflowErrorRecovery
```

### Run Component-Specific Tests
```bash
# gRPC server tests
cd ../poon-server && go test -v

# HTTP server tests  
cd ../poon-git && go test -v

# Return to workflow tests
cd ../poon-tests && go test -v
```

### Run Tests with Timeout
```bash
go test -v -timeout 30s ./...
```

## Test Categories

### 1. CLI Command Tests (`cli_test.go`)

**TestCLIWorkflow**
- ✅ CLI help command
- ✅ Workspace initialization (`poon start`)
- ✅ Workspace status (`poon status`)
- ✅ Directory tracking (`poon track`)
- ✅ Git integration (commit, status, log)
- ⚠️ Push and sync commands (partial implementation)

**TestCLIErrorHandling**
- ✅ Start without workspace name (uses default)
- ✅ Commands without initialized workspace
- ✅ Duplicate workspace initialization
- ✅ Invalid commands

**TestCLICommandValidation**
- ✅ Invalid command handling
- ✅ Missing required arguments
- ✅ Help for all commands

### 2. Workflow Integration Tests (`workflow_test.go`)

**TestFullWorkflowIntegration**
- ✅ End-to-end workflow: CLI → poon-git → poon-server → monorepo
- ✅ Workspace initialization and configuration
- ✅ Git repository creation and management
- ⚠️ Directory tracking workflow (protobuf issues expected)
- ⚠️ Push/sync operations (partial implementation)

**TestMultiWorkspaceIntegration**
- ✅ Multiple independent workspaces
- ✅ Server handling multiple clients
- ✅ Workspace isolation and configuration

**TestWorkflowErrorRecovery**  
- ✅ CLI operations without servers
- ✅ Graceful error handling
- ✅ Server restart during workflow
- ✅ Connection recovery

## Test Utilities

### TestServer (`testutil/server.go`)
- Manages temporary gRPC and HTTP servers
- Creates sample monorepo content
- Provides server lifecycle management
- Handles port allocation and server readiness

### CLIRunner (`testutil/cli.go`)
- Builds and executes CLI commands
- Captures command output and exit codes
- Provides assertion helpers
- Manages workspace directories

### WorkspaceHelper (`testutil/cli.go`)
- Validates workspace structure (`.poon/`, `.git/`)
- Reads configuration and state files
- Provides git command execution
- Creates test files and directories

## Test Features

### Automatic Cleanup
- Uses `t.TempDir()` for isolated test environments
- Automatic server shutdown with `defer server.Stop()`
- No manual cleanup required

### Comprehensive Assertions
- Success/failure validation
- Output content verification
- Configuration file validation
- Git repository state checking

### Server Management
- Automatic port allocation
- Server readiness detection
- Process lifecycle management
- Graceful shutdown handling

## Expected Test Results

### ✅ **Working Tests**
- CLI help and basic commands
- Workspace initialization and configuration
- Git repository creation and management
- Server startup and health checks
- Error handling and validation
- Resource management and cleanup

### ⚠️ **Partial/Expected Failures**
- gRPC operations (protobuf serialization issues)
- Directory tracking functionality
- Patch merging operations
- Some HTTP API endpoints

These partial failures are expected due to protobuf generation issues and incomplete server implementations, but the test framework successfully validates the overall architecture and workflow.

## Running Tests in CI/CD

```bash
#!/bin/bash
# CI test script example

set -e

echo "Building Poon components..."
go build ./poon-server
go build ./poon-git  
go build ./poon-cli

echo "Running integration tests..."
cd poon-tests
go mod tidy
go test -v -timeout 60s ./...

echo "Tests completed successfully!"
```

## Test Development Guidelines

1. **Use testutil helpers** for common operations
2. **Leverage t.TempDir()** for isolated environments
3. **Always defer server.Stop()** for cleanup
4. **Test both success and failure cases**
5. **Use descriptive test names** and subtests
6. **Add logging for expected failures** to aid debugging
7. **Keep tests independent** and parallelizable where possible

## Troubleshooting

### Common Issues

1. **Port conflicts**: Tests automatically allocate free ports
2. **Server startup delays**: Tests include readiness checks
3. **Protobuf errors**: Expected due to generation issues
4. **Build failures**: Ensure parent Go modules are tidy

### Debug Mode
```bash
# Run with verbose output and detailed logging
go test -v -run TestCLIWorkflow -args -debug

# Run single test with race detection
go test -race -run TestServerResilience
```