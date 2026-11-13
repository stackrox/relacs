package common

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/pkg/client"
)

type RunCommandFunc func(cmd *cobra.Command, args []string, c *client.Client) (any, error)

// PartialResultError marks a command result as incomplete: the output
// returned alongside this error is still rendered, but the command exits
// non-zero.
type PartialResultError struct {
	Err error
}

func (e *PartialResultError) Error() string { return e.Err.Error() }
func (e *PartialResultError) Unwrap() error { return e.Err }

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

		output, err := run(cmd, args, c)
		var partial *PartialResultError
		if errors.As(err, &partial) && output != nil {
			// Render the partial output before returning the error so the
			// command exits non-zero without losing what was collected.
			if formatErr := FormatOutput(outputFormat, output); formatErr != nil {
				return formatErr
			}
		}
		if err != nil {
			return err
		}

		return FormatOutput(outputFormat, output)
	}
}
