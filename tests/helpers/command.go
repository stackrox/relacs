package helpers

import (
	"github.com/spf13/cobra"
)

// NewMockCommand creates a mock cobra.Command for testing.
// It sets up the command with flags and allows setting flag values.
func NewMockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
	}
	return cmd
}

// SetFlag sets a flag value on a command for testing.
func SetFlag(cmd *cobra.Command, name, value string) error {
	cmd.Flags().String(name, "", "")
	return cmd.Flags().Set(name, value)
}
