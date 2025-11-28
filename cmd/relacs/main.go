package main

import (
	"fmt"
	"os"

	"github.com/stackrox/rel-acs/cmd/relacs/github"
)

func main() {
	// Add github command group to root
	rootCmd.AddCommand(github.GitHubCmd)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
