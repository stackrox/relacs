package github

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

type ScanVulnerabilitiesOutput struct {
	Organization string                  `json:"organization" yaml:"organization"`
	Repository   string                  `json:"repository" yaml:"repository"`
	Tag          string                  `json:"tag" yaml:"tag"`
	ImageScans   []github.ImageScanEntry `json:"imageScans" yaml:"imageScans"`
	FixableOnly  bool                    `json:"-" yaml:"-"`
	Severities   map[string]bool         `json:"-" yaml:"-"`
}

func (o ScanVulnerabilitiesOutput) String() string {
	output := fmt.Sprintf("Image vulnerability scans for %s/%s@%s\n", o.Organization, o.Repository, o.Tag)
	if len(o.ImageScans) == 0 {
		output += "  No scan results found.\n"
		return output
	}

	output += renderSummaryTable(o.ImageScans, o.Severities, o.FixableOnly)
	return output
}

func renderSummaryTable(entries []github.ImageScanEntry, severities map[string]bool, fixableOnly bool) string {
	var buf bytes.Buffer
	headers := []string{"Source", "Image", "Result", "Critical", "Important", "Moderate", "Low", "Total"}
	table := tablewriter.NewTable(&buf, tablewriter.WithHeader(headers))
	for _, entry := range entries {
		summary := countFilteredVulns(entry.Vulnerabilities, severities, fixableOnly)
		result := entry.Conclusion
		if result == "success" {
			result = "passed"
		}
		//nolint:errcheck
		table.Append([]string{
			entry.Source,
			entry.ImageName,
			result,
			fmt.Sprintf("%d", summary.Critical),
			fmt.Sprintf("%d", summary.Important),
			fmt.Sprintf("%d", summary.Moderate),
			fmt.Sprintf("%d", summary.Low),
			fmt.Sprintf("%d", summary.TotalVulnerabilities),
		})
	}
	//nolint:errcheck
	table.Render()
	return buf.String()
}

func countFilteredVulns(vulns []github.ScanVulnerability, severities map[string]bool, fixableOnly bool) github.ScanSummary {
	var s github.ScanSummary
	seen := make(map[string]bool)
	for _, v := range vulns {
		if seen[v.CVEId] {
			continue
		}
		if fixableOnly && v.ComponentFixedVersion == "" {
			continue
		}
		if len(severities) > 0 && !severities[v.CVESeverity] {
			continue
		}
		seen[v.CVEId] = true
		switch v.CVESeverity {
		case "CRITICAL":
			s.Critical++
		case "IMPORTANT":
			s.Important++
		case "MODERATE":
			s.Moderate++
		case "LOW":
			s.Low++
		}
		s.TotalVulnerabilities++
	}
	return s
}

var listVulnsByImageCmd = &cobra.Command{
	Use:   "list-vulns-by-image",
	Short: "List image vulnerability scan results for a given Git tag",
	Long:  `Retrieve vulnerability scan results from the GitHub Actions scan workflow for a specific Git tag`,
	RunE:  common.RunCommand(RunListScanVulnerabilities),
}

func init() {
	listVulnsByImageCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listVulnsByImageCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	listVulnsByImageCmd.Flags().String("tag", "", "Git tag (required)")
	listVulnsByImageCmd.Flags().Bool("failed", false, "Only show failed scans")
	listVulnsByImageCmd.Flags().Bool("fixable", false, "Show only fixable vulnerability counts")
	listVulnsByImageCmd.Flags().StringSlice("severity", nil, "Filter counts by severities (e.g. CRITICAL,IMPORTANT)")

	//nolint:errcheck
	listVulnsByImageCmd.MarkFlagRequired("tag")
}

func RunListScanVulnerabilities(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	failedOnly, _ := cmd.Flags().GetBool("failed")
	fixableOnly, _ := cmd.Flags().GetBool("fixable")
	severityList, _ := cmd.Flags().GetStringSlice("severity")

	severities := make(map[string]bool)
	for _, s := range severityList {
		severities[strings.ToUpper(s)] = true
	}

	entries, err := github.ListScanVulnerabilities(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to list scan vulnerabilities: %w", err)
	}

	if failedOnly {
		var failed []github.ImageScanEntry
		for _, e := range entries {
			if e.Conclusion != "success" {
				failed = append(failed, e)
			}
		}
		entries = failed
	}

	return ScanVulnerabilitiesOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		ImageScans:   entries,
		FixableOnly:  fixableOnly,
		Severities:   severities,
	}, nil
}
