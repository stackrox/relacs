package github

import (
	"bytes"
	"context"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

// ListMergeWindowsOutput represents merge window statuses for all supported product streams.
type ListMergeWindowsOutput struct {
	SupportedProductStreams []string          `json:"supported-product-streams" yaml:"supported-product-streams"`
	MergeWindowStatuses     map[string]string `json:"merge-window-statuses" yaml:"merge-window-statuses"`
}

func (c ListMergeWindowsOutput) String() string {
	var buf strings.Builder
	buf.WriteString("# Merge Window Statuses\n\n")
	buf.WriteString(renderMergeWindowStatusesTable(c))
	return buf.String()
}

func renderMergeWindowStatusesTable(output ListMergeWindowsOutput) string {
	if len(output.SupportedProductStreams) == 0 {
		return "  No supported product streams found.\n"
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf,
		tablewriter.WithHeader([]string{"Product Stream", "Status"}),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)

	for _, productStream := range output.SupportedProductStreams {
		//nolint:errcheck
		table.Append([]string{productStream, output.MergeWindowStatuses[productStream]})
	}

	//nolint:errcheck
	table.Render()
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

	// Get product streams from Konflux applications
	supportedProductStreams, err := c.GetKonflux().GetActiveProductStreamsFromApplications()
	if err != nil {
		return nil, err
	}

	mergeWindowStatuses := map[string]string{}

	for _, productStream := range supportedProductStreams {
		mergeWindowStatus, err := github.GetMergeWindowStatus(ctx, c, org, repo, productStream)
		if err != nil {
			return nil, err
		}
		mergeWindowStatuses[productStream] = mergeWindowStatus
	}

	return ListMergeWindowsOutput{
		SupportedProductStreams: supportedProductStreams,
		MergeWindowStatuses:     mergeWindowStatuses,
	}, nil
}
