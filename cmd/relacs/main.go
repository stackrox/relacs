package main

import (
	"fmt"
	"os"

	"github.com/stackrox/rel-acs/cmd/relacs/github"
	"github.com/stackrox/rel-acs/cmd/relacs/konflux"
)

func main() {
	// Add command groups to root
	rootCmd.AddCommand(github.GitHubCmd)
	rootCmd.AddCommand(konflux.KonfluxCmd)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
