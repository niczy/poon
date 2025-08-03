package get

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
		Use:   "get <workspace-id>",
		Short: "Get workspace information",
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

			resp, err := c.GetClient().GetWorkspace(ctx, &pb.GetWorkspaceRequest{
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
}
