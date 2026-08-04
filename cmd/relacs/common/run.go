package common

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/pkg/client"
)

type RunCommandFunc func(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error)

func RunCommand(run RunCommandFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Silence usage for errors that occur during command execution.
		// Flag validation errors (like missing required flags) happen before
		// RunE is called, so they will still show usage.
		cmd.SilenceUsage = true

		outputFormat, err := GetOutputFormat(cmd)
		if err != nil {
			return err
		}

		c, err := client.NewClient()
		if err != nil {
			return err
		}

		ctx := context.Background()
		output, err := run(cmd, args, c, ctx)

		// Format output even if there's an error, as long as output is not nil
		if output != nil {
			formatErr := FormatOutput(outputFormat, output)
			if formatErr != nil {
				return formatErr
			}
			// When we've formatted structured output with an error,
			// call os.Exit(1) to set exit code without printing error message
			// Tests call the underlying function directly and won't hit this path
			if err != nil {
				os.Exit(1)
			}
		}

		return err
	}
}
