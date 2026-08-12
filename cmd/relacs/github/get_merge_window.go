package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

// GetMergeWindowOutput represents the merge window status for a single product stream.
type GetMergeWindowOutput struct {
	ProductStream string `json:"product-stream" yaml:"product-stream"`
	Status        string `json:"status" yaml:"status"`
}

func (c GetMergeWindowOutput) String() string {
	return fmt.Sprintf("%s: %s", c.ProductStream, c.Status)
}

var getMergeWindowCmd = &cobra.Command{
	Use:   "get-merge-window",
	Short: "Get the merge window status for a given product stream",
	Long:  `Get the merge window status for a given product stream`,
	RunE:  common.RunCommand(RunGetMergeWindow),
}

func init() {
	getMergeWindowCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	getMergeWindowCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	getMergeWindowCmd.Flags().String("product-stream", "", "Product stream (required)")

	//nolint:errcheck
	getMergeWindowCmd.MarkFlagRequired("product-stream")
}

// RunGetMergeWindow retrieves the merge window status for a specific product stream.
func RunGetMergeWindow(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	org, err := cmd.Flags().GetString("org")
	if err != nil {
		return nil, err
	}
	repo, err := cmd.Flags().GetString("repo")
	if err != nil {
		return nil, err
	}

	productStream, err := cmd.Flags().GetString("product-stream")
	if err != nil {
		return nil, err
	}

	mergeWindowStatus, err := github.GetMergeWindowStatus(ctx, c, org, repo, productStream)
	if err != nil {
		return nil, err
	}

	return GetMergeWindowOutput{
		ProductStream: productStream,
		Status:        mergeWindowStatus,
	}, nil
}
