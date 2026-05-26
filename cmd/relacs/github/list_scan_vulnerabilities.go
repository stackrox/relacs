package github

import (
	"bytes"
	"context"
	"fmt"
	"sort"
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
	ShowCVEs     bool                    `json:"-" yaml:"-"`
	ShowAllCVEs  bool                    `json:"-" yaml:"-"`
	FixableOnly  bool                    `json:"-" yaml:"-"`
}

func (o ScanVulnerabilitiesOutput) String() string {
	output := fmt.Sprintf("Image vulnerability scans for %s/%s@%s\n", o.Organization, o.Repository, o.Tag)
	if len(o.ImageScans) == 0 {
		output += "  No scan results found.\n"
		return output
	}

	output += renderSummaryTable(o.ImageScans)

	if o.ShowCVEs || o.ShowAllCVEs {
		output += "\n"
		output += renderCVEList(o.ImageScans, o.ShowAllCVEs, o.FixableOnly)
	}

	return output
}

func renderSummaryTable(entries []github.ImageScanEntry) string {
	var buf bytes.Buffer
	headers := []string{"Source", "Image", "Result", "Critical", "Important", "Moderate", "Low", "Total"}
	table := tablewriter.NewTable(&buf, tablewriter.WithHeader(headers))
	for _, entry := range entries {
		//nolint:errcheck
		table.Append([]string{
			entry.Source,
			entry.ImageName,
			entry.Conclusion,
			fmt.Sprintf("%d", entry.Summary.Critical),
			fmt.Sprintf("%d", entry.Summary.Important),
			fmt.Sprintf("%d", entry.Summary.Moderate),
			fmt.Sprintf("%d", entry.Summary.Low),
			fmt.Sprintf("%d", entry.Summary.TotalVulnerabilities),
		})
	}
	//nolint:errcheck
	table.Render()
	return buf.String()
}

type cveAggregation struct {
	severity string
	cvss     float64
	fixable  bool
	imageSet map[string]bool
}

func renderCVEList(entries []github.ImageScanEntry, showAll, fixableOnly bool) string {
	cves := make(map[string]*cveAggregation)
	for _, entry := range entries {
		for _, v := range entry.Vulnerabilities {
			if !showAll && v.CVESeverity != "CRITICAL" && v.CVESeverity != "IMPORTANT" {
				continue
			}
			if fixableOnly && v.ComponentFixedVersion == "" {
				continue
			}
			agg, ok := cves[v.CVEId]
			if !ok {
				agg = &cveAggregation{
					severity: v.CVESeverity,
					cvss:     v.CVECVSS,
					fixable:  v.ComponentFixedVersion != "",
					imageSet: make(map[string]bool),
				}
				cves[v.CVEId] = agg
			}
			agg.imageSet[entry.Source+": "+entry.ImageName] = true
			if v.ComponentFixedVersion != "" {
				agg.fixable = true
			}
		}
	}

	if len(cves) == 0 {
		return "  No CVEs found.\n"
	}

	type cveRow struct {
		id       string
		severity string
		cvss     float64
		fixable  bool
		images   []string
	}
	var rows []cveRow
	for id, agg := range cves {
		var images []string
		for img := range agg.imageSet {
			images = append(images, img)
		}
		sort.Strings(images)
		rows = append(rows, cveRow{id: id, severity: agg.severity, cvss: agg.cvss, fixable: agg.fixable, images: images})
	}
	sort.Slice(rows, func(i, j int) bool {
		if severityRank(rows[i].severity) != severityRank(rows[j].severity) {
			return severityRank(rows[i].severity) > severityRank(rows[j].severity)
		}
		return rows[i].id < rows[j].id
	})

	var sb strings.Builder
	sb.WriteString("CVEs:\n")
	for _, r := range rows {
		fixable := ""
		if r.fixable {
			fixable = " [fixable]"
		}
		fmt.Fprintf(&sb, "\n  %s (%s, CVSS %.1f)%s\n", r.id, r.severity, r.cvss, fixable)
		for _, img := range r.images {
			fmt.Fprintf(&sb, "    - %s\n", img)
		}
	}
	return sb.String()
}

func severityRank(s string) int {
	switch s {
	case "CRITICAL":
		return 4
	case "IMPORTANT":
		return 3
	case "MODERATE":
		return 2
	case "LOW":
		return 1
	default:
		return 0
	}
}

var listScanVulnsCmd = &cobra.Command{
	Use:   "list-scan-vulns",
	Short: "List image vulnerability scan results for a given Git tag",
	Long:  `Retrieve vulnerability scan results from the GitHub Actions scan workflow for a specific Git tag`,
	RunE:  common.RunCommand(RunListScanVulnerabilities),
}

func init() {
	listScanVulnsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listScanVulnsCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	listScanVulnsCmd.Flags().String("tag", "", "Git tag (required)")
	listScanVulnsCmd.Flags().Bool("failed", false, "Only show failed scans")
	listScanVulnsCmd.Flags().Bool("cves", false, "Show critical and important CVEs with affected images")
	listScanVulnsCmd.Flags().Bool("all", false, "Show all CVEs (implies --cves)")
	listScanVulnsCmd.Flags().Bool("fixable", false, "Show only fixable CVEs (implies --cves)")

	//nolint:errcheck
	listScanVulnsCmd.MarkFlagRequired("tag")
}

func RunListScanVulnerabilities(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	failedOnly, _ := cmd.Flags().GetBool("failed")
	showCVEs, _ := cmd.Flags().GetBool("cves")
	showAllCVEs, _ := cmd.Flags().GetBool("all")
	fixableOnly, _ := cmd.Flags().GetBool("fixable")

	entries, err := github.ListScanVulnerabilities(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to list scan vulnerabilities: %w", err)
	}

	if failedOnly {
		var failed []github.ImageScanEntry
		for _, e := range entries {
			if e.Conclusion == "failure" {
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
		ShowCVEs:     showCVEs || fixableOnly,
		ShowAllCVEs:  showAllCVEs,
		FixableOnly:  fixableOnly,
	}, nil
}
