package cat

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
		Use:   "cat <file>",
		Short: "Display file contents",
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

			resp, err := c.GetClient().ReadFile(ctx, &pb.ReadFileRequest{
				Path: args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to read file: %v", err)
			}

			fmt.Print(string(resp.Content))
			return nil
		},
	}
}
