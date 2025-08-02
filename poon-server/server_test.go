package main

import (
	"context"
	"fmt"
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
		assert.Contains(t, err.Error(), "invalid path")
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
		assert.Contains(t, err.Error(), "invalid path")
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

func TestPathValidation(t *testing.T) {
	repoRoot := createTestRepo(t)
	srv := &server{
		repoRoot: repoRoot,
	}

	t.Run("Path with Double Dots - ReadFile", func(t *testing.T) {
		testCases := []string{
			"../etc/passwd",
			"../../etc/passwd",
			"src/../../../etc/passwd",
			"docs/../config/../../../etc/passwd",
			"config/app.yaml/../../../etc/passwd",
		}

		for _, path := range testCases {
			t.Run(fmt.Sprintf("Path: %s", path), func(t *testing.T) {
				req := &pb.ReadFileRequest{Path: path}
				resp, err := srv.ReadFile(context.Background(), req)
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "path traversal not allowed")
			})
		}
	})

	t.Run("Path with Double Dots - ReadDirectory", func(t *testing.T) {
		testCases := []string{
			"../etc",
			"../../etc",
			"src/../../../etc",
			"docs/../config/../../../etc",
		}

		for _, path := range testCases {
			t.Run(fmt.Sprintf("Path: %s", path), func(t *testing.T) {
				req := &pb.ReadDirectoryRequest{Path: path}
				resp, err := srv.ReadDirectory(context.Background(), req)
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "path traversal not allowed")
			})
		}
	})

	t.Run("Absolute Paths - ReadFile", func(t *testing.T) {
		testCases := []string{
			"/etc/passwd",
			"/usr/bin/sh",
			"/home/user/.ssh/id_rsa",
		}

		for _, path := range testCases {
			t.Run(fmt.Sprintf("Path: %s", path), func(t *testing.T) {
				req := &pb.ReadFileRequest{Path: path}
				resp, err := srv.ReadFile(context.Background(), req)
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "invalid path")
			})
		}
	})

	t.Run("Absolute Paths - ReadDirectory", func(t *testing.T) {
		testCases := []string{
			"/etc",
			"/usr/bin",
			"/home/user",
		}

		for _, path := range testCases {
			t.Run(fmt.Sprintf("Path: %s", path), func(t *testing.T) {
				req := &pb.ReadDirectoryRequest{Path: path}
				resp, err := srv.ReadDirectory(context.Background(), req)
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "invalid path")
			})
		}
	})

	t.Run("Valid Relative Paths - ReadFile", func(t *testing.T) {
		testCases := []string{
			"docs/README.md",
			"src/frontend/app.js",
			"config/app.yaml",
		}

		for _, path := range testCases {
			t.Run(fmt.Sprintf("Path: %s", path), func(t *testing.T) {
				req := &pb.ReadFileRequest{Path: path}
				resp, err := srv.ReadFile(context.Background(), req)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Content)
			})
		}
	})

	t.Run("Valid Relative Paths - ReadDirectory", func(t *testing.T) {
		testCases := []string{
			"",
			"src",
			"docs",
			"config",
			"src/frontend",
		}

		for _, path := range testCases {
			t.Run(fmt.Sprintf("Path: %s", path), func(t *testing.T) {
				req := &pb.ReadDirectoryRequest{Path: path}
				resp, err := srv.ReadDirectory(context.Background(), req)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Items)
			})
		}
	})

	t.Run("Edge Cases - Clean Path Resolution", func(t *testing.T) {
		testCases := []struct {
			path        string
			shouldError bool
			description string
		}{
			{"./docs/README.md", false, "Current directory prefix"},
			{"docs/./README.md", false, "Current directory in middle"},
			{"docs/../docs/README.md", true, "Parent directory reference"},
			{"docs/README.md", false, "Valid file path"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				req := &pb.ReadFileRequest{Path: tc.path}
				resp, err := srv.ReadFile(context.Background(), req)

				if tc.shouldError {
					assert.Error(t, err)
					assert.Nil(t, resp)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, resp)
				}
			})
		}
	})
}

func TestMergePatchEndpoint(t *testing.T) {
	repoRoot := createTestRepo(t)
	srv := &server{
		repoRoot: repoRoot,
	}

	t.Run("Empty Patch Data", func(t *testing.T) {
		req := &pb.MergePatchRequest{
			Path:    "docs/README.md",
			Patch:   []byte{},
			Message: "Test patch",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Patch data is empty")
	})

	t.Run("Invalid Path", func(t *testing.T) {
		req := &pb.MergePatchRequest{
			Path:    "../../../etc/passwd",
			Patch:   []byte("--- a/test.txt\n+++ b/test.txt\n@@ -1,1 +1,1 @@\n-old\n+new\n"),
			Message: "Test patch",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Invalid path")
	})

	t.Run("Invalid Patch Format", func(t *testing.T) {
		req := &pb.MergePatchRequest{
			Path:    "docs/README.md",
			Patch:   []byte("not a valid patch"),
			Message: "Test patch",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Failed to parse patch")
	})

	t.Run("Apply Simple Patch to Existing File", func(t *testing.T) {
		patch := `--- a/docs/README.md
+++ b/docs/README.md
@@ -1,4 +1,5 @@
 # Poon Monorepo Documentation
 
+This line was added by patch.
 This is a sample monorepo for testing.
 
`

		req := &pb.MergePatchRequest{
			Path:    "docs/README.md",
			Patch:   []byte(patch),
			Message: "Add line to README",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "successfully")
		assert.NotEmpty(t, resp.CommitHash)

		content, err := os.ReadFile(filepath.Join(repoRoot, "docs/README.md"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "This line was added by patch.")
	})

	t.Run("Apply Patch to Create New File", func(t *testing.T) {
		patch := `--- /dev/null
+++ b/docs/NEW_FILE.md
@@ -0,0 +1,3 @@
+# New File
+
+This file was created by a patch.
`

		req := &pb.MergePatchRequest{
			Path:    "docs/NEW_FILE.md",
			Patch:   []byte(patch),
			Message: "Create new file",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "successfully")

		content, err := os.ReadFile(filepath.Join(repoRoot, "docs/NEW_FILE.md"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "This file was created by a patch.")
	})

	t.Run("Apply Multi-hunk Patch", func(t *testing.T) {
		patch := `--- a/config/app.yaml
+++ b/config/app.yaml
@@ -1,2 +1,3 @@
 environment: test
+version: 1.0
 services:
@@ -4,2 +5,3 @@
   backend:
     port: 8080
+    timeout: 30s
`

		req := &pb.MergePatchRequest{
			Path:    "config/app.yaml",
			Patch:   []byte(patch),
			Message: "Update config with version and timeout",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		content, err := os.ReadFile(filepath.Join(repoRoot, "config/app.yaml"))
		require.NoError(t, err)
		contentStr := string(content)
		assert.Contains(t, contentStr, "version: 1.0")
		assert.Contains(t, contentStr, "timeout: 30s")
	})

	t.Run("Invalid Target File in Patch", func(t *testing.T) {
		patch := `--- a/docs/README.md
+++ b/../../../etc/passwd
@@ -1,1 +1,1 @@
-old
+new
`

		req := &pb.MergePatchRequest{
			Path:    "docs/README.md",
			Patch:   []byte(patch),
			Message: "Malicious patch",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Invalid target file in patch")
	})

	t.Run("Patch with Deletion Lines", func(t *testing.T) {
		patch := `--- a/src/frontend/app.js
+++ b/src/frontend/app.js
@@ -1,2 +1,3 @@
-// Sample frontend application
-console.log("Hello from frontend");
+// Updated frontend application
+console.log("Hello from updated frontend");
+console.log("Additional logging");
`

		req := &pb.MergePatchRequest{
			Path:    "src/frontend/app.js",
			Patch:   []byte(patch),
			Message: "Update frontend app",
			Author:  "test@example.com",
		}

		resp, err := srv.MergePatch(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		content, err := os.ReadFile(filepath.Join(repoRoot, "src/frontend/app.js"))
		require.NoError(t, err)
		contentStr := string(content)
		assert.Contains(t, contentStr, "Updated frontend application")
		assert.Contains(t, contentStr, "Additional logging")
		assert.NotContains(t, contentStr, "Sample frontend application")
	})
}

func TestPatchParsing(t *testing.T) {
	t.Run("Valid Simple Patch", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+modified line 2
 line 3
`

		patch, err := parsePatch([]byte(patchData))
		require.NoError(t, err)
		assert.Equal(t, "test.txt", patch.Header.OldFile)
		assert.Equal(t, "test.txt", patch.Header.NewFile)
		assert.Len(t, patch.Hunks, 1)
		
		hunk := patch.Hunks[0]
		assert.Equal(t, 1, hunk.OldStart)
		assert.Equal(t, 3, hunk.OldCount)
		assert.Equal(t, 1, hunk.NewStart)
		assert.Equal(t, 3, hunk.NewCount)
		assert.Len(t, hunk.Lines, 4)
	})

	t.Run("Invalid Patch Format", func(t *testing.T) {
		patchData := `not a valid patch`

		_, err := parsePatch([]byte(patchData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not contain valid unified diff headers")
	})

	t.Run("Empty Patch", func(t *testing.T) {
		_, err := parsePatch([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "patch data is empty")
	})

	t.Run("Multi-hunk Patch", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,2 +1,3 @@
 line 1
+new line
 line 2
@@ -10,1 +11,2 @@
 line 10
+another new line
`

		patch, err := parsePatch([]byte(patchData))
		require.NoError(t, err)
		assert.Len(t, patch.Hunks, 2)
		
		assert.Equal(t, 1, patch.Hunks[0].OldStart)
		assert.Equal(t, 10, patch.Hunks[1].OldStart)
	})
}

func TestPatchValidation(t *testing.T) {
	t.Run("Valid Patch", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,1 +1,1 @@
-old
+new
`
		err := validatePatch([]byte(patchData))
		assert.NoError(t, err)
	})

	t.Run("Missing Headers", func(t *testing.T) {
		patchData := `@@ -1,1 +1,1 @@
-old
+new
`
		err := validatePatch([]byte(patchData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "patch has hunk without proper file headers")
	})

	t.Run("No Hunks", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
`
		err := validatePatch([]byte(patchData))
		assert.NoError(t, err) // Valid headers, no hunks is OK
	})
}
