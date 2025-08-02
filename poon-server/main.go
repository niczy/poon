package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	pb "github.com/nic/poon/poon-proto/gen/go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMonorepoServiceServer
	repoRoot string
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
		repoRoot: repoRoot,
	})

	log.Printf("gRPC server listening on port %s", port)
	log.Printf("Repository root: %s", repoRoot)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
