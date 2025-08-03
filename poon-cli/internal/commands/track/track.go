package track

import (
	"context"
	"fmt"
	"time"

	"github.com/nic/poon/poon-cli/pkg/client"
	"github.com/nic/poon/poon-cli/pkg/config"
	"github.com/nic/poon/poon-cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand creates the track command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "track <path> [path...]",
		Short: "Track directories from the monorepo",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runTrack,
		Example: `  poon track /src/backend
  poon track /docs /config`,
	}
}

func runTrack(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	serverAddr, _ := cmd.Flags().GetString("server")
	c, err := client.New(serverAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	// Test server connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	// Sync with remote before adding new paths
	if err := util.SyncFromRemote(); err != nil {
		fmt.Printf("Warning: failed to sync with remote: %v\n", err)
		fmt.Printf("Continuing with local state...\n")
	}

	for _, path := range args {
		fmt.Printf("Tracking %s...\n", path)

		// Check if path exists in monorepo
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := c.ReadDirectory(ctx, path)
		cancel()
		if err != nil {
			return fmt.Errorf("failed to access path %s: %v", path, err)
		}

		// Check if already tracked
		alreadyTracked := false
		for _, tracked := range cfg.TrackedPaths {
			if tracked == path {
				fmt.Printf("Path %s is already tracked\n", path)
				alreadyTracked = true
				break
			}
		}
		if alreadyTracked {
			continue
		}

		// Use gRPC to add the tracked path to workspace
		fmt.Printf("  Adding %s to workspace via gRPC...\n", path)
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		addResp, err := c.AddTrackedPath(ctx, cfg.WorkspaceName, path, "main")
		cancel()
		if err != nil {
			return fmt.Errorf("failed to add tracked path %s: %v", path, err)
		}

		if !addResp.Success {
			return fmt.Errorf("server failed to add path %s: %s", path, addResp.Message)
		}

		// Add to tracked paths in local config
		cfg.TrackedPaths = append(cfg.TrackedPaths, path)
		fmt.Printf("  ✓ Successfully added %s to workspace (commit: %s)\n", path, addResp.CommitHash)

		// Pull the updated main branch from remote
		fmt.Printf("  Pulling latest changes from remote...\n")
		if err := util.GitPull("origin", "main"); err != nil {
			fmt.Printf("  Warning: failed to pull from remote: %v\n", err)
			fmt.Printf("  You can pull later with: git pull origin main\n")
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully tracked %d path(s)\n", len(args))
	fmt.Printf("  Tracked paths: %v\n", cfg.TrackedPaths)
	fmt.Printf("  Remote is synced with main branch\n")
	return nil
}
