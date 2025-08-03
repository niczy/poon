package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/nic/poon/poon-proto/gen/go"
	"github.com/nic/poon/poon-server/storage"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMonorepoServiceServer
	repoRoot      string
	workspaceRoot string
	workspaces    map[string]*Workspace
	mu            sync.RWMutex
	repository    storage.Repository
}

type Workspace struct {
	ID           string
	Name         string
	TrackedPaths []string
	CreatedAt    time.Time
	LastSync     time.Time
	Status       pb.WorkspaceStatus
	Metadata     map[string]string
	GitRepoPath  string
}

func validatePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: path contains '..'")
	}

	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return fmt.Errorf("invalid path: path must be relative and within repository")
	}

	return nil
}

func (s *server) initializeWorkspaceGitRepo(ctx context.Context, gitRepoPath string, trackedPaths []string) error {
	// Create git repository directory
	if err := os.MkdirAll(gitRepoPath, 0755); err != nil {
		return fmt.Errorf("failed to create git repo directory: %v", err)
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.email", "poon-server@example.com")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure git user email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Poon Server")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure git user name: %v", err)
	}

	// Get current version from repository
	currentVersion, err := s.repository.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %v", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no repository versions exist - cannot create workspace")
	}

	// Copy tracked paths from repository to git repo
	for _, path := range trackedPaths {
		if err := s.copyPathToGitRepo(ctx, currentVersion, path, gitRepoPath); err != nil {
			return fmt.Errorf("failed to copy path %s: %v", path, err)
		}
	}

	// Create .poon-workspace metadata file
	metadataContent := fmt.Sprintf(`# Poon Workspace Metadata
# This file is managed by poon-server
workspace_version: 1
tracked_paths:
%s
created_at: %s
`, formatTrackedPaths(trackedPaths), time.Now().Format(time.RFC3339))

	metadataPath := filepath.Join(gitRepoPath, ".poon-workspace")
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		return fmt.Errorf("failed to create metadata file: %v", err)
	}

	// Create .gitignore
	gitignoreContent := `# Poon workspace files
.poon/
*.tmp
.DS_Store
`
	gitignorePath := filepath.Join(gitRepoPath, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %v", err)
	}

	// Add all files to git
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %v", err)
	}

	// Create initial commit
	commitMsg := fmt.Sprintf("Initial workspace commit\n\nTracked paths:\n%s", formatTrackedPaths(trackedPaths))
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %v", err)
	}

	log.Printf("Successfully initialized git repository at %s with %d tracked paths", gitRepoPath, len(trackedPaths))
	return nil
}

func (s *server) copyPathToGitRepo(ctx context.Context, version int64, srcPath string, gitRepoPath string) error {
	// Check if path is a directory or file
	_, err := s.repository.ReadDirectory(ctx, version, srcPath)
	if err != nil {
		// Try as a file
		content, err := s.repository.ReadFile(ctx, version, srcPath)
		if err != nil {
			return fmt.Errorf("path %s not found as file or directory", srcPath)
		}

		// Create target directory if needed
		targetPath := filepath.Join(gitRepoPath, srcPath)
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", targetDir, err)
		}

		// Write file content
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", targetPath, err)
		}

		log.Printf("Copied file: %s", srcPath)
		return nil
	}

	// It's a directory, copy recursively
	return s.copyDirectoryToGitRepo(ctx, version, srcPath, gitRepoPath)
}

func (s *server) copyDirectoryToGitRepo(ctx context.Context, version int64, srcPath string, gitRepoPath string) error {
	entries, err := s.repository.ReadDirectory(ctx, version, srcPath)
	if err != nil {
		return err
	}

	// Create target directory
	targetDir := filepath.Join(gitRepoPath, srcPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", targetDir, err)
	}

	// Copy each entry
	for _, entry := range entries {
		entryPath := filepath.Join(srcPath, entry.Name)

		if entry.Type == storage.ObjectTypeTree {
			// Recursively copy subdirectory
			if err := s.copyDirectoryToGitRepo(ctx, version, entryPath, gitRepoPath); err != nil {
				return err
			}
		} else if entry.Type == storage.ObjectTypeBlob {
			// Copy file
			content, err := s.repository.ReadFile(ctx, version, entryPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", entryPath, err)
			}

			targetPath := filepath.Join(gitRepoPath, entryPath)
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %v", targetPath, err)
			}
		}
	}

	log.Printf("Copied directory: %s (%d entries)", srcPath, len(entries))
	return nil
}

func formatTrackedPaths(paths []string) string {
	result := ""
	for _, path := range paths {
		result += fmt.Sprintf("  - %s\n", path)
	}
	return strings.TrimSuffix(result, "\n")
}

func (s *server) MergePatch(ctx context.Context, req *pb.MergePatchRequest) (*pb.MergePatchResponse, error) {
	log.Printf("Merging patch for path: %s", req.Path)

	if err := validatePath(req.Path); err != nil {
		return &pb.MergePatchResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid path: %v", err),
		}, nil
	}

	if len(req.Patch) == 0 {
		return &pb.MergePatchResponse{
			Success: false,
			Message: "Patch data is empty",
		}, nil
	}

	// Apply patch using content-addressable storage directly
	versionInfo, err := s.repository.ApplyPatch(ctx, req.Patch, req.Author, req.Message)
	if err != nil {
		return &pb.MergePatchResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to apply patch: %v", err),
		}, nil
	}

	log.Printf("Successfully applied patch, created version %d with commit %s", versionInfo.Version, versionInfo.CommitHash)

	return &pb.MergePatchResponse{
		Success:    true,
		Message:    fmt.Sprintf("Patch applied successfully, created version %d", versionInfo.Version),
		CommitHash: string(versionInfo.CommitHash),
	}, nil
}

func (s *server) ReadDirectory(ctx context.Context, req *pb.ReadDirectoryRequest) (*pb.ReadDirectoryResponse, error) {
	log.Printf("Reading directory: %s", req.Path)

	if err := validatePath(req.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Get current version
	currentVersion, err := s.repository.GetCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %v", err)
	}

	if currentVersion == 0 {
		return nil, fmt.Errorf("no repository versions exist - create an initial commit first")
	}

	// Read from content-addressable storage
	entries, err := s.repository.ReadDirectory(ctx, currentVersion, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var items []*pb.DirectoryItem
	for _, entry := range entries {
		item := &pb.DirectoryItem{
			Name:    entry.Name,
			IsDir:   entry.Type == storage.ObjectTypeTree,
			Size:    entry.Size,
			ModTime: entry.ModTime,
		}
		items = append(items, item)
	}

	return &pb.ReadDirectoryResponse{
		Items: items,
	}, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileRequest) (*pb.ReadFileResponse, error) {
	log.Printf("Reading file: %s", req.Path)

	if err := validatePath(req.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Get current version
	currentVersion, err := s.repository.GetCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %v", err)
	}

	if currentVersion == 0 {
		return nil, fmt.Errorf("no repository versions exist - create an initial commit first")
	}

	// Read from content-addressable storage
	content, err := s.repository.ReadFile(ctx, currentVersion, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return &pb.ReadFileResponse{
		Content: content,
	}, nil
}

func (s *server) GetFileHistory(ctx context.Context, req *pb.FileHistoryRequest) (*pb.FileHistoryResponse, error) {
	log.Printf("Getting file history for: %s", req.Path)

	// TODO: Implement actual git log functionality
	// For now, return mock data
	commits := []*pb.Commit{
		{
			Hash:         "abc123",
			Author:       "developer@example.com",
			Message:      fmt.Sprintf("Updated %s", req.Path),
			Timestamp:    time.Now().Unix(),
			ChangedFiles: []string{req.Path},
		},
	}

	return &pb.FileHistoryResponse{
		Commits: commits,
	}, nil
}

func (s *server) GetBranches(ctx context.Context, req *pb.BranchesRequest) (*pb.BranchesResponse, error) {
	log.Printf("Getting branches")

	// TODO: Implement actual git branch listing
	// For now, return mock data
	return &pb.BranchesResponse{
		Branches:      []string{"main", "develop", "feature/test"},
		DefaultBranch: "main",
	}, nil
}

func (s *server) CreateBranch(ctx context.Context, req *pb.CreateBranchRequest) (*pb.CreateBranchResponse, error) {
	log.Printf("Creating branch: %s", req.Name)

	// TODO: Implement actual git branch creation
	// For now, return success
	return &pb.CreateBranchResponse{
		Success:    true,
		Message:    fmt.Sprintf("Branch '%s' created successfully", req.Name),
		BranchName: req.Name,
		CommitHash: "def456",
	}, nil
}

func (s *server) CreateWorkspace(ctx context.Context, req *pb.CreateWorkspaceRequest) (*pb.CreateWorkspaceResponse, error) {
	log.Printf("Creating workspace with tracked paths: %v", req.TrackedPaths)

	// Generate UUID for workspace
	workspaceID := uuid.New().String()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create workspace directory
	workspaceDir := filepath.Join(s.workspaceRoot, workspaceID)
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return &pb.CreateWorkspaceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create workspace directory: %v", err),
		}, nil
	}

	// Initialize git repository
	gitRepoPath := filepath.Join(workspaceDir, "repo")
	if err := s.initializeWorkspaceGitRepo(ctx, gitRepoPath, req.TrackedPaths); err != nil {
		// Clean up on failure
		os.RemoveAll(workspaceDir)
		return &pb.CreateWorkspaceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to initialize git repository: %v", err),
		}, nil
	}

	// Create workspace metadata
	workspace := &Workspace{
		ID:           workspaceID,
		Name:         workspaceID, // Use UUID as name
		TrackedPaths: req.TrackedPaths,
		CreatedAt:    time.Now(),
		LastSync:     time.Now(),
		Status:       pb.WorkspaceStatus_ACTIVE,
		Metadata:     req.Metadata,
		GitRepoPath:  gitRepoPath,
	}

	s.workspaces[workspaceID] = workspace

	// Generate remote URL for poon-git server
	gitServerPort := os.Getenv("GIT_SERVER_PORT")
	if gitServerPort == "" {
		gitServerPort = "3000"
	}
	remoteURL := fmt.Sprintf("http://localhost:%s/%s.git", gitServerPort, workspaceID)

	log.Printf("Successfully created workspace %s with git repo at %s", workspaceID, gitRepoPath)

	return &pb.CreateWorkspaceResponse{
		Success:     true,
		Message:     fmt.Sprintf("Workspace created successfully with %d tracked paths", len(req.TrackedPaths)),
		WorkspaceId: workspaceID,
		RemoteUrl:   remoteURL,
	}, nil
}

func (s *server) GetWorkspace(ctx context.Context, req *pb.GetWorkspaceRequest) (*pb.GetWorkspaceResponse, error) {
	log.Printf("Getting workspace: %s", req.WorkspaceId)

	s.mu.RLock()
	defer s.mu.RUnlock()

	workspace, exists := s.workspaces[req.WorkspaceId]
	if !exists {
		return &pb.GetWorkspaceResponse{
			Success: false,
			Message: "Workspace not found",
		}, nil
	}

	workspaceInfo := &pb.WorkspaceInfo{
		Id:           workspace.ID,
		Name:         workspace.Name,
		TrackedPaths: workspace.TrackedPaths,
		CreatedAt:    workspace.CreatedAt.Format(time.RFC3339),
		LastSync:     workspace.LastSync.Format(time.RFC3339),
		Status:       workspace.Status,
		Metadata:     workspace.Metadata,
	}

	return &pb.GetWorkspaceResponse{
		Success:   true,
		Message:   "Workspace retrieved successfully",
		Workspace: workspaceInfo,
	}, nil
}

func (s *server) UpdateWorkspace(ctx context.Context, req *pb.UpdateWorkspaceRequest) (*pb.UpdateWorkspaceResponse, error) {
	log.Printf("Updating workspace: %s", req.WorkspaceId)

	s.mu.Lock()
	defer s.mu.Unlock()

	workspace, exists := s.workspaces[req.WorkspaceId]
	if !exists {
		return &pb.UpdateWorkspaceResponse{
			Success: false,
			Message: "Workspace not found",
		}, nil
	}

	if len(req.TrackedPaths) > 0 {
		workspace.TrackedPaths = req.TrackedPaths
	}
	if req.Metadata != nil {
		workspace.Metadata = req.Metadata
	}
	workspace.LastSync = time.Now()

	workspaceInfo := &pb.WorkspaceInfo{
		Id:           workspace.ID,
		Name:         workspace.Name,
		TrackedPaths: workspace.TrackedPaths,
		CreatedAt:    workspace.CreatedAt.Format(time.RFC3339),
		LastSync:     workspace.LastSync.Format(time.RFC3339),
		Status:       workspace.Status,
		Metadata:     workspace.Metadata,
	}

	return &pb.UpdateWorkspaceResponse{
		Success:   true,
		Message:   "Workspace updated successfully",
		Workspace: workspaceInfo,
	}, nil
}

func (s *server) DeleteWorkspace(ctx context.Context, req *pb.DeleteWorkspaceRequest) (*pb.DeleteWorkspaceResponse, error) {
	log.Printf("Deleting workspace: %s", req.WorkspaceId)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workspaces[req.WorkspaceId]; !exists {
		return &pb.DeleteWorkspaceResponse{
			Success: false,
			Message: "Workspace not found",
		}, nil
	}

	delete(s.workspaces, req.WorkspaceId)

	return &pb.DeleteWorkspaceResponse{
		Success: true,
		Message: "Workspace deleted successfully",
	}, nil
}

func (s *server) ConfigureSparseCheckout(ctx context.Context, req *pb.SparseCheckoutRequest) (*pb.SparseCheckoutResponse, error) {
	log.Printf("Configuring sparse checkout for %d paths", len(req.Paths))

	// TODO: Implement actual sparse checkout configuration
	// This would involve:
	// 1. Creating a sparse-checkout file
	// 2. Configuring git to use sparse checkout
	// 3. Updating the working directory

	return &pb.SparseCheckoutResponse{
		Success:         true,
		Message:         fmt.Sprintf("Sparse checkout configured for %d paths", len(req.Paths)),
		ConfiguredPaths: req.Paths,
	}, nil
}

func (s *server) DownloadPath(ctx context.Context, req *pb.DownloadPathRequest) (*pb.DownloadPathResponse, error) {
	log.Printf("Downloading path: %s", req.Path)

	// TODO: Implement actual path download with archiving
	// This would involve:
	// 1. Recursively reading the directory/file
	// 2. Creating tar/zip archive based on format
	// 3. Returning the compressed content

	return &pb.DownloadPathResponse{
		Success:  true,
		Message:  fmt.Sprintf("Download prepared for path: %s", req.Path),
		Content:  []byte("mock archive content"),
		Filename: fmt.Sprintf("%s.tar.gz", filepath.Base(req.Path)),
	}, nil
}

func (s *server) AddTrackedPath(ctx context.Context, req *pb.AddTrackedPathRequest) (*pb.AddTrackedPathResponse, error) {
	log.Printf("Adding tracked path %s to workspace %s", req.Path, req.WorkspaceId)

	if err := validatePath(req.Path); err != nil {
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid path: %v", err),
		}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	workspace, exists := s.workspaces[req.WorkspaceId]
	if !exists {
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: "Workspace not found",
		}, nil
	}

	// Check if path already exists in tracked paths
	for _, trackedPath := range workspace.TrackedPaths {
		if trackedPath == req.Path {
			return &pb.AddTrackedPathResponse{
				Success: false,
				Message: fmt.Sprintf("Path %s is already tracked", req.Path),
			}, nil
		}
	}

	// Check if path exists in monorepo
	currentVersion, err := s.repository.GetCurrentVersion(ctx)
	if err != nil {
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get current version: %v", err),
		}, nil
	}

	_, err = s.repository.ReadDirectory(ctx, currentVersion, req.Path)
	if err != nil {
		// Try as file
		_, err = s.repository.ReadFile(ctx, currentVersion, req.Path)
		if err != nil {
			return &pb.AddTrackedPathResponse{
				Success: false,
				Message: fmt.Sprintf("Path %s not found in monorepo: %v", req.Path, err),
			}, nil
		}
	}

	// Add the path to tracked paths
	workspace.TrackedPaths = append(workspace.TrackedPaths, req.Path)
	workspace.LastSync = time.Now()

	// Copy the new path to the workspace git repo
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	if err := s.copyPathToGitRepo(ctx, currentVersion, req.Path, workspace.GitRepoPath); err != nil {
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to copy path to git repo: %v", err),
		}, nil
	}

	// Update .poon-workspace metadata file
	metadataContent := fmt.Sprintf(`# Poon Workspace Metadata
# This file is managed by poon-server
workspace_version: 1
tracked_paths:
%s
created_at: %s
`, formatTrackedPaths(workspace.TrackedPaths), workspace.CreatedAt.Format(time.RFC3339))

	metadataPath := filepath.Join(workspace.GitRepoPath, ".poon-workspace")
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update metadata file: %v", err),
		}, nil
	}

	// Commit the changes
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = workspace.GitRepoPath
	if err := cmd.Run(); err != nil {
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to add files to git: %v", err),
		}, nil
	}

	commitMsg := fmt.Sprintf("Add %s to tracked paths", req.Path)
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = workspace.GitRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if there are no changes to commit
		if strings.Contains(string(output), "nothing to commit") {
			// Still return success, path was already tracked
			return &pb.AddTrackedPathResponse{
				Success:    true,
				Message:    fmt.Sprintf("Path %s was already in workspace", req.Path),
				CommitHash: "",
				NewVersion: currentVersion,
			}, nil
		}
		return &pb.AddTrackedPathResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to commit changes: %v - %s", err, string(output)),
		}, nil
	}

	// Get the commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspace.GitRepoPath
	commitHashBytes, err := cmd.Output()
	commitHash := strings.TrimSpace(string(commitHashBytes))
	if err != nil {
		commitHash = "unknown"
	}

	log.Printf("Successfully added tracked path %s to workspace %s", req.Path, req.WorkspaceId)

	return &pb.AddTrackedPathResponse{
		Success:    true,
		Message:    fmt.Sprintf("Successfully added %s to workspace", req.Path),
		CommitHash: commitHash,
		NewVersion: currentVersion,
	}, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	repoRoot := os.Getenv("REPO_ROOT")
	if repoRoot == "" {
		repoRoot = "."
	}

	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		// Use a temporary directory for workspaces
		var err error
		workspaceRoot, err = os.MkdirTemp("", "poon-workspaces-*")
		if err != nil {
			log.Fatalf("failed to create temporary workspace directory: %v", err)
		}
		log.Printf("Using temporary workspace directory: %s", workspaceRoot)
	} else {
		// Ensure workspace root directory exists if explicitly set
		if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
			log.Fatalf("failed to create workspace root directory: %v", err)
		}
	}

	// Initialize storage backend (in-memory for now)
	backend := storage.NewMemoryBackend()
	repository := storage.NewRepository(backend)

	// Create initial repository version from filesystem if it exists and is empty
	currentVersion, err := repository.GetCurrentVersion(context.Background())
	if err != nil {
		log.Fatalf("failed to get current version: %v", err)
	}

	if currentVersion == 0 {
		// Create initial commit from filesystem
		log.Printf("Creating initial repository version from filesystem: %s", repoRoot)
		_, err := repository.CreateCommitFromFileSystem(context.Background(), repoRoot, "poon-server@example.com", "Initial repository commit")
		if err != nil {
			log.Fatalf("failed to create initial repository version: %v", err)
		}
		log.Printf("✓ Initial repository version created successfully")
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMonorepoServiceServer(s, &server{
		repoRoot:      repoRoot,
		workspaceRoot: workspaceRoot,
		workspaces:    make(map[string]*Workspace),
		repository:    repository,
	})

	log.Printf("gRPC server listening on port %s", port)
	log.Printf("Repository root: %s", repoRoot)
	log.Printf("Workspace root: %s", workspaceRoot)
	log.Printf("Using in-memory content-addressable storage")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
