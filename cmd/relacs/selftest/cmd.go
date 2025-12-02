package selftest

import (
	"github.com/spf13/cobra"
)

// SelfTestCmd represents the self-test command group
var SelfTestCmd = &cobra.Command{
	Use:   "self-test",
	Short: "Self-test utilities to verify release preparation setup",
	Long:  `Commands to check if the current user has the necessary setup to prepare a release`,
}

func init() {
	// Subcommands will be added here
	SelfTestCmd.AddCommand(checkSetupCmd)
}
