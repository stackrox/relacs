package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

type OperatorIndexImageOutput struct {
	Organization          string `json:"organization" yaml:"organization"`
	Repository            string `json:"repository" yaml:"repository"`
	Tag                   string `json:"tag" yaml:"tag"`
	OperatorIndexImageURL string `json:"operator-index-image-url" yaml:"operator-index-image-url"`
	ImageDigest           string `json:"image-digest" yaml:"image-digest"`
}

func (o OperatorIndexImageOutput) String() string {
	output := fmt.Sprintf("Input: %s/%s@%s\n", o.Organization, o.Repository, o.Tag)
	output += fmt.Sprintf("Operator Index Image URL: %s\n", o.OperatorIndexImageURL)
	output += fmt.Sprintf("Image Digest: %s\n", o.ImageDigest)
	return output
}

var getOperatorIndexImageCmd = &cobra.Command{
	Use:   "get-operator-index-image",
	Short: "Get the operator index image for a given Git tag",
	Long:  `Retrieve the operator index image from the operator-index-on-push pipeline run for a specific Git tag`,
	RunE:  common.RunCommand(runGetOperatorIndexImage),
}

func init() {
	getOperatorIndexImageCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	getOperatorIndexImageCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	getOperatorIndexImageCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getOperatorIndexImageCmd.MarkFlagRequired("tag")
}

func runGetOperatorIndexImage(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")

	result, err := konflux.GetOperatorIndexImage(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator index image: %w", err)
	}

	return OperatorIndexImageOutput{
		Organization:          org,
		Repository:            repo,
		Tag:                   tag,
		OperatorIndexImageURL: result.ImageURL,
		ImageDigest:           result.ImageDigest,
	}, nil
}
