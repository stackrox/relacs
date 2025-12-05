package github

import (
	"github.com/spf13/cobra"
)

// GitHubCmd represents the github command group
var GitHubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub-related utilities",
	Long:  `Commands for interacting with GitHub repositories, teams, and organizations`,
}

func init() {
	// Subcommands are added here
	GitHubCmd.AddCommand(getCommitForTagCmd)
	GitHubCmd.AddCommand(listPRsCmd)
}
