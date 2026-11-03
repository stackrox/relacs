package helpers

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"
)

// NewMockCommand creates a mock cobra.Command for testing.
// It sets up the command with flags and allows setting flag values.
func NewMockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.SetContext(context.Background())
	return cmd
}

// SetFlag sets a flag value on a command for testing.
func SetFlag(cmd *cobra.Command, name, value string) error {
	cmd.Flags().String(name, "", "")
	return cmd.Flags().Set(name, value)
}

// SetBoolFlag sets a boolean flag value on a command for testing.
func SetBoolFlag(cmd *cobra.Command, name string, value bool) error {
	cmd.Flags().Bool(name, false, "")
	return cmd.Flags().Set(name, strconv.FormatBool(value))
}
