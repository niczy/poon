// Package poon_tests contains end-to-end tests for the poon monorepo system.
// This file specifically tests the new workspace creation workflow where:
// - CLI takes initial tracking path as argument (not workspace name)
// - Server generates UUID for workspace names
// - Server creates git repositories and copies tracked content
// - Client connects to the server-created workspace
package poon_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nic/poon/poon-tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWorkspaceCreationWorkflow tests the new poon start workflow where:
// 1. CLI takes initial path as argument (not workspace name)
// 2. Server creates workspace with UUID
// 3. Server initializes git repo and copies tracked files
// 4. Client clones/connects to the workspace
func TestNewWorkspaceCreationWorkflow(t *testing.T) {
	server := testutil.NewTestServer(t)
	defer server.Stop()
	server.Start(t)

	// Give server time to initialize
	time.Sleep(2 * time.Second)

	workDir := t.TempDir()
	cli := testutil.NewCLIRunner(t, workDir)
	workspace := testutil.NewWorkspaceHelper(workDir)

	t.Run("CreateWorkspaceWithInitialPath", func(t *testing.T) {
		// Test the new workflow: poon start <initial-path>
		result := cli.RunCommandWithServer(t, server, "start", "src")
		
		if result.Error != nil {
			t.Logf("Command output: %s", result.Output)
			t.Logf("Command error: %v", result.Error)
		}
		
		result.AssertSuccess(t)
		
		// Should contain UUID workspace ID, not "src"
		assert.Contains(t, result.Output, "Server created workspace:")
		assert.Contains(t, result.Output, "Workspace initialized successfully")
		assert.Contains(t, result.Output, "Tracking: src")
		
		// Extract workspace UUID from output
		lines := strings.Split(result.Output, "\n")
		var workspaceID string
		for _, line := range lines {
			if strings.Contains(line, "Server created workspace:") {
				parts := strings.Split(line, ": ")
				if len(parts) >= 2 {
					workspaceID = strings.TrimSpace(parts[1])
					break
				}
			}
		}
		
		require.NotEmpty(t, workspaceID, "Should extract workspace UUID from output")
		t.Logf("Created workspace with ID: %s", workspaceID)
		
		// Workspace ID should be a UUID format (36 characters with dashes)
		assert.Len(t, workspaceID, 36, "Workspace ID should be UUID format")
		assert.Contains(t, workspaceID, "-", "Workspace ID should contain dashes")
	})

	t.Run("VerifyLocalWorkspaceSetup", func(t *testing.T) {
		// Check that local workspace was set up correctly
		assert.True(t, workspace.HasPoonDirectory(t), "Should have .poon directory")
		assert.True(t, workspace.HasGitDirectory(t), "Should have .git directory")
		
		// Check .poon/config.json
		config := workspace.GetConfig(t)
		
		// Workspace name should be the UUID (not "src")
		workspaceName, ok := config["workspaceName"].(string)
		require.True(t, ok, "workspaceName should be string")
		assert.Len(t, workspaceName, 36, "Workspace name should be UUID")
		
		// Tracked paths should include "src"
		trackedPaths, ok := config["trackedPaths"].([]interface{})
		require.True(t, ok, "trackedPaths should be array")
		require.Len(t, trackedPaths, 1, "Should have one tracked path")
		assert.Equal(t, "src", trackedPaths[0], "Should track 'src' path")
		
		// Should have server addresses (test uses dynamic ports)
		grpcServerUrl, ok := config["grpcServerUrl"].(string)
		require.True(t, ok, "grpcServerUrl should be string")
		assert.Contains(t, grpcServerUrl, "localhost:", "Should have localhost grpc server")
		
		gitServerUrl, ok := config["gitServerUrl"].(string)
		require.True(t, ok, "gitServerUrl should be string")
		assert.Contains(t, gitServerUrl, "localhost:", "Should have localhost git server")
		
		t.Logf("Config: %+v", config)
	})

	t.Run("VerifyGitSetup", func(t *testing.T) {
		// Check git configuration
		result := workspace.RunGitCommand(t, "remote", "-v")
		result.AssertSuccess(t)
		
		// Remote should use workspace UUID
		assert.Contains(t, result.Output, ".git", "Should have git remote")
		
		// Check git status
		result = workspace.RunGitCommand(t, "status")
		result.AssertSuccess(t)
		
		// Should be a valid git repository (may have untracked files from test setup)
		assert.Contains(t, result.Output, "On branch", "Should be on a git branch")
	})

	t.Run("VerifyServerWorkspaceDirectory", func(t *testing.T) {
		// Check that server created workspace directory structure
		// Note: This assumes server creates workspaces in ./workspaces/ relative to test
		
		// Look for workspace directories
		entries, err := os.ReadDir(".")
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() == "workspaces" {
					// Found workspaces directory
					workspaceEntries, err := os.ReadDir("workspaces")
					if err == nil && len(workspaceEntries) > 0 {
						// Check first workspace
						workspaceDir := filepath.Join("workspaces", workspaceEntries[0].Name())
						repoDir := filepath.Join(workspaceDir, "repo")
						
						// Check for git repository
						gitDir := filepath.Join(repoDir, ".git")
						if _, err := os.Stat(gitDir); err == nil {
							t.Logf("✓ Server created git repository at %s", repoDir)
							
							// Check for tracked files
							srcDir := filepath.Join(repoDir, "src")
							if _, err := os.Stat(srcDir); err == nil {
								t.Logf("✓ Server copied tracked path 'src'")
								
								// Check for workspace metadata
								metadataFile := filepath.Join(repoDir, ".poon-workspace")
								if content, err := os.ReadFile(metadataFile); err == nil {
									t.Logf("✓ Server created workspace metadata")
									assert.Contains(t, string(content), "workspace_version: 1")
									assert.Contains(t, string(content), "- src")
								}
							}
						}
					}
				}
			}
		}
	})
}

func TestWorkspaceCreationWithDifferentPaths(t *testing.T) {
	server := testutil.NewTestServer(t)
	defer server.Stop()
	server.Start(t)
	
	time.Sleep(2 * time.Second)

	testCases := []struct {
		name        string
		initialPath string
		expectError bool
	}{
		{
			name:        "TrackDocsDirectory",
			initialPath: "docs",
			expectError: false,
		},
		{
			name:        "TrackRootDirectory",
			initialPath: ".",
			expectError: false,
		},
		{
			name:        "TrackNestedPath",
			initialPath: "src/frontend",
			expectError: false,
		},
		{
			name:        "TrackNonexistentPath",
			initialPath: "nonexistent/path",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			cli := testutil.NewCLIRunner(t, workDir)
			
			result := cli.RunCommandWithServer(t, server, "start", tc.initialPath)
			
			if tc.expectError {
				result.AssertError(t)
				assert.Contains(t, result.Output, "failed to access")
			} else {
				if result.Error != nil {
					t.Logf("Unexpected error for path %s: %v", tc.initialPath, result.Error)
					t.Logf("Output: %s", result.Output)
				}
				
				result.AssertSuccess(t)
				assert.Contains(t, result.Output, "Server created workspace:")
				assert.Contains(t, result.Output, fmt.Sprintf("Tracking: %s", tc.initialPath))
				
				// Verify config
				workspace := testutil.NewWorkspaceHelper(workDir)
				config := workspace.GetConfig(t)
				
				trackedPaths, ok := config["trackedPaths"].([]interface{})
				require.True(t, ok, "trackedPaths should be array")
				require.Len(t, trackedPaths, 1, "Should have one tracked path")
				assert.Equal(t, tc.initialPath, trackedPaths[0], "Should track specified path")
			}
		})
	}
}

func TestMultipleWorkspaceCreation(t *testing.T) {
	server := testutil.NewTestServer(t)
	defer server.Stop()
	server.Start(t)
	
	time.Sleep(2 * time.Second)

	t.Run("CreateMultipleWorkspacesWithSamePath", func(t *testing.T) {
		var workspaceIDs []string
		
		// Create multiple workspaces tracking the same path
		for i := 0; i < 3; i++ {
			workDir := t.TempDir()
			cli := testutil.NewCLIRunner(t, workDir)
			
			result := cli.RunCommandWithServer(t, server, "start", "src")
			result.AssertSuccess(t)
			
			// Extract workspace ID
			lines := strings.Split(result.Output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Server created workspace:") {
					parts := strings.Split(line, ": ")
					if len(parts) >= 2 {
						workspaceID := strings.TrimSpace(parts[1])
						workspaceIDs = append(workspaceIDs, workspaceID)
						break
					}
				}
			}
		}
		
		// All workspace IDs should be unique
		require.Len(t, workspaceIDs, 3, "Should have created 3 workspaces")
		
		uniqueIDs := make(map[string]bool)
		for _, id := range workspaceIDs {
			assert.False(t, uniqueIDs[id], "Workspace ID should be unique: %s", id)
			uniqueIDs[id] = true
		}
		
		t.Logf("Created unique workspace IDs: %v", workspaceIDs)
	})
}

func TestWorkspaceCreationErrorHandling(t *testing.T) {
	t.Run("ServerNotRunning", func(t *testing.T) {
		workDir := t.TempDir()
		cli := testutil.NewCLIRunner(t, workDir)
		
		// Try to create workspace without server running
		result := cli.RunCommand(t, "start", "src")
		result.AssertError(t)
		assert.Contains(t, result.Output, "connection refused")
	})

	t.Run("InvalidPath", func(t *testing.T) {
		server := testutil.NewTestServer(t)
		defer server.Stop()
		server.Start(t)
		
		time.Sleep(2 * time.Second)
		
		workDir := t.TempDir()
		cli := testutil.NewCLIRunner(t, workDir)
		
		// Try with path that doesn't exist in monorepo
		result := cli.RunCommandWithServer(t, server, "start", "invalid/path/that/does/not/exist")
		result.AssertError(t)
		
		// Should get path validation error
		assert.Contains(t, result.Output, "failed to access")
	})

	t.Run("WorkspaceAlreadyExists", func(t *testing.T) {
		server := testutil.NewTestServer(t)
		defer server.Stop()
		server.Start(t)
		
		time.Sleep(2 * time.Second)
		
		workDir := t.TempDir()
		cli := testutil.NewCLIRunner(t, workDir)
		
		// Create first workspace
		result := cli.RunCommandWithServer(t, server, "start", "src")
		result.AssertSuccess(t)
		
		// Try to create another workspace in same directory
		result = cli.RunCommandWithServer(t, server, "start", "docs")
		result.AssertError(t)
		assert.Contains(t, result.Output, "poon workspace already exists")
	})
}

// Helper function to extract workspace ID from CLI output
func extractWorkspaceID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Server created workspace:") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}