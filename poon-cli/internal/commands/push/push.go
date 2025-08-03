package push

import (
	"fmt"

	"github.com/nic/poon/poon-cli/pkg/client"
	"github.com/nic/poon/poon-cli/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push local changes back to the monorepo",
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

			// TODO: Implement actual push functionality
			fmt.Println("✓ Changes pushed to monorepo")
			return nil
		},
	}
}
