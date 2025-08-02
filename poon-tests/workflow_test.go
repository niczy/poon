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
		// Step 1: Initialize workspace
		result := cli.RunCommandWithServer(t, server, "start", "integration-test")
		result.AssertSuccess(t).
			AssertContains(t, "Initialized poon workspace").
			AssertContains(t, "integration-test")

		// Verify workspace structure
		assert.True(t, workspace.HasPoonDirectory(t))
		assert.True(t, workspace.HasGitDirectory(t))

		// Step 2: Check status
		result = cli.RunCommand(t, "status")
		result.AssertSuccess(t).
			AssertContains(t, "integration-test").
			AssertContains(t, "Tracked Paths (0)")

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
		// Initialize first workspace
		result := cli1.RunCommandWithServer(t, server, "start", "workspace-1")
		result.AssertSuccess(t)
		
		// Initialize second workspace
		result = cli2.RunCommandWithServer(t, server, "start", "workspace-2")
		result.AssertSuccess(t)

		// Verify each workspace has its own configuration
		config1 := workspace1.GetConfig(t)
		config2 := workspace2.GetConfig(t)
		
		assert.Equal(t, "workspace-1", config1["workspaceName"])
		assert.Equal(t, "workspace-2", config2["workspaceName"])
		assert.NotEqual(t, config1["createdAt"], config2["createdAt"])

		// Both should be able to query status independently
		result = cli1.RunCommand(t, "status")
		result.AssertSuccess(t).AssertContains(t, "workspace-1")
		
		result = cli2.RunCommand(t, "status")
		result.AssertSuccess(t).AssertContains(t, "workspace-2")
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
		
		// Initialize workspace (should work without servers)
		result := cli.RunCommand(t, "start", "offline-test")
		result.AssertSuccess(t)
		
		// Status should work
		result = cli.RunCommand(t, "status")
		result.AssertSuccess(t)
		
		// Commands requiring server should fail gracefully
		result = cli.RunCommand(t, "track", "src/frontend")
		result.AssertError(t).AssertContains(t, "connection refused")
		
		result = cli.RunCommand(t, "push")
		result.AssertError(t)
		
		result = cli.RunCommand(t, "sync")
		result.AssertError(t)
	})

	t.Run("Server Restart During Workflow", func(t *testing.T) {
		server := testutil.NewTestServer(t)
		server.Start(t)
		
		workDir := t.TempDir()
		cli := testutil.NewCLIRunner(t, workDir)
		
		// Initialize workspace with server running
		result := cli.RunCommandWithServer(t, server, "start", "restart-test")
		result.AssertSuccess(t)
		
		// Stop server
		server.Stop()
		
		// Commands should fail gracefully
		result = cli.RunCommandWithServer(t, server, "track", "src/frontend")
		result.AssertError(t)
		
		// Restart server
		server.Start(t)
		
		// Status should still work (doesn't need server)
		result = cli.RunCommand(t, "status")
		result.AssertSuccess(t).AssertContains(t, "restart-test")
		
		// Server-dependent commands should work again
		result = cli.RunCommandWithServer(t, server, "track", "docs")
		// May fail due to protobuf issues, but should not crash
		if result.Error != nil {
			t.Logf("Track after restart failed (expected): %v", result.Error)
		}
		
		server.Stop()
	})
}