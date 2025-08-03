package cli

import (
	"github.com/nic/poon/poon-cli/internal/commands"
	"github.com/spf13/cobra"
)

// Execute is the main entrypoint for the CLI
func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "poon",
		Short: "Poon CLI - Internet-scale monorepo client",
		Long: `Poon CLI - A CLI tool for interacting with the Poon monorepo system via gRPC.

Poon enables working with internet-scale monorepos by providing:
- Partial checkout and sparse-checkout functionality
- Git-compatible workflow with familiar commands
- gRPC-based communication with the monorepo server
- Workspace management for tracking specific directories

Usage:
  poon [command]

Available Commands:
  start       Initialize a new poon workspace
  track       Track directories from the monorepo
  push        Push local changes back to the monorepo
  sync        Sync with latest monorepo state
  status      Show workspace status
  ls          List directory contents
  cat         Display file contents
  branches    List available branches
  workspace   Workspace management commands

Examples:
  poon start src/frontend           # Initialize workspace tracking src/frontend
  poon track src/backend docs       # Track additional directories
  poon status                       # Show current workspace status
  poon sync                         # Sync with latest monorepo changes
`,

		// Persistent flags
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set up logging, configuration, etc.
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().String("server", "localhost:50051", "gRPC server address")
	rootCmd.PersistentFlags().String("git-server", "localhost:3000", "Git server address")

	// Add all commands
	commands.AddCommands(rootCmd)

	return rootCmd.Execute()
}
