package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/cmd/relacs/konflux"
	"github.com/stackrox/relacs/cmd/relacs/selftest"
)

var (
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "relacs",
	Short: "relacs automates the ACS release process",
	Long:  `relacs is a collection of utility commands to interact with GitHub, Konflux & co.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")
}

func main() {
	// Add command groups to root
	rootCmd.AddCommand(github.GitHubCmd)
	rootCmd.AddCommand(konflux.KonfluxCmd)
	rootCmd.AddCommand(selftest.SelfTestCmd)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
