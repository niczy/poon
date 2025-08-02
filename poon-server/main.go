package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/nic/poon/poon-proto/gen/go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMonorepoServiceServer
	repoRoot   string
	workspaces map[string]*Workspace
	mu         sync.RWMutex
}

type Workspace struct {
	ID           string
	Name         string
	TrackedPaths []string
	CreatedAt    time.Time
	LastSync     time.Time
	Status       pb.WorkspaceStatus
	Metadata     map[string]string
}

func (s *server) MergePatch(ctx context.Context, req *pb.MergePatchRequest) (*pb.MergePatchResponse, error) {
	log.Printf("Merging patch for path: %s", req.Path)

	// TODO: Implement patch merging logic
	// This would involve:
	// 1. Validating the patch
	// 2. Applying the patch to the target files
	// 3. Running any necessary validation/tests
	// 4. Committing the changes

	return &pb.MergePatchResponse{
		Success: true,
		Message: fmt.Sprintf("Patch merged successfully for %s", req.Path),
	}, nil
}

func (s *server) ReadDirectory(ctx context.Context, req *pb.ReadDirectoryRequest) (*pb.ReadDirectoryResponse, error) {
	log.Printf("Reading directory: %s", req.Path)

	fullPath := filepath.Join(s.repoRoot, req.Path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var items []*pb.DirectoryItem
	for _, entry := range entries {
		item := &pb.DirectoryItem{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		}

		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				item.Size = info.Size()
				item.ModTime = info.ModTime().Unix()
			}
		}

		items = append(items, item)
	}

	return &pb.ReadDirectoryResponse{
		Items: items,
	}, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileRequest) (*pb.ReadFileResponse, error) {
	log.Printf("Reading file: %s", req.Path)

	fullPath := filepath.Join(s.repoRoot, req.Path)
	content, err := os.ReadFile(fullPath)
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
	log.Printf("Creating workspace: %s", req.Name)

	s.mu.Lock()
	defer s.mu.Unlock()

	workspaceID := fmt.Sprintf("ws_%s_%d", req.Name, time.Now().Unix())
	workspace := &Workspace{
		ID:           workspaceID,
		Name:         req.Name,
		TrackedPaths: req.TrackedPaths,
		CreatedAt:    time.Now(),
		LastSync:     time.Now(),
		Status:       pb.WorkspaceStatus_ACTIVE,
		Metadata:     req.Metadata,
	}

	s.workspaces[workspaceID] = workspace

	return &pb.CreateWorkspaceResponse{
		Success:     true,
		Message:     fmt.Sprintf("Workspace '%s' created successfully", req.Name),
		WorkspaceId: workspaceID,
		RemoteUrl:   fmt.Sprintf("http://localhost:3000/%s.git", req.Name),
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	repoRoot := os.Getenv("REPO_ROOT")
	if repoRoot == "" {
		repoRoot = "."
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMonorepoServiceServer(s, &server{
		repoRoot:   repoRoot,
		workspaces: make(map[string]*Workspace),
	})

	log.Printf("gRPC server listening on port %s", port)
	log.Printf("Repository root: %s", repoRoot)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
