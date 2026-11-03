package github

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

type CommitOutput struct {
	Organization string            `json:"organization" yaml:"organization"`
	Repository   string            `json:"repository" yaml:"repository"`
	Tag          string            `json:"tag" yaml:"tag"`
	Commit       github.CommitInfo `json:"commit" yaml:"commit"`
	URL          string            `json:"url" yaml:"url"`
}

func (c CommitOutput) String() string {
	output := fmt.Sprintf("%s/%s@%s:\n", c.Organization, c.Repository, c.Tag)
	output += fmt.Sprintf("Commit: %s\n", c.Commit.SHA)
	output += fmt.Sprintf("URL: %s\n", c.URL)
	return output
}

var getCommitForTagCmd = &cobra.Command{
	Use:   "get-commit-for-tag",
	Short: "Get the commit SHA for a given Git tag",
	Long:  `Retrieve the commit SHA associated with a specific Git tag in a GitHub repository`,
	RunE:  common.RunCommand(RunGetCommitForTag),
}

func init() {
	getCommitForTagCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	getCommitForTagCmd.Flags().String("repo", "", "GitHub repository (required)")
	getCommitForTagCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getCommitForTagCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	getCommitForTagCmd.MarkFlagRequired("tag")
}

// RunGetCommitForTag is the main function of this command and exported for testing purposes.
func RunGetCommitForTag(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")

	// Get commit SHA
	commit, err := github.GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	return CommitOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		Commit:       commit,
		URL:          fmt.Sprintf("https://github.com/%s/%s/commit/%s", org, repo, commit.SHA),
	}, nil
}
