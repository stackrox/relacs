package konflux

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ConformaRunsOutput struct {
	Organization string                     `json:"organization" yaml:"organization"`
	Repository   string                     `json:"repository" yaml:"repository"`
	Tag          string                     `json:"tag" yaml:"tag"`
	ConformaRuns []konflux.ConformaRunEntry `json:"conformaRuns" yaml:"conformaRuns"`
}

func (o ConformaRunsOutput) String() string {
	output := fmt.Sprintf("Conforma runs for %s/%s@%s\n", o.Organization, o.Repository, o.Tag)
	output += renderConformaRunsTable(o.ConformaRuns)
	return output
}

func renderConformaRunsTable(runs []konflux.ConformaRunEntry) string {
	if len(runs) == 0 {
		return "  No conforma runs found."
	}

	var buf bytes.Buffer
	table := tablewriter.NewTable(&buf, tablewriter.WithHeader(getConformaRunColumnHeaders()))
	for _, entry := range runs {
		//nolint:errcheck
		table.Append(getConformaRunTableRow(entry))
	}
	//nolint:errcheck
	table.Render()
	return buf.String()
}

func getConformaRunColumnHeaders() []string {
	return []string{"Name", "Component", "Scenario", "Status", "URL"}
}

func getConformaRunTableRow(entry konflux.ConformaRunEntry) []string {
	return []string{
		entry.Name,
		entry.Component,
		entry.Scenario,
		entry.Status,
		entry.URL,
	}
}

var listConformaRunsCmd = &cobra.Command{
	Use:   "list-conforma-runs",
	Short: "List conforma pipeline runs for a given Git tag",
	Long:  `List conforma pipeline run results for a specific Git tag`,
	RunE:  common.RunCommand(runListConformaRuns),
}

func init() {
	listConformaRunsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listConformaRunsCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	listConformaRunsCmd.Flags().String("tag", "", "Git tag (required)")
	listConformaRunsCmd.Flags().Bool("failed", false, "Only show failed runs")

	//nolint:errcheck
	listConformaRunsCmd.MarkFlagRequired("tag")
}

func runListConformaRuns(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	failedOnly, _ := cmd.Flags().GetBool("failed")

	entries, err := konflux.ListConformaRuns(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to list conforma runs: %w", err)
	}

	if failedOnly {
		var failed []konflux.ConformaRunEntry
		for _, e := range entries {
			if e.Status != "Succeeded" && e.Status != "Running" {
				failed = append(failed, e)
			}
		}
		entries = failed
	}

	return ConformaRunsOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		ConformaRuns: entries,
	}, nil
}
