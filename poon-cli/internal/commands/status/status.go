package status

import (
	"fmt"

	"github.com/nic/poon/poon-cli/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show workspace status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			fmt.Printf("Workspace: %s\n", cfg.WorkspaceName)
			fmt.Printf("Git Server: %s\n", cfg.GitServerURL)
			fmt.Printf("gRPC Server: %s\n", cfg.GrpcServerURL)
			fmt.Printf("Created: %s\n", cfg.CreatedAt)
			fmt.Printf("\nTracked Paths (%d):\n", len(cfg.TrackedPaths))
			for _, path := range cfg.TrackedPaths {
				fmt.Printf("  %s\n", path)
			}

			return nil
		},
	}
}
