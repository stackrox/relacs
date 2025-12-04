package konflux

import (
	"github.com/spf13/cobra"
)

// KonfluxCmd represents the konflux command group
var KonfluxCmd = &cobra.Command{
	Use:   "konflux",
	Short: "Konflux-related utilities",
	Long:  `Commands for interacting with Konflux build system`,
}

func init() {
	// Subcommands will be added here
	KonfluxCmd.AddCommand(getCustomSnapshotCmd)
}
