package main

import (
	"fmt"
	"os"

	"github.com/stackrox/rel-acs/cmd/relacs/git"
)

func main() {
	// Add git command group to root
	rootCmd.AddCommand(git.GitCmd)

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
