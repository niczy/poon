package workspace

import (
	"github.com/nic/poon/poon-cli/internal/commands/workspace/create"
	"github.com/nic/poon/poon-cli/internal/commands/workspace/get"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace management commands",
	}

	cmd.AddCommand(create.NewCommand())
	cmd.AddCommand(get.NewCommand())

	return cmd
}
