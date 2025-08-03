package start

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nic/poon/poon-cli/pkg/client"
	"github.com/nic/poon/poon-cli/pkg/config"
	"github.com/nic/poon/poon-cli/pkg/util"
	pb "github.com/nic/poon/poon-proto/gen/go"
	"github.com/spf13/cobra"
)

// NewCommand creates the start command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <initial-path>",
		Short: "Initialize a new poon workspace with initial tracking path",
		Args:  cobra.ExactArgs(1),
		RunE:  runStart,
		Example: `  poon start src/frontend
  poon start docs --server localhost:50051 --git-server localhost:3000`,
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	initialPath := args[0]

	// Check if already initialized
	if _, err := os.Stat(".poon"); err == nil {
		return fmt.Errorf("poon workspace already exists")
	}

	// Get server addresses from flags
	serverAddr, _ := cmd.Flags().GetString("server")
	gitServerAddr, _ := cmd.Flags().GetString("git-server")

	// Connect to server
	c, err := client.New(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer c.Close()

	// Test server connectivity and validate path exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.ReadDirectory(ctx, initialPath)
	if err != nil {
		return fmt.Errorf("failed to access initial path '%s': %v", initialPath, err)
	}

	// Create workspace on server
	fmt.Printf("Creating workspace with initial path: %s\n", initialPath)
	createReq := &pb.CreateWorkspaceRequest{
		Name:         "", // Server will generate UUID
		TrackedPaths: []string{initialPath},
		BaseBranch:   "main",
		Metadata: map[string]string{
			"client_version": "1.0.0",
			"created_by":     "poon-cli",
		},
	}

	createResp, err := c.CreateWorkspace(ctx, createReq)
	if err != nil {
		return fmt.Errorf("failed to create workspace on server: %v", err)
	}

	if !createResp.Success {
		return fmt.Errorf("server failed to create workspace: %s", createResp.Message)
	}

	fmt.Printf("✓ Server created workspace: %s\n", createResp.WorkspaceId)

	// Clone the server-created git repository via poon-git
	gitRemoteURL := createResp.RemoteUrl
	fmt.Printf("Cloning workspace repository from server...\n")

	tempDir := ".poon-temp-clone"
	if err := util.RunCommand("git", "clone", gitRemoteURL, tempDir); err != nil {
		return fmt.Errorf("failed to clone workspace repository: %v", err)
	}

	// Move contents from temp directory to current directory
	if err := util.MoveDirectoryContents(tempDir, "."); err != nil {
		return fmt.Errorf("failed to move cloned repository: %v", err)
	}

	// Clean up temp directory
	if err := os.RemoveAll(tempDir); err != nil {
		fmt.Printf("Warning: failed to clean up temporary directory: %v\n", err)
	}

	// Configure git identity
	if err := util.RunCommand("git", "config", "user.email", "poon@example.com"); err != nil {
		return fmt.Errorf("failed to configure git user email: %v", err)
	}
	if err := util.RunCommand("git", "config", "user.name", "Poon CLI"); err != nil {
		return fmt.Errorf("failed to configure git user name: %v", err)
	}

	fmt.Printf("✓ Successfully cloned workspace repository\n")

	// Create poon config
	cfg := config.CreateConfig(createResp.WorkspaceId, gitServerAddr, serverAddr, []string{initialPath})
	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	// Add .poon/ to .gitignore if not already present
	gitignoreContent := ".poon/\n"
	gitignoreFile := ".gitignore"

	needsGitignore := true
	if existingContent, err := os.ReadFile(gitignoreFile); err == nil {
		if strings.Contains(string(existingContent), ".poon/") {
			needsGitignore = false
		}
	}

	if needsGitignore {
		file, err := os.OpenFile(gitignoreFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Warning: failed to update .gitignore: %v\n", err)
		} else {
			defer file.Close()
			if _, err := file.WriteString(gitignoreContent); err != nil {
				fmt.Printf("Warning: failed to write to .gitignore: %v\n", err)
			} else {
				fmt.Printf("✓ Added .poon/ to .gitignore\n")
			}
		}
	}

	fmt.Printf("✓ Workspace initialized successfully\n")
	fmt.Printf("   Workspace ID: %s\n", createResp.WorkspaceId)
	fmt.Printf("   Tracking: %s\n", initialPath)
	fmt.Printf("   Remote URL: %s\n", gitRemoteURL)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  poon track <path>     # Track additional directories\n")
	fmt.Printf("  poon status           # Show workspace status\n")
	fmt.Printf("  poon sync             # Sync with latest changes\n")

	return nil
}
