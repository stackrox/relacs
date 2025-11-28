package konflux

import (
	"github.com/spf13/cobra"
)

// KonfluxCmd represents the konflux command group
var KonfluxCmd = &cobra.Command{
	Use:   "konflux",
	Short: "Konflux-related utilities",
	Long:  `Commands for interacting with Konflux build system and snapshots`,
}

func init() {
	// Subcommands will be added here
	KonfluxCmd.AddCommand(findCustomSnapshotForTagCmd)
}

// getOutputFormat retrieves the output format flag from the command
func getOutputFormat(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString("output")
}
