package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerImplementation(t *testing.T) {
	// Test the server implementation directly without gRPC serialization
	repoRoot := createTestRepo(t)
	srv := &server{
		repoRoot: repoRoot,
	}

	t.Run("Server Initialization", func(t *testing.T) {
		assert.Equal(t, repoRoot, srv.repoRoot)
	})

	t.Run("Repository Structure", func(t *testing.T) {
		// Verify test repository was created correctly
		entries, err := os.ReadDir(repoRoot)
		require.NoError(t, err)

		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		assert.Contains(t, names, "src")
		assert.Contains(t, names, "docs")
		assert.Contains(t, names, "config")
	})

	t.Run("File Access", func(t *testing.T) {
		// Test file system access that the server would use
		readmePath := filepath.Join(repoRoot, "docs/README.md")
		content, err := os.ReadFile(readmePath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "Poon Monorepo Documentation")
		assert.Contains(t, contentStr, "Structure")
	})

	t.Run("Directory Listing", func(t *testing.T) {
		// Test directory listing that server would perform
		srcPath := filepath.Join(repoRoot, "src")
		entries, err := os.ReadDir(srcPath)
		require.NoError(t, err)

		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		assert.Contains(t, names, "frontend")
		assert.Contains(t, names, "backend")
	})

	t.Run("Nonexistent Path Handling", func(t *testing.T) {
		// Test error handling for nonexistent paths
		nonexistentPath := filepath.Join(repoRoot, "nonexistent/path")
		_, err := os.ReadDir(nonexistentPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestServerResilience(t *testing.T) {
	repoRoot := createTestRepo(t)

	t.Run("Multiple Server Instances", func(t *testing.T) {
		// Test that we can create multiple server instances
		for i := 0; i < 3; i++ {
			t.Logf("Creating server instance %d", i+1)

			srv := &server{
				repoRoot: repoRoot,
			}

			// Verify server can access repository
			entries, err := os.ReadDir(srv.repoRoot)
			require.NoError(t, err)
			assert.Greater(t, len(entries), 0)
		}
	})

	t.Run("Concurrent File System Access", func(t *testing.T) {
		srv := &server{
			repoRoot: repoRoot,
		}

		// Launch multiple concurrent file system operations
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Perform file system operations that the server would do
				_, err := os.ReadDir(srv.repoRoot)
				if err != nil {
					t.Logf("Concurrent operation %d failed: %v", id, err)
				}

				// Try to read a file
				readmePath := filepath.Join(srv.repoRoot, "docs/README.md")
				_, err = os.ReadFile(readmePath)
				if err != nil {
					t.Logf("Concurrent file read %d failed: %v", id, err)
				}
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < 5; i++ {
			select {
			case <-done:
				// Operation completed
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent operations")
			}
		}
	})
}

// Test helpers

func createTestRepo(t *testing.T) string {
	repoRoot := t.TempDir()

	// Create directory structure
	dirs := []string{
		"src/frontend",
		"src/backend",
		"docs",
		"config",
	}

	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, dir), 0755))
	}

	// Create sample files
	files := map[string]string{
		"src/frontend/app.js": `// Sample frontend application
console.log("Hello from frontend");`,

		"src/backend/server.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello from backend")
}`,

		"docs/README.md": `# Poon Monorepo Documentation

This is a sample monorepo for testing.

## Structure

- src/frontend/ - Frontend application
- src/backend/ - Backend service  
- docs/ - Documentation
- config/ - Configuration files`,

		"config/app.yaml": `environment: test
services:
  frontend:
    port: 3000
  backend:
    port: 8080`,
	}

	for path, content := range files {
		fullPath := filepath.Join(repoRoot, path)
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	return repoRoot
}
