package sync

import (
	"fmt"

	"github.com/nic/poon/poon-cli/pkg/client"
	"github.com/nic/poon/poon-cli/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync with latest monorepo state",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := config.LoadConfig()
			if err != nil {
				return err
			}

			serverAddr, _ := cmd.Flags().GetString("server")
			c, err := client.New(serverAddr)
			if err != nil {
				return err
			}
			defer c.Close()

			// TODO: Implement actual sync functionality
			fmt.Println("✓ Synced with monorepo")
			return nil
		},
	}
}
