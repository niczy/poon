package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/nic/poon/poon-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a gRPC client connection
type Client struct {
	conn   *grpc.ClientConn
	client pb.MonorepoServiceClient
}

// New creates a new gRPC client connection
func New(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewMonorepoServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetClient returns the underlying gRPC client
func (c *Client) GetClient() pb.MonorepoServiceClient {
	return c.client
}

// TestConnection tests the gRPC connection by calling GetBranches
func (c *Client) TestConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.GetBranches(ctx, &pb.BranchesRequest{})
	return err
}

// ReadDirectory lists the contents of a directory
func (c *Client) ReadDirectory(ctx context.Context, path string) (*pb.ReadDirectoryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.ReadDirectory(ctx, &pb.ReadDirectoryRequest{Path: path})
}

// CreateWorkspace creates a new workspace on the server
func (c *Client) CreateWorkspace(ctx context.Context, req *pb.CreateWorkspaceRequest) (*pb.CreateWorkspaceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return c.client.CreateWorkspace(ctx, req)
}

// AddTrackedPath adds a tracked path to an existing workspace
func (c *Client) AddTrackedPath(ctx context.Context, workspaceID, path, branch string) (*pb.AddTrackedPathResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return c.client.AddTrackedPath(ctx, &pb.AddTrackedPathRequest{
		WorkspaceId: workspaceID,
		Path:        path,
		Branch:      branch,
	})
}
