package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/nic/poon/poon-proto/gen/go"
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

func TestReadFileEndpoint(t *testing.T) {
	repoRoot := createTestRepo(t)
	srv := &server{
		repoRoot: repoRoot,
	}

	t.Run("Read Existing File", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "docs/README.md",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)

		content := string(resp.Content)
		assert.Contains(t, content, "Poon Monorepo Documentation")
		assert.Contains(t, content, "Structure")
	})

	t.Run("Read Frontend JavaScript File", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "src/frontend/app.js",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)

		content := string(resp.Content)
		assert.Contains(t, content, "Sample frontend application")
		assert.Contains(t, content, "console.log")
	})

	t.Run("Read Backend Go File", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "src/backend/server.go",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)

		content := string(resp.Content)
		assert.Contains(t, content, "package main")
		assert.Contains(t, content, "Hello from backend")
	})

	t.Run("Read Config YAML File", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "config/app.yaml",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)

		content := string(resp.Content)
		assert.Contains(t, content, "environment: test")
		assert.Contains(t, content, "services:")
	})

	t.Run("Read Nonexistent File", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "nonexistent/file.txt",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("Read File with Empty Path", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Read File with Invalid Path", func(t *testing.T) {
		req := &pb.ReadFileRequest{
			Path: "../../../etc/passwd",
		}

		resp, err := srv.ReadFile(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestReadDirectoryEndpoint(t *testing.T) {
	repoRoot := createTestRepo(t)
	srv := &server{
		repoRoot: repoRoot,
	}

	t.Run("Read Root Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)

		itemNames := make([]string, len(resp.Items))
		for i, item := range resp.Items {
			itemNames[i] = item.Name
		}

		assert.Contains(t, itemNames, "src")
		assert.Contains(t, itemNames, "docs")
		assert.Contains(t, itemNames, "config")

		for _, item := range resp.Items {
			if item.Name == "src" || item.Name == "docs" || item.Name == "config" {
				assert.True(t, item.IsDir, "Directory %s should be marked as directory", item.Name)
			}
		}
	})

	t.Run("Read Src Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "src",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)

		itemNames := make([]string, len(resp.Items))
		for i, item := range resp.Items {
			itemNames[i] = item.Name
		}

		assert.Contains(t, itemNames, "frontend")
		assert.Contains(t, itemNames, "backend")

		for _, item := range resp.Items {
			assert.True(t, item.IsDir, "Item %s should be marked as directory", item.Name)
		}
	})

	t.Run("Read Frontend Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "src/frontend",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)

		itemNames := make([]string, len(resp.Items))
		for i, item := range resp.Items {
			itemNames[i] = item.Name
		}

		assert.Contains(t, itemNames, "app.js")

		for _, item := range resp.Items {
			if item.Name == "app.js" {
				assert.False(t, item.IsDir, "File %s should not be marked as directory", item.Name)
				assert.Greater(t, item.Size, int64(0), "File size should be greater than 0")
				assert.Greater(t, item.ModTime, int64(0), "ModTime should be set")
			}
		}
	})

	t.Run("Read Backend Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "src/backend",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)

		itemNames := make([]string, len(resp.Items))
		for i, item := range resp.Items {
			itemNames[i] = item.Name
		}

		assert.Contains(t, itemNames, "server.go")

		for _, item := range resp.Items {
			if item.Name == "server.go" {
				assert.False(t, item.IsDir, "File %s should not be marked as directory", item.Name)
				assert.Greater(t, item.Size, int64(0), "File size should be greater than 0")
				assert.Greater(t, item.ModTime, int64(0), "ModTime should be set")
			}
		}
	})

	t.Run("Read Docs Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "docs",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)

		itemNames := make([]string, len(resp.Items))
		for i, item := range resp.Items {
			itemNames[i] = item.Name
		}

		assert.Contains(t, itemNames, "README.md")

		for _, item := range resp.Items {
			if item.Name == "README.md" {
				assert.False(t, item.IsDir, "File %s should not be marked as directory", item.Name)
				assert.Greater(t, item.Size, int64(0), "File size should be greater than 0")
				assert.Greater(t, item.ModTime, int64(0), "ModTime should be set")
			}
		}
	})

	t.Run("Read Config Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "config",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Items)

		itemNames := make([]string, len(resp.Items))
		for i, item := range resp.Items {
			itemNames[i] = item.Name
		}

		assert.Contains(t, itemNames, "app.yaml")

		for _, item := range resp.Items {
			if item.Name == "app.yaml" {
				assert.False(t, item.IsDir, "File %s should not be marked as directory", item.Name)
				assert.Greater(t, item.Size, int64(0), "File size should be greater than 0")
				assert.Greater(t, item.ModTime, int64(0), "ModTime should be set")
			}
		}
	})

	t.Run("Read Nonexistent Directory", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "nonexistent/directory",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to read directory")
	})

	t.Run("Read Directory with Invalid Path", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "../../../etc",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Verify File Metadata Accuracy", func(t *testing.T) {
		req := &pb.ReadDirectoryRequest{
			Path: "docs",
		}

		resp, err := srv.ReadDirectory(context.Background(), req)
		require.NoError(t, err)

		for _, item := range resp.Items {
			if item.Name == "README.md" {
				fullPath := filepath.Join(repoRoot, "docs", "README.md")
				info, err := os.Stat(fullPath)
				require.NoError(t, err)

				assert.Equal(t, info.Size(), item.Size)
				assert.Equal(t, info.ModTime().Unix(), item.ModTime)
				assert.False(t, item.IsDir)
			}
		}
	})
}
