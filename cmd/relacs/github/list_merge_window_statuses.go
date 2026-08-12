package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	productlifecycle "github.com/stackrox/relacs/pkg/client/product_lifecycle"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

// ListMergeWindowStatusesOutput represents merge window statuses for all supported product streams.
type ListMergeWindowStatusesOutput struct {
	SupportedProductStreams []string          `json:"supported-product-streams" yaml:"supported-product-streams"`
	MergeWindowStatuses     map[string]string `json:"merge-window-statuses" yaml:"merge-window-statuses"`
}

func (c ListMergeWindowStatusesOutput) String() string {
	var buf strings.Builder
	buf.WriteString("# Merge Window Statuses\n\n")
	buf.WriteString("| Product Stream | Status |\n")
	buf.WriteString("| ------- | ------ |\n")

	for _, productStream := range c.SupportedProductStreams {
		fmt.Fprintf(&buf, "| %s | %s |\n", productStream, c.MergeWindowStatuses[productStream])
	}
	return buf.String()
}

var listMergeWindowsCmd = &cobra.Command{
	Use:   "list-merge-windows",
	Short: "List merge window statuses for all supported product streams",
	Long:  `List merge window statuses for all supported product streams in different formats`,
	RunE:  common.RunCommand(RunListMergeWindowStatuses),
}

func init() {
	listMergeWindowsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listMergeWindowsCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
}

// RunListMergeWindowStatuses retrieves and displays merge window statuses for all supported product streams.
func RunListMergeWindowStatuses(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	org, err := cmd.Flags().GetString("org")
	if err != nil {
		return nil, err
	}
	repo, err := cmd.Flags().GetString("repo")
	if err != nil {
		return nil, err
	}

	productLifecycleData, err := productlifecycle.FetchProductLifecycleDataForProduct(config.ProductName)
	if err != nil {
		return nil, err
	}

	supportedProductStreams := productlifecycle.GetSupportedProductStreams(productLifecycleData)
	mergeWindowStatuses := map[string]string{}

	for _, productStream := range supportedProductStreams {
		mergeWindowStatus, err := github.GetMergeWindowStatus(ctx, c, org, repo, productStream)
		if err != nil {
			return nil, err
		}
		mergeWindowStatuses[productStream] = mergeWindowStatus
	}

	return ListMergeWindowStatusesOutput{
		SupportedProductStreams: supportedProductStreams,
		MergeWindowStatuses:     mergeWindowStatuses,
	}, nil
}
