package ls

import (
	"context"
	"fmt"
	"time"

	"github.com/nic/poon/poon-cli/pkg/client"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ls [path]",
		Short: "List directory contents",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			serverAddr, _ := cmd.Flags().GetString("server")
			c, err := client.New(serverAddr)
			if err != nil {
				return err
			}
			defer c.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := c.ReadDirectory(ctx, path)
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
}
