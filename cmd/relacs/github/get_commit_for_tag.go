package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
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
	getCommitForTagCmd.Flags().String("org", "", "GitHub organization (required)")
	getCommitForTagCmd.Flags().String("repo", "", "GitHub repository (required)")
	getCommitForTagCmd.Flags().String("tag", "", "Git tag (required)")

	getCommitForTagCmd.MarkFlagRequired("org")
	getCommitForTagCmd.MarkFlagRequired("repo")
	getCommitForTagCmd.MarkFlagRequired("tag")
}

func runGetCommitForTag(cmd *cobra.Command, args []string) error {
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")

	outputFormat, err := getOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := createClient()
	if err != nil {
		return err
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
