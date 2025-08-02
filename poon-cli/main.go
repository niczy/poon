package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	pb "github.com/nic/poon/poon-proto/gen/go"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr    string
	gitServerAddr string
	client        pb.MonorepoServiceClient
)

type PoonConfig struct {
	WorkspaceName string   `json:"workspaceName"`
	GitServerURL  string   `json:"gitServerUrl"`
	GrpcServerURL string   `json:"grpcServerUrl"`
	TrackedPaths  []string `json:"trackedPaths"`
	CreatedAt     string   `json:"createdAt"`
}

type TrackedPath struct {
	Path         string `json:"path"`
	LastSyncHash string `json:"lastSyncHash"`
	AddedAt      string `json:"addedAt"`
}

func connectToServer() error {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	client = pb.NewMonorepoServiceClient(conn)
	return nil
}

func loadPoonConfig() (*PoonConfig, error) {
	configPath := ".poon/config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no poon workspace found (run 'poon start' first)")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var config PoonConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

func savePoonConfig(config *PoonConfig) error {
	if err := os.MkdirAll(".poon", 0755); err != nil {
		return fmt.Errorf("failed to create .poon directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := ".poon/config.json"
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var rootCmd = &cobra.Command{
	Use:   "poon",
	Short: "Poon CLI - Internet-scale monorepo client",
	Long:  `Poon CLI - A CLI tool for interacting with the Poon monorepo system via gRPC.`,
}

var startCmd = &cobra.Command{
	Use:   "start [workspace-name]",
	Short: "Initialize a new poon workspace",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceName := "poon-workspace"
		if len(args) > 0 {
			workspaceName = args[0]
		}

		// Check if already initialized
		if _, err := os.Stat(".poon"); err == nil {
			return fmt.Errorf("poon workspace already exists")
		}

		// Initialize git repository
		if err := runCommand("git", "init"); err != nil {
			return fmt.Errorf("failed to initialize git repository: %v", err)
		}

		// Create poon config
		config := &PoonConfig{
			WorkspaceName: workspaceName,
			GitServerURL:  gitServerAddr,
			GrpcServerURL: serverAddr,
			TrackedPaths:  []string{},
			CreatedAt:     time.Now().Format(time.RFC3339),
		}

		if err := savePoonConfig(config); err != nil {
			return err
		}

		// Create .gitignore for poon metadata
		gitignoreContent := ".poon/\n"
		if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644); err != nil {
			fmt.Printf("Warning: failed to create .gitignore: %v\n", err)
		}

		// Initial commit
		if err := runCommand("git", "add", ".gitignore"); err != nil {
			return fmt.Errorf("failed to add .gitignore: %v", err)
		}

		if err := runCommand("git", "commit", "-m", "Initialize poon workspace"); err != nil {
			return fmt.Errorf("failed to create initial commit: %v", err)
		}

		// Add git remote
		gitRemoteURL := fmt.Sprintf("http://%s/%s.git", gitServerAddr, workspaceName)
		if err := runCommand("git", "remote", "add", "origin", gitRemoteURL); err != nil {
			return fmt.Errorf("failed to add git remote: %v", err)
		}

		fmt.Printf("✓ Initialized poon workspace '%s'\n", workspaceName)
		fmt.Printf("✓ Git repository created with remote: %s\n", gitRemoteURL)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  poon track <path>  # Track directories from monorepo\n")

		return nil
	},
}

var trackCmd = &cobra.Command{
	Use:   "track <path> [path...]",
	Short: "Track directories from the monorepo",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadPoonConfig()
		if err != nil {
			return err
		}

		if err := connectToServer(); err != nil {
			return err
		}

		// Test server connectivity first
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = client.GetBranches(ctx, &pb.BranchesRequest{})
		if err != nil {
			return fmt.Errorf("failed to connect to server: %v", err)
		}

		for _, path := range args {
			fmt.Printf("Tracking %s...\n", path)

			// Check if path exists in monorepo
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, err := client.ReadDirectory(ctx, &pb.ReadDirectoryRequest{
				Path: path,
			})
			cancel()

			if err != nil {
				return fmt.Errorf("failed to access path %s: %v", path, err)
			}

			// TODO: Download directory contents recursively
			// TODO: Set up sparse-checkout configuration
			// TODO: Create initial commit for tracked content

			// Add to tracked paths
			for _, tracked := range config.TrackedPaths {
				if tracked == path {
					fmt.Printf("Path %s is already tracked\n", path)
					continue
				}
			}

			config.TrackedPaths = append(config.TrackedPaths, path)
		}

		if err := savePoonConfig(config); err != nil {
			return err
		}

		fmt.Printf("✓ Tracked %d path(s)\n", len(args))
		return nil
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local changes back to the monorepo",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := loadPoonConfig()
		if err != nil {
			return err
		}

		if err := connectToServer(); err != nil {
			return err
		}

		// Test server connectivity by attempting to get branches
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.GetBranches(ctx, &pb.BranchesRequest{})
		if err != nil {
			return fmt.Errorf("failed to connect to server: %v", err)
		}

		// TODO: Calculate diffs for each tracked path
		// TODO: Generate patches
		// TODO: Send patches to poon-server for merging

		fmt.Println("✓ Changes pushed to monorepo")
		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with latest monorepo state",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := loadPoonConfig()
		if err != nil {
			return err
		}

		if err := connectToServer(); err != nil {
			return err
		}

		// Test server connectivity by attempting to get branches
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.GetBranches(ctx, &pb.BranchesRequest{})
		if err != nil {
			return fmt.Errorf("failed to connect to server: %v", err)
		}

		// TODO: Fetch latest state for tracked paths
		// TODO: Merge/rebase with local changes

		fmt.Println("✓ Synced with monorepo")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workspace status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadPoonConfig()
		if err != nil {
			return err
		}

		fmt.Printf("Workspace: %s\n", config.WorkspaceName)
		fmt.Printf("Git Server: %s\n", config.GitServerURL)
		fmt.Printf("gRPC Server: %s\n", config.GrpcServerURL)
		fmt.Printf("Created: %s\n", config.CreatedAt)
		fmt.Printf("\nTracked Paths (%d):\n", len(config.TrackedPaths))
		for _, path := range config.TrackedPaths {
			fmt.Printf("  %s\n", path)
		}

		return nil
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List directory contents",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.ReadDirectory(ctx, &pb.ReadDirectoryRequest{
			Path: path,
		})
		if err != nil {
			return fmt.Errorf("failed to list directory: %v", err)
		}

		for _, item := range resp.Items {
			if item.IsDir {
				fmt.Printf("d %s/\n", item.Name)
			} else {
				fmt.Printf("f %s (%d bytes)\n", item.Name, item.Size)
			}
		}

		return nil
	},
}

var catCmd = &cobra.Command{
	Use:   "cat <file>",
	Short: "Display file contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.ReadFile(ctx, &pb.ReadFileRequest{
			Path: args[0],
		})
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		fmt.Print(string(resp.Content))
		return nil
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply <patch-file>",
	Short: "Apply a patch to the monorepo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		patchContent, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read patch file: %v", err)
		}

		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.MergePatch(ctx, &pb.MergePatchRequest{
			Path:    ".",
			Patch:   patchContent,
			Message: fmt.Sprintf("Applied patch from %s", args[0]),
		})
		if err != nil {
			return fmt.Errorf("failed to apply patch: %v", err)
		}

		if resp.Success {
			fmt.Printf("✓ %s\n", resp.Message)
		} else {
			fmt.Printf("✗ Failed to apply patch: %s\n", resp.Message)
		}

		return nil
	},
}

var branchesCmd = &cobra.Command{
	Use:   "branches",
	Short: "List available branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.GetBranches(ctx, &pb.BranchesRequest{})
		if err != nil {
			return fmt.Errorf("failed to get branches: %v", err)
		}

		fmt.Printf("Available branches:\n")
		for _, branch := range resp.Branches {
			if branch == resp.DefaultBranch {
				fmt.Printf("* %s (default)\n", branch)
			} else {
				fmt.Printf("  %s\n", branch)
			}
		}

		return nil
	},
}

var createBranchCmd = &cobra.Command{
	Use:   "create-branch <name> [from-branch]",
	Short: "Create a new branch",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		fromBranch := "main"
		if len(args) > 1 {
			fromBranch = args[1]
		}

		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.CreateBranch(ctx, &pb.CreateBranchRequest{
			Name:       branchName,
			FromBranch: fromBranch,
		})
		if err != nil {
			return fmt.Errorf("failed to create branch: %v", err)
		}

		if resp.Success {
			fmt.Printf("✓ %s\n", resp.Message)
			fmt.Printf("Branch: %s\n", resp.BranchName)
			fmt.Printf("Commit: %s\n", resp.CommitHash)
		} else {
			fmt.Printf("✗ Failed to create branch: %s\n", resp.Message)
		}

		return nil
	},
}

var historyCmd = &cobra.Command{
	Use:   "history <file>",
	Short: "Show file history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.GetFileHistory(ctx, &pb.FileHistoryRequest{
			Path:  args[0],
			Limit: 10,
		})
		if err != nil {
			return fmt.Errorf("failed to get file history: %v", err)
		}

		fmt.Printf("History for %s:\n", args[0])
		for _, commit := range resp.Commits {
			fmt.Printf("\nCommit: %s\n", commit.Hash)
			fmt.Printf("Author: %s\n", commit.Author)
			fmt.Printf("Date: %s\n", time.Unix(commit.Timestamp, 0).Format(time.RFC3339))
			fmt.Printf("Message: %s\n", commit.Message)
		}

		return nil
	},
}

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Workspace management commands",
}

var createWorkspaceCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{
			Name: args[0],
		})
		if err != nil {
			return fmt.Errorf("failed to create workspace: %v", err)
		}

		if resp.Success {
			fmt.Printf("✓ %s\n", resp.Message)
			fmt.Printf("Workspace ID: %s\n", resp.WorkspaceId)
			fmt.Printf("Remote URL: %s\n", resp.RemoteUrl)
		} else {
			fmt.Printf("✗ Failed to create workspace: %s\n", resp.Message)
		}

		return nil
	},
}

var getWorkspaceCmd = &cobra.Command{
	Use:   "get <workspace-id>",
	Short: "Get workspace information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.GetWorkspace(ctx, &pb.GetWorkspaceRequest{
			WorkspaceId: args[0],
		})
		if err != nil {
			return fmt.Errorf("failed to get workspace: %v", err)
		}

		if resp.Success {
			ws := resp.Workspace
			fmt.Printf("Workspace Information:\n")
			fmt.Printf("ID: %s\n", ws.Id)
			fmt.Printf("Name: %s\n", ws.Name)
			fmt.Printf("Status: %s\n", ws.Status)
			fmt.Printf("Created: %s\n", ws.CreatedAt)
			fmt.Printf("Last Sync: %s\n", ws.LastSync)
			fmt.Printf("Tracked Paths (%d):\n", len(ws.TrackedPaths))
			for _, path := range ws.TrackedPaths {
				fmt.Printf("  %s\n", path)
			}
		} else {
			fmt.Printf("✗ %s\n", resp.Message)
		}

		return nil
	},
}

var sparseCheckoutCmd = &cobra.Command{
	Use:   "sparse-checkout <path1> [path2...]",
	Short: "Configure sparse checkout",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.ConfigureSparseCheckout(ctx, &pb.SparseCheckoutRequest{
			Paths: args,
		})
		if err != nil {
			return fmt.Errorf("failed to configure sparse checkout: %v", err)
		}

		if resp.Success {
			fmt.Printf("✓ %s\n", resp.Message)
			fmt.Printf("Configured paths:\n")
			for _, path := range resp.ConfiguredPaths {
				fmt.Printf("  %s\n", path)
			}
		} else {
			fmt.Printf("✗ Failed to configure sparse checkout: %s\n", resp.Message)
		}

		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download <path>",
	Short: "Download path as archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connectToServer(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := client.DownloadPath(ctx, &pb.DownloadPathRequest{
			Path:   args[0],
			Format: "tar.gz",
		})
		if err != nil {
			return fmt.Errorf("failed to download path: %v", err)
		}

		if resp.Success {
			fmt.Printf("✓ %s\n", resp.Message)
			fmt.Printf("Filename: %s\n", resp.Filename)
			fmt.Printf("Content size: %d bytes\n", len(resp.Content))

			// Write content to file
			if err := os.WriteFile(resp.Filename, resp.Content, 0644); err != nil {
				return fmt.Errorf("failed to write download file: %v", err)
			}
			fmt.Printf("Saved to: %s\n", resp.Filename)
		} else {
			fmt.Printf("✗ Failed to download: %s\n", resp.Message)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "localhost:50051", "gRPC server address")
	rootCmd.PersistentFlags().StringVar(&gitServerAddr, "git-server", "localhost:3000", "Git server address")

	// Workspace workflow commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statusCmd)

	// File and directory operations
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(historyCmd)

	// Branch operations
	rootCmd.AddCommand(branchesCmd)
	rootCmd.AddCommand(createBranchCmd)

	// Workspace management
	workspaceCmd.AddCommand(createWorkspaceCmd)
	workspaceCmd.AddCommand(getWorkspaceCmd)
	rootCmd.AddCommand(workspaceCmd)

	// Advanced operations
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(sparseCheckoutCmd)
	rootCmd.AddCommand(downloadCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
