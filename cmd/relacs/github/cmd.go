package github

import (
	"regexp"

	"github.com/spf13/cobra"
)

var (
	releaseBranchPattern = regexp.MustCompile(`^release-(\d+)\.(\d+)$`)
	issueKeyPattern      = regexp.MustCompile(`[A-Z]+-[0-9]+`)
)

// GitHubCmd represents the github command group
var GitHubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub-related utilities",
	Long:  `Commands for interacting with GitHub repositories, teams, and organizations`,
}

func init() {
	// Subcommands are added here
	GitHubCmd.AddCommand(addTeamMemberCmd)
	GitHubCmd.AddCommand(backportCheckCmd)
	GitHubCmd.AddCommand(getCommitForTagCmd)
	GitHubCmd.AddCommand(getCommitFromACSTagCmd)
	GitHubCmd.AddCommand(listPRsCmd)
	GitHubCmd.AddCommand(listVulnsByCVECmd)
	GitHubCmd.AddCommand(listVulnsByImageCmd)
}
