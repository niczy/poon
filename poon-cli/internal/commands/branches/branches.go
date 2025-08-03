package branches

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
		Use:   "branches",
		Short: "List available branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverAddr, _ := cmd.Flags().GetString("server")
			c, err := client.New(serverAddr)
			if err != nil {
				return err
			}
			defer c.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := c.GetClient().GetBranches(ctx, &pb.BranchesRequest{})
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
}
