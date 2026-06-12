package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/cmd/relacs/konflux"
	"github.com/stackrox/relacs/cmd/relacs/version"
	"github.com/stackrox/relacs/pkg/versioncheck"
)

var (
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:           "relacs",
	Short:         "relacs automates the ACS release process",
	Long:          `relacs is a collection of utility commands to interact with GitHub, Konflux & co.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip version check for the version command itself to avoid circular dependency
		if cmd.Name() == "version" {
			return nil
		}

		// Get the current version from the version package
		currentVersion := version.GetVersion()

		// Check if current version meets minimum requirements
		return versioncheck.CheckMinimumVersion(currentVersion)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")
}

func main() {
	rootCmd.AddCommand(version.VersionCmd)
	rootCmd.AddCommand(github.GitHubCmd)
	rootCmd.AddCommand(konflux.KonfluxCmd)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
