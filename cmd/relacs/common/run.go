package common

import (
	"context"

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
		if err != nil {
			return err
		}

		return FormatOutput(outputFormat, output)
	}
}
