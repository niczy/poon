package testutil

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/nic/poon/poon-proto/gen/go"
)

// TestServer manages test server instances for workflow integration testing
type TestServer struct {
	GrpcPort    int
	HttpPort    int
	RepoRoot    string
	grpcCmd     *exec.Cmd
	httpCmd     *exec.Cmd
	grpcClient  pb.MonorepoServiceClient
	grpcConn    *grpc.ClientConn
	mu          sync.Mutex
	running     bool
}

// NewTestServer creates a new test server instance
func NewTestServer(t *testing.T) *TestServer {
	tempDir := t.TempDir()
	
	// Create sample monorepo structure
	setupSampleRepo(t, tempDir)
	
	return &TestServer{
		GrpcPort: GetFreePort(t),
		HttpPort: GetFreePort(t),
		RepoRoot: tempDir,
	}
}

// Start starts both gRPC and HTTP servers
func (ts *TestServer) Start(t *testing.T) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if ts.running {
		return
	}
	
	// Start gRPC server
	ts.startGrpcServer(t)
	
	// Start HTTP server
	ts.startHttpServer(t)
	
	// Wait for servers to be ready
	ts.waitForReady(t)
	
	ts.running = true
}

// Stop stops both servers
func (ts *TestServer) Stop() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if !ts.running {
		return
	}
	
	if ts.grpcConn != nil {
		ts.grpcConn.Close()
	}
	
	if ts.grpcCmd != nil && ts.grpcCmd.Process != nil {
		ts.grpcCmd.Process.Kill()
		ts.grpcCmd.Wait()
	}
	
	if ts.httpCmd != nil && ts.httpCmd.Process != nil {
		ts.httpCmd.Process.Kill()
		ts.httpCmd.Wait()
	}
	
	ts.running = false
}

// GetGrpcClient returns a gRPC client for testing
func (ts *TestServer) GetGrpcClient(t *testing.T) pb.MonorepoServiceClient {
	if ts.grpcClient != nil {
		return ts.grpcClient
	}
	
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", ts.GrpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	
	ts.grpcConn = conn
	ts.grpcClient = pb.NewMonorepoServiceClient(conn)
	return ts.grpcClient
}

// GetHttpURL returns the HTTP server URL
func (ts *TestServer) GetHttpURL() string {
	return fmt.Sprintf("http://localhost:%d", ts.HttpPort)
}

// GetGrpcAddr returns the gRPC server address
func (ts *TestServer) GetGrpcAddr() string {
	return fmt.Sprintf("localhost:%d", ts.GrpcPort)
}

func (ts *TestServer) startGrpcServer(t *testing.T) {
	serverPath := filepath.Join("..", "poon-server")
	workspaceRoot := filepath.Join(ts.RepoRoot, "workspaces")
	
	ts.grpcCmd = exec.Command("go", "run", ".")
	ts.grpcCmd.Dir = serverPath
	ts.grpcCmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", ts.GrpcPort),
		fmt.Sprintf("REPO_ROOT=%s", ts.RepoRoot),
		fmt.Sprintf("WORKSPACE_ROOT=%s", workspaceRoot),
		fmt.Sprintf("GIT_SERVER_PORT=%d", ts.HttpPort),
	)
	
	if err := ts.grpcCmd.Start(); err != nil {
		t.Fatalf("Failed to start gRPC server: %v", err)
	}
}

func (ts *TestServer) startHttpServer(t *testing.T) {
	serverPath := filepath.Join("..", "poon-git")
	workspaceRoot := filepath.Join(ts.RepoRoot, "workspaces")
	
	ts.httpCmd = exec.Command("go", "run", ".")
	ts.httpCmd.Dir = serverPath
	ts.httpCmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", ts.HttpPort),
		fmt.Sprintf("GRPC_SERVER=localhost:%d", ts.GrpcPort),
		fmt.Sprintf("WORKSPACE_ROOT=%s", workspaceRoot),
	)
	
	if err := ts.httpCmd.Start(); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (ts *TestServer) waitForReady(t *testing.T) {
	// Wait for gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timeout waiting for servers to start")
		default:
			conn, err := grpc.Dial(
				fmt.Sprintf("localhost:%d", ts.GrpcPort),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err == nil {
				conn.Close()
				time.Sleep(100 * time.Millisecond) // Give HTTP server time too
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// GetFreePort finds an available port for testing
func GetFreePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	defer listener.Close()
	
	return listener.Addr().(*net.TCPAddr).Port
}

func setupSampleRepo(t *testing.T, repoRoot string) {
	// Create directory structure
	dirs := []string{
		"src/frontend",
		"src/backend", 
		"docs",
		"config",
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(repoRoot, dir), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	
	// Create sample files
	files := map[string]string{
		"src/frontend/app.js": `// Sample frontend application
import React from 'react';
import { render } from 'react-dom';

const App = () => {
  return (
    <div>
      <h1>Welcome to Poon Monorepo</h1>
      <p>This is a sample frontend application.</p>
    </div>
  );
};

render(<App />, document.getElementById('root'));`,
		
		"src/frontend/package.json": `{
  "name": "poon-frontend",
  "version": "1.0.0",
  "description": "Sample frontend application",
  "main": "app.js",
  "dependencies": {
    "react": "^18.0.0",
    "react-dom": "^18.0.0"
  }
}`,

		"src/backend/server.go": `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Message   string    ` + "`json:\"message\"`" + `
	Timestamp time.Time ` + "`json:\"timestamp\"`" + `
	Version   string    ` + "`json:\"version\"`" + `
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Message:   "Poon backend service is healthy",
			Timestamp: time.Now(),
			Version:   "1.0.0",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	fmt.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}`,

		"docs/README.md": `# Poon Monorepo Documentation

This is a sample monorepo for testing the Poon system.

## Structure

- ` + "`src/frontend/`" + ` - React frontend application
- ` + "`src/backend/`" + ` - Go backend service  
- ` + "`docs/`" + ` - Documentation
- ` + "`config/`" + ` - Configuration files`,

		"config/app.yaml": `apiVersion: v1
kind: Config
metadata:
  name: poon-config
spec:
  environment: development
  services:
    frontend:
      port: 3000
      build: webpack
    backend:
      port: 8080
      runtime: go`,
	}
	
	for path, content := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}
}