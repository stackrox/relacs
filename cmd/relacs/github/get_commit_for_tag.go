package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/github"
)

type CommitOutput struct {
	Organization string `json:"organization" yaml:"organization"`
	Repository   string `json:"repository" yaml:"repository"`
	Tag          string `json:"tag" yaml:"tag"`
	Commit       string `json:"commit" yaml:"commit"`
}

func (c CommitOutput) String() string {
	return fmt.Sprintf("%s/%s@%s: %s", c.Organization, c.Repository, c.Tag, c.Commit)
}

var getCommitForTagCmd = &cobra.Command{
	Use:   "get-commit-for-tag",
	Short: "Get the commit SHA for a given Git tag",
	Long:  `Retrieve the commit SHA associated with a specific Git tag in a GitHub repository`,
	RunE:  common.RunCommand(runGetCommitForTag),
}

func init() {
	getCommitForTagCmd.Flags().String("org", "stackrox", "GitHub organization (optional)")
	getCommitForTagCmd.Flags().String("repo", "", "GitHub repository (required)")
	getCommitForTagCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getCommitForTagCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	getCommitForTagCmd.MarkFlagRequired("tag")
}

func runGetCommitForTag(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
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
		Commit:       *commit.SHA,
	}, nil
}
