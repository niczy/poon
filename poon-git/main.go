package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pb "github.com/nic/poon/poon-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GitServer struct {
	client   pb.MonorepoServiceClient
	repoName string
}

type DirectoryItem struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"`
}

type DirectoryResponse struct {
	Path  string          `json:"path"`
	Items []DirectoryItem `json:"items"`
}

type SparseCheckoutRequest struct {
	Paths     []string `json:"paths"`
	TargetDir string   `json:"targetDir"`
}

type SparseCheckoutResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Paths   []string `json:"paths"`
}

type WorkspaceRequest struct {
	Name string `json:"name"`
}

type WorkspaceResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	RemoteURL   string `json:"remoteUrl"`
	WorkspaceID string `json:"workspaceId"`
}

type PatchRequest struct {
	WorkspaceID string `json:"workspaceId"`
	Path        string `json:"path"`
	Patch       string `json:"patch"`
	Message     string `json:"message"`
	Author      string `json:"author"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewGitServer(grpcAddr string) (*GitServer, error) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	client := pb.NewMonorepoServiceClient(conn)

	return &GitServer{
		client:   client,
		repoName: "monorepo",
	}, nil
}

// Git HTTP protocol handlers
func (gs *GitServer) handleInfoRefs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")

	if service == "git-upload-pack" {
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
		w.Header().Set("Cache-Control", "no-cache")

		// Git protocol pkt-line format
		fmt.Fprintf(w, "001e# service=%s\n", service)
		fmt.Fprint(w, "0000")

		// TODO: Generate proper git refs from monorepo state
		// For now, return minimal refs
		fmt.Fprint(w, "003f0000000000000000000000000000000000000000 refs/heads/main\x00multi_ack thin-pack\n")
		fmt.Fprint(w, "0000")
	} else {
		http.Error(w, "Service not supported", http.StatusForbidden)
	}
}

func (gs *GitServer) handleUploadPack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// TODO: Implement proper pack generation from monorepo
	// This would involve:
	// 1. Parsing the want/have refs from client request body
	// 2. Fetching relevant data from poon-server
	// 3. Generating git pack format with proper objects

	// For now, return empty pack
	fmt.Fprint(w, "0008NAK\n")

	// Empty pack file header
	packHeader := []byte{
		'P', 'A', 'C', 'K', // signature
		0, 0, 0, 2, // version 2
		0, 0, 0, 0, // number of objects (0)
	}
	w.Write(packHeader)

	// Pack checksum (20 bytes of zeros for empty pack)
	checksum := make([]byte, 20)
	w.Write(checksum)
}

// API handlers for directory listing and file access
func (gs *GitServer) handleListDirectory(w http.ResponseWriter, r *http.Request) {
	targetPath := strings.TrimPrefix(r.URL.Path, "/api/ls/")
	if targetPath == "" {
		targetPath = "."
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := gs.client.ReadDirectory(ctx, &pb.ReadDirectoryRequest{
		Path: targetPath,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	items := make([]DirectoryItem, len(resp.Items))
	for i, item := range resp.Items {
		itemType := "file"
		if item.IsDir {
			itemType = "directory"
		}

		items[i] = DirectoryItem{
			Name:    item.Name,
			Type:    itemType,
			Size:    item.Size,
			ModTime: item.ModTime,
		}
	}

	response := DirectoryResponse{
		Path:  targetPath,
		Items: items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (gs *GitServer) handleReadFile(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/api/cat/")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := gs.client.ReadFile(ctx, &pb.ReadFileRequest{
		Path: filePath,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(resp.Content)
}

func (gs *GitServer) handleSparseCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SparseCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON request"})
		return
	}

	if len(req.Paths) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "paths array is required"})
		return
	}

	// TODO: Implement sparse checkout logic
	// This would involve:
	// 1. Creating a local git repository in targetDir
	// 2. Fetching only the specified paths from poon-server
	// 3. Setting up sparse-checkout configuration
	// 4. Generating appropriate git objects and refs

	response := SparseCheckoutResponse{
		Success: true,
		Message: fmt.Sprintf("Sparse checkout configured for %d paths", len(req.Paths)),
		Paths:   req.Paths,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (gs *GitServer) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON request"})
		return
	}

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "workspace name is required"})
		return
	}

	// TODO: Create workspace directory and git repository
	// TODO: Generate unique workspace ID
	// TODO: Set up workspace-specific git configuration

	workspaceID := fmt.Sprintf("ws_%s_%d", req.Name, time.Now().Unix())
	remoteURL := fmt.Sprintf("http://%s/%s.git", r.Host, req.Name)

	response := WorkspaceResponse{
		Success:     true,
		Message:     fmt.Sprintf("Workspace '%s' created successfully", req.Name),
		RemoteURL:   remoteURL,
		WorkspaceID: workspaceID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (gs *GitServer) handleMergePatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON request"})
		return
	}

	if req.Patch == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "patch content is required"})
		return
	}

	// Send patch to poon-server for merging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := gs.client.MergePatch(ctx, &pb.MergePatchRequest{
		Path:    req.Path,
		Patch:   []byte(req.Patch),
		Message: req.Message,
		Author:  req.Author,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to merge patch: %v", err)})
		return
	}

	response := map[string]interface{}{
		"success":    resp.Success,
		"message":    resp.Message,
		"commitHash": resp.CommitHash,
		"conflicts":  resp.Conflicts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (gs *GitServer) handleDownloadPath(w http.ResponseWriter, r *http.Request) {
	targetPath := r.URL.Query().Get("path")
	if targetPath == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "path parameter is required"})
		return
	}

	// TODO: Recursively download directory contents
	// TODO: Create tar/zip archive of the content
	// TODO: Stream the archive to client

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": fmt.Sprintf("Download initiated for path: %s", targetPath),
		"path":    targetPath,
	}
	json.NewEncoder(w).Encode(response)
}

func (gs *GitServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Git HTTP protocol endpoints
	mux.HandleFunc("/info/refs", gs.handleInfoRefs)
	mux.HandleFunc("/git-upload-pack", gs.handleUploadPack)

	// Custom API endpoints
	mux.HandleFunc("/api/ls/", gs.handleListDirectory)
	mux.HandleFunc("/api/cat/", gs.handleReadFile)
	mux.HandleFunc("/api/sparse-checkout", gs.handleSparseCheckout)

	// Workflow API endpoints
	mux.HandleFunc("/api/workspace", gs.handleCreateWorkspace)
	mux.HandleFunc("/api/merge-patch", gs.handleMergePatch)
	mux.HandleFunc("/api/download", gs.handleDownloadPath)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	grpcServer := os.Getenv("GRPC_SERVER")
	if grpcServer == "" {
		grpcServer = "localhost:50051"
	}

	gitServer, err := NewGitServer(grpcServer)
	if err != nil {
		log.Fatalf("Failed to create git server: %v", err)
	}

	mux := gitServer.setupRoutes()

	log.Printf("Poon Git server listening on port %s", port)
	log.Printf("Connected to gRPC server at %s", grpcServer)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
