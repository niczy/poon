package create

import (
	"context"
	"fmt"
	"time"

	"github.com/nic/poon/poon-cli/pkg/client"
	pb "github.com/nic/poon/poon-proto/gen/go"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverAddr, _ := cmd.Flags().GetString("server")
			c, err := client.New(serverAddr)
			if err != nil {
				return err
			}
			defer c.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := c.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{
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
}
