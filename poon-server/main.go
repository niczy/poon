package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

type PatchHeader struct {
	OldFile string
	NewFile string
	OldMode string
	NewMode string
}

type PatchHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []PatchLine
}

type PatchLine struct {
	Type    string // "+", "-", " " (context)
	Content string
}

type ParsedPatch struct {
	Header PatchHeader
	Hunks  []PatchHunk
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

func validatePatch(patchData []byte) error {
	if len(patchData) == 0 {
		return fmt.Errorf("patch data is empty")
	}
	
	scanner := bufio.NewScanner(bytes.NewReader(patchData))
	hasValidHeader := false
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			hasValidHeader = true
		}
		if strings.HasPrefix(line, "@@") {
			if !hasValidHeader {
				return fmt.Errorf("patch has hunk without proper file headers")
			}
		}
	}
	
	if !hasValidHeader {
		return fmt.Errorf("patch does not contain valid unified diff headers")
	}
	
	return nil
}

func parsePatch(patchData []byte) (*ParsedPatch, error) {
	if err := validatePatch(patchData); err != nil {
		return nil, err
	}
	
	scanner := bufio.NewScanner(bytes.NewReader(patchData))
	patch := &ParsedPatch{}
	var currentHunk *PatchHunk
	
	hunkRegex := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.HasPrefix(line, "--- ") {
			oldFile := strings.TrimPrefix(line, "--- ")
			if strings.HasPrefix(oldFile, "a/") {
				oldFile = oldFile[2:]
			}
			patch.Header.OldFile = oldFile
		} else if strings.HasPrefix(line, "+++ ") {
			newFile := strings.TrimPrefix(line, "+++ ")
			if strings.HasPrefix(newFile, "b/") {
				newFile = newFile[2:]
			}
			patch.Header.NewFile = newFile
		} else if matches := hunkRegex.FindStringSubmatch(line); matches != nil {
			if currentHunk != nil {
				patch.Hunks = append(patch.Hunks, *currentHunk)
			}
			
			oldStart, _ := strconv.Atoi(matches[1])
			oldCount := 1
			if matches[2] != "" {
				oldCount, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newCount := 1
			if matches[4] != "" {
				newCount, _ = strconv.Atoi(matches[4])
			}
			
			currentHunk = &PatchHunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
			}
		} else if currentHunk != nil && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ")) {
			patchLine := PatchLine{
				Type:    string(line[0]),
				Content: line[1:],
			}
			currentHunk.Lines = append(currentHunk.Lines, patchLine)
		}
	}
	
	if currentHunk != nil {
		patch.Hunks = append(patch.Hunks, *currentHunk)
	}
	
	return patch, nil
}

func backupFile(filePath string) (string, error) {
	backupPath := filePath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
	
	input, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to open file for backup: %v", err)
	}
	defer input.Close()
	
	output, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer output.Close()
	
	_, err = io.Copy(output, input)
	if err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to copy file for backup: %v", err)
	}
	
	return backupPath, nil
}

func applyPatch(filePath string, patch *ParsedPatch) error {
	var originalLines []string
	
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %v", err)
		}
		originalContent := string(content)
		if originalContent != "" {
			originalLines = strings.Split(originalContent, "\n")
			if len(originalLines) > 0 && originalLines[len(originalLines)-1] == "" {
				originalLines = originalLines[:len(originalLines)-1]
			}
		}
	}
	
	result := make([]string, 0, len(originalLines)+100)
	originalIndex := 0
	
	for _, hunk := range patch.Hunks {
		for originalIndex < hunk.OldStart-1 && originalIndex < len(originalLines) {
			result = append(result, originalLines[originalIndex])
			originalIndex++
		}
		
		for _, patchLine := range hunk.Lines {
			switch patchLine.Type {
			case " ":
				if originalIndex < len(originalLines) {
					result = append(result, originalLines[originalIndex])
					originalIndex++
				}
			case "-":
				if originalIndex < len(originalLines) {
					originalIndex++
				}
			case "+":
				result = append(result, patchLine.Content)
			}
		}
	}
	
	for originalIndex < len(originalLines) {
		result = append(result, originalLines[originalIndex])
		originalIndex++
	}
	
	newContent := strings.Join(result, "\n")
	if len(result) > 0 {
		newContent += "\n"
	}
	
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write patched file: %v", err)
	}
	
	return nil
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

	parsed, err := parsePatch(req.Patch)
	if err != nil {
		return &pb.MergePatchResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse patch: %v", err),
		}, nil
	}

	if err := validatePath(parsed.Header.NewFile); err != nil {
		return &pb.MergePatchResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid target file in patch: %v", err),
		}, nil
	}

	targetFile := filepath.Join(s.repoRoot, parsed.Header.NewFile)
	
	backupPath, err := backupFile(targetFile)
	if err != nil {
		return &pb.MergePatchResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to backup file: %v", err),
		}, nil
	}

	if err := applyPatch(targetFile, parsed); err != nil {
		if backupPath != "" {
			if restoreErr := os.Rename(backupPath, targetFile); restoreErr != nil {
				log.Printf("Failed to restore backup: %v", restoreErr)
			}
		}
		return &pb.MergePatchResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to apply patch: %v", err),
		}, nil
	}

	if backupPath != "" {
		if err := os.Remove(backupPath); err != nil {
			log.Printf("Warning: Failed to remove backup file %s: %v", backupPath, err)
		}
	}

	commitHash := fmt.Sprintf("commit_%d", time.Now().Unix())
	
	log.Printf("Successfully applied patch to %s", targetFile)
	
	return &pb.MergePatchResponse{
		Success:    true,
		Message:    fmt.Sprintf("Patch applied successfully to %s", req.Path),
		CommitHash: commitHash,
	}, nil
}

func (s *server) ReadDirectory(ctx context.Context, req *pb.ReadDirectoryRequest) (*pb.ReadDirectoryResponse, error) {
	log.Printf("Reading directory: %s", req.Path)

	if err := validatePath(req.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

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

	if err := validatePath(req.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

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
