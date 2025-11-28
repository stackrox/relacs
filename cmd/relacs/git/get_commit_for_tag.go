package git

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/rel-acs/pkg/client"
)

var (
	org  string
	repo string
	tag  string
)

type CommitOutput struct {
	Organization string `json:"organization" yaml:"organization"`
	Repository   string `json:"repository" yaml:"repository"`
	Tag          string `json:"tag" yaml:"tag"`
	Commit       string `json:"commit" yaml:"commit"`
}

var getCommitForTagCmd = &cobra.Command{
	Use:   "get-commit-for-tag",
	Short: "Get the commit SHA for a given Git tag",
	Long:  `Retrieve the commit SHA associated with a specific Git tag in a GitHub repository`,
	RunE:  runGetCommitForTag,
}

func init() {
	getCommitForTagCmd.Flags().StringVar(&org, "org", "", "GitHub organization (required)")
	getCommitForTagCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository (required)")
	getCommitForTagCmd.Flags().StringVar(&tag, "tag", "", "Git tag (required)")

	getCommitForTagCmd.MarkFlagRequired("org")
	getCommitForTagCmd.MarkFlagRequired("repo")
	getCommitForTagCmd.MarkFlagRequired("tag")
}

func runGetCommitForTag(cmd *cobra.Command, args []string) error {
	// Get the output format from the parent command's persistent flags
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("failed to get output format: %w", err)
	}

	// Create client
	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get commit SHA
	commit, err := c.GetCommitForTag(context.Background(), org, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to get commit for tag: %w", err)
	}

	// Format output
	return formatOutput(outputFormat, CommitOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		Commit:       commit,
	})
}
