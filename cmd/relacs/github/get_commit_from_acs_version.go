package github

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

type CommitFromACSVersionOutput struct {
	Organization string            `json:"organization" yaml:"organization"`
	Repository   string            `json:"repository" yaml:"repository"`
	ACSVersion   string            `json:"acs-version" yaml:"acs-version"`
	Commit       github.CommitInfo `json:"commit" yaml:"commit"`
	URL          string            `json:"url" yaml:"url"`
}

func (c CommitFromACSVersionOutput) String() string {
	output := fmt.Sprintf("%s/%s@%s:\n", c.Organization, c.Repository, c.ACSVersion)
	output += fmt.Sprintf("Commit: %s\n", c.Commit.SHA)
	output += fmt.Sprintf("URL: %s\n", c.URL)
	return output
}

var getCommitFromACSTagCmd = &cobra.Command{
	Use:   "get-commit-from-acs-version",
	Short: "Get the commit SHA for a given ACS version (from make tag)",
	Long:  `Retrieve the commit SHA associated with a specific ACS version (from make tag) in a GitHub repository`,
	RunE:  common.RunCommand(RunGetCommitFromACSVersion),
}

func init() {
	getCommitFromACSTagCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	getCommitFromACSTagCmd.Flags().String("repo", "", "GitHub repository (required)")
	getCommitFromACSTagCmd.Flags().String("acs-version", "", "ACS version (required)")

	//nolint:errcheck
	getCommitFromACSTagCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	getCommitFromACSTagCmd.MarkFlagRequired("acs-version")
}

// RunGetCommitFromACSVersion is the main function of this command and exported for testing purposes.
func RunGetCommitFromACSVersion(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	acsVersion, _ := cmd.Flags().GetString("acs-version")

	// Get commit SHA
	commit, err := github.GetCommitFromACSVersion(ctx, c, org, repo, acsVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit from ACS version: %w", err)
	}

	return CommitFromACSVersionOutput{
		Organization: org,
		Repository:   repo,
		ACSVersion:   acsVersion,
		Commit:       commit,
		URL:          fmt.Sprintf("https://github.com/%s/%s/commit/%s", org, repo, commit.SHA),
	}, nil
}
