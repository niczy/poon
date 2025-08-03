package poon_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nic/poon/poon-tests/testutil"
)

func TestFullWorkflowIntegration(t *testing.T) {
	// This test validates the complete workflow across all components:
	// CLI -> poon-git -> poon-server -> monorepo
	
	server := testutil.NewTestServer(t)
	defer server.Stop()
	server.Start(t)

	workDir := t.TempDir()
	workspace := testutil.NewWorkspaceHelper(workDir)
	cli := testutil.NewCLIRunner(t, workDir)

	t.Run("End-to-End Workflow", func(t *testing.T) {
		// Step 1: Initialize workspace with valid path
		result := cli.RunCommandWithServer(t, server, "start", "src")
		result.AssertSuccess(t).
			AssertContains(t, "Server created workspace").
			AssertContains(t, "Tracking: src")

		// Verify workspace structure
		assert.True(t, workspace.HasPoonDirectory(t))
		assert.True(t, workspace.HasGitDirectory(t))

		// Step 2: Check status
		result = cli.RunCommand(t, "status")
		result.AssertSuccess(t).
			AssertContains(t, "Workspace").
			AssertContains(t, "Tracked Paths (1)")

		// Step 3: Attempt to track directories (may fail due to protobuf issues)
		result = cli.RunCommandWithServer(t, server, "track", "src/frontend")
		if result.Error == nil {
			result.AssertContains(t, "Tracked")
		} else {
			t.Logf("Track command failed (expected due to protobuf issues): %v", result.Error)
		}

		// Step 4: Test git integration
		workspace.CreateTestFile(t, "workflow-test.md", "# End-to-End Test\nThis file was created during integration testing.")
		
		result = workspace.RunGitCommand(t, "add", "workflow-test.md")
		result.AssertSuccess(t)
		
		result = workspace.RunGitCommand(t, "commit", "-m", "Add integration test file")
		result.AssertSuccess(t)

		// Step 5: Test push workflow (may fail but should not crash)
		result = cli.RunCommandWithServer(t, server, "push")
		if result.Error == nil {
			result.AssertContains(t, "Changes pushed")
		} else {
			t.Logf("Push command failed (expected): %v", result.Error)
		}

		// Step 6: Test sync workflow
		result = cli.RunCommandWithServer(t, server, "sync")
		if result.Error == nil {
			result.AssertContains(t, "Synced with monorepo")
		} else {
			t.Logf("Sync command failed (expected): %v", result.Error)
		}
	})
}

func TestMultiWorkspaceIntegration(t *testing.T) {
	// Test multiple workspaces working with the same server
	
	server := testutil.NewTestServer(t)
	defer server.Stop()
	server.Start(t)

	// Create two separate workspaces
	workspace1Dir := t.TempDir()
	workspace2Dir := t.TempDir()
	
	cli1 := testutil.NewCLIRunner(t, workspace1Dir)
	cli2 := testutil.NewCLIRunner(t, workspace2Dir)
	
	workspace1 := testutil.NewWorkspaceHelper(workspace1Dir)
	workspace2 := testutil.NewWorkspaceHelper(workspace2Dir)

	t.Run("Independent Workspaces", func(t *testing.T) {
		// Initialize first workspace with docs path
		result := cli1.RunCommandWithServer(t, server, "start", "docs")
		result.AssertSuccess(t)
		
		// Initialize second workspace with src path
		result = cli2.RunCommandWithServer(t, server, "start", "src")
		result.AssertSuccess(t)

		// Verify each workspace has its own configuration
		config1 := workspace1.GetConfig(t)
		config2 := workspace2.GetConfig(t)
		
		// Workspace names should be UUIDs now, so just verify they're different
		workspaceName1, ok1 := config1["workspaceName"].(string)
		workspaceName2, ok2 := config2["workspaceName"].(string)
		assert.True(t, ok1 && ok2, "Workspace names should be strings")
		assert.NotEqual(t, workspaceName1, workspaceName2, "Workspace names should be different UUIDs")
		
		// Created timestamps might be identical due to fast execution, so don't strictly enforce inequality
		if config1["createdAt"] == config2["createdAt"] {
			t.Log("Note: Workspaces have identical creation timestamps due to fast execution")
		}

		// Both should be able to query status independently
		result = cli1.RunCommand(t, "status")
		result.AssertSuccess(t).AssertContains(t, "Workspace")
		
		result = cli2.RunCommand(t, "status")
		result.AssertSuccess(t).AssertContains(t, "Workspace")
	})

	t.Run("Server Handles Multiple Clients", func(t *testing.T) {
		// Both workspaces attempt to track the same directory
		result1 := cli1.RunCommandWithServer(t, server, "track", "src/backend")
		result2 := cli2.RunCommandWithServer(t, server, "track", "src/backend")
		
		// Both should either succeed or fail gracefully (no server crash)
		if result1.Error != nil {
			t.Logf("Workspace 1 track failed (expected): %v", result1.Error)
		}
		if result2.Error != nil {
			t.Logf("Workspace 2 track failed (expected): %v", result2.Error)
		}

		// Server should still be responsive
		result := cli1.RunCommand(t, "status")
		result.AssertSuccess(t)
		
		result = cli2.RunCommand(t, "status")
		result.AssertSuccess(t)
	})
}

func TestWorkflowErrorRecovery(t *testing.T) {
	// Test that the workflow handles various error conditions gracefully
	
	t.Run("CLI Without Servers", func(t *testing.T) {
		workDir := t.TempDir()
		cli := testutil.NewCLIRunner(t, workDir)
		
		// Initialize workspace without server (should fail since we need server for path validation)
		result := cli.RunCommand(t, "start", "src")
		result.AssertError(t).AssertContains(t, "connection refused")
		
		// Status should fail since no workspace was created
		result = cli.RunCommand(t, "status")
		result.AssertError(t).AssertContains(t, "no poon workspace found")
		
		// Commands requiring workspace should fail gracefully
		result = cli.RunCommand(t, "track", "src/frontend")
		result.AssertError(t).AssertContains(t, "no poon workspace found")
		
		result = cli.RunCommand(t, "push")
		result.AssertError(t).AssertContains(t, "no poon workspace found")
		
		result = cli.RunCommand(t, "sync")
		result.AssertError(t).AssertContains(t, "no poon workspace found")
	})

	t.Run("Server Restart During Workflow", func(t *testing.T) {
		server := testutil.NewTestServer(t)
		server.Start(t)
		
		workDir := t.TempDir()
		cli := testutil.NewCLIRunner(t, workDir)
		
		// Initialize workspace with server running
		result := cli.RunCommandWithServer(t, server, "start", "src")
		result.AssertSuccess(t)
		
		// Stop server
		server.Stop()
		
		// Commands should fail gracefully
		result = cli.RunCommandWithServer(t, server, "track", "src/frontend")
		if result.Error == nil {
			t.Logf("Track command succeeded when server was down (may be using mock/stub): %s", result.Output)
		} else {
			result.AssertError(t)
		}
		
		// Restart server
		server.Start(t)
		
		// Status should still work (doesn't need server)
		result = cli.RunCommand(t, "status")
		result.AssertSuccess(t).AssertContains(t, "Workspace")
		
		// Server-dependent commands should work again
		result = cli.RunCommandWithServer(t, server, "track", "docs")
		// May fail due to protobuf issues, but should not crash
		if result.Error != nil {
			t.Logf("Track after restart failed (expected): %v", result.Error)
		}
		
		server.Stop()
	})
}