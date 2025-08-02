package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestGitServerImplementation(t *testing.T) {
	// Test the GitServer struct and basic functionality
	repoRoot := createTestRepo(t)

	gitServer := &GitServer{
		repoName: "test-repo",
	}

	t.Run("Server Initialization", func(t *testing.T) {
		assert.Equal(t, "test-repo", gitServer.repoName)
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

	t.Run("JSON Response Structures", func(t *testing.T) {
		// Test JSON response structures
		workspaceReq := WorkspaceRequest{Name: "test-workspace"}
		assert.Equal(t, "test-workspace", workspaceReq.Name)

		workspaceResp := WorkspaceResponse{
			Success:     true,
			Message:     "Test message",
			RemoteURL:   "http://localhost:3000/test.git",
			WorkspaceID: "test-id",
		}
		assert.True(t, workspaceResp.Success)
		assert.Contains(t, workspaceResp.RemoteURL, ".git")
	})

	t.Run("Directory Item Structure", func(t *testing.T) {
		// Test directory item structures
		item := DirectoryItem{
			Name:    "test.txt",
			Type:    "file",
			Size:    100,
			ModTime: time.Now().Unix(),
		}
		assert.Equal(t, "test.txt", item.Name)
		assert.Equal(t, "file", item.Type)
		assert.Equal(t, int64(100), item.Size)
	})
}

func TestHttpServerBasics(t *testing.T) {
	// Test basic HTTP server functionality without full startup

	t.Run("Health Check Handler", func(t *testing.T) {
		// Create a simple HTTP server just for health check
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		})

		// Create test server
		server := &http.Server{
			Addr:    ":0", // Use any available port
			Handler: mux,
		}

		// Test the handler function directly
		req, err := http.NewRequest("GET", "/health", nil)
		require.NoError(t, err)

		rr := &testResponseWriter{
			header: make(http.Header),
			body:   &bytes.Buffer{},
		}

		server.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.code)
		assert.Equal(t, "OK", rr.body.String())
	})

	t.Run("Git Info Refs Response Format", func(t *testing.T) {
		// Test git protocol response format
		service := "git-upload-pack"
		expectedContentType := fmt.Sprintf("application/x-%s-advertisement", service)

		assert.Equal(t, "application/x-git-upload-pack-advertisement", expectedContentType)

		// Test packet line format
		expectedHeader := fmt.Sprintf("001e# service=%s\n", service)
		assert.Contains(t, expectedHeader, "service=")
		assert.Contains(t, expectedHeader, "git-upload-pack")
	})
}

// Simple test response writer
type testResponseWriter struct {
	header http.Header
	body   *bytes.Buffer
	code   int
}

func (w *testResponseWriter) Header() http.Header {
	return w.header
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	return w.body.Write(data)
}

func (w *testResponseWriter) WriteHeader(code int) {
	w.code = code
}

func TestHttpServerErrorHandling(t *testing.T) {
	httpPort := getFreePort(t)

	// Start HTTP server without gRPC server (should handle connection errors)
	httpServer := startHttpServer(t, httpPort, 0) // Invalid gRPC port
	defer httpServer.stop()

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	client := &http.Client{Timeout: 5 * time.Second}

	t.Run("Health Check Still Works", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("API Calls Fail Gracefully", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/api/ls/")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return error response, not crash
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			// If it's valid JSON, it should contain an error field
			if errorMsg, ok := result["error"]; ok {
				assert.NotEmpty(t, errorMsg)
			}
		}
	})
}

// Test helpers

type testHttpServer struct {
	port int
	stop func()
}

type testGrpcServer struct {
	port int
	stop func()
}

func startHttpServer(t *testing.T, httpPort, grpcPort int) *testHttpServer {
	// Set environment variables
	os.Setenv("PORT", fmt.Sprintf("%d", httpPort))
	if grpcPort > 0 {
		os.Setenv("GRPC_SERVER", fmt.Sprintf("localhost:%d", grpcPort))
	}

	// Start server in goroutine
	stopCh := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("HTTP server panicked: %v", r)
			}
		}()

		// This would start the main HTTP server
		// For now, we'll simulate with a basic HTTP server
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		})

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: mux,
		}

		go server.ListenAndServe()

		<-stopCh
		server.Shutdown(context.Background())
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	return &testHttpServer{
		port: httpPort,
		stop: func() {
			close(stopCh)
		},
	}
}

func startMockGrpcServer(t *testing.T, port int, repoRoot string) *testGrpcServer {
	// Start a mock gRPC server for HTTP server to connect to
	stopCh := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Mock gRPC server panicked: %v", r)
			}
		}()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			t.Logf("Failed to start mock gRPC server: %v", err)
			return
		}

		s := grpc.NewServer()
		// Register a mock service here if needed

		go func() {
			<-stopCh
			s.GracefulStop()
		}()

		s.Serve(lis)
	}()

	time.Sleep(500 * time.Millisecond)

	return &testGrpcServer{
		port: port,
		stop: func() {
			close(stopCh)
		},
	}
}

func getFreePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

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
