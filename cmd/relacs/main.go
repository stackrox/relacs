package main

import (
	"fmt"
	"os"

	"github.com/stackrox/rel-acs/cmd/relacs/github"
	"github.com/stackrox/rel-acs/cmd/relacs/konflux"
	"github.com/stackrox/rel-acs/cmd/relacs/selftest"
)

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
