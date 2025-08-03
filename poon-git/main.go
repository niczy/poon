package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type GitServer struct {
	workspaceRoot string
}

func NewGitServer(workspaceRoot string) *GitServer {
	return &GitServer{
		workspaceRoot: workspaceRoot,
	}
}

// Extract workspace ID from URL path like /workspace-uuid.git/info/refs
func (gs *GitServer) extractWorkspaceID(path string) string {
	// Match patterns like /workspace-uuid.git/info/refs or /workspace-uuid.git/git-upload-pack
	re := regexp.MustCompile(`^/([a-f0-9-]+)\.git/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// Get the git repository path for a workspace
func (gs *GitServer) getWorkspaceRepoPath(workspaceID string) string {
	return filepath.Join(gs.workspaceRoot, workspaceID, "repo")
}

// Git HTTP protocol handlers
func (gs *GitServer) handleInfoRefs(w http.ResponseWriter, r *http.Request) {
	workspaceID := gs.extractWorkspaceID(r.URL.Path)
	if workspaceID == "" {
		http.Error(w, "Invalid workspace URL", http.StatusNotFound)
		return
	}

	repoPath := gs.getWorkspaceRepoPath(workspaceID)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return
	}

	service := r.URL.Query().Get("service")

	if service == "git-upload-pack" {
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
		w.Header().Set("Cache-Control", "no-cache")

		// Git protocol pkt-line format for service advertisement
		fmt.Fprintf(w, "001e# service=%s\n", service)
		fmt.Fprint(w, "0000")

		// Use git command to get actual refs
		cmd := exec.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
		cmd.Stdout = w
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Printf("Error running git upload-pack: %v", err)
			http.Error(w, "Git command failed", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Service not supported", http.StatusForbidden)
	}
}

func (gs *GitServer) handleUploadPack(w http.ResponseWriter, r *http.Request) {
	workspaceID := gs.extractWorkspaceID(r.URL.Path)
	if workspaceID == "" {
		http.Error(w, "Invalid workspace URL", http.StatusNotFound)
		return
	}

	repoPath := gs.getWorkspaceRepoPath(workspaceID)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// Use git command to handle actual pack generation
	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error running git upload-pack: %v", err)
		// Don't send HTTP error here as we might have already started writing response
		return
	}
}

func (gs *GitServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Git HTTP protocol endpoints for workspace repositories
	// URLs like /workspace-uuid.git/info/refs
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Handle workspace git endpoints
		if strings.Contains(path, ".git/info/refs") {
			gs.handleInfoRefs(w, r)
		} else if strings.Contains(path, ".git/git-upload-pack") {
			gs.handleUploadPack(w, r)
		} else if path == "/health" {
			// Health check
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		log.Fatalf("WORKSPACE_ROOT environment variable must be set for poon-git server")
	}

	// Ensure workspace root directory exists (create it if poon-server hasn't started yet)
	if _, err := os.Stat(workspaceRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
			log.Fatalf("Failed to create workspace root directory: %v", err)
		}
		log.Printf("Created workspace root directory: %s", workspaceRoot)
	}

	gitServer := NewGitServer(workspaceRoot)
	mux := gitServer.setupRoutes()

	log.Printf("Poon Git server listening on port %s", port)
	log.Printf("Serving workspace git repositories from %s", workspaceRoot)
	log.Printf("Git repository URLs: http://localhost:%s/<workspace-uuid>.git", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
