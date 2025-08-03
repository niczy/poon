package commands

import (
	"github.com/nic/poon/poon-cli/internal/commands/branches"
	"github.com/nic/poon/poon-cli/internal/commands/cat"
	"github.com/nic/poon/poon-cli/internal/commands/ls"
	"github.com/nic/poon/poon-cli/internal/commands/push"
	"github.com/nic/poon/poon-cli/internal/commands/start"
	"github.com/nic/poon/poon-cli/internal/commands/status"
	"github.com/nic/poon/poon-cli/internal/commands/sync"
	"github.com/nic/poon/poon-cli/internal/commands/track"
	"github.com/nic/poon/poon-cli/internal/commands/workspace"
	"github.com/spf13/cobra"
)

// AddCommands adds all subcommands to the root command
func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(start.NewCommand())
	rootCmd.AddCommand(track.NewCommand())
	rootCmd.AddCommand(push.NewCommand())
	rootCmd.AddCommand(sync.NewCommand())
	rootCmd.AddCommand(status.NewCommand())
	rootCmd.AddCommand(ls.NewCommand())
	rootCmd.AddCommand(cat.NewCommand())
	rootCmd.AddCommand(branches.NewCommand())
	rootCmd.AddCommand(workspace.NewCommand())
}
