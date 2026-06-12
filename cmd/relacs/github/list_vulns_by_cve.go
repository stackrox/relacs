package github

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

type ScanCVEsOutput struct {
	Organization string   `json:"organization" yaml:"organization"`
	Repository   string   `json:"repository" yaml:"repository"`
	Tag          string   `json:"tag" yaml:"tag"`
	CVEs         []cveRow `json:"cves" yaml:"cves"`
	ScanErrors   []string `json:"scanErrors,omitempty" yaml:"scanErrors,omitempty"`
}

func (o ScanCVEsOutput) String() string {
	output := fmt.Sprintf("CVEs for %s/%s@%s\n", o.Organization, o.Repository, o.Tag)
	if len(o.CVEs) == 0 {
		output += "  No CVEs found.\n"
	} else {
		var sb strings.Builder
		for _, r := range o.CVEs {
			fixable := ""
			if r.Fixable {
				fixable = " [fixable]"
			}
			fmt.Fprintf(&sb, "\n  %s (%s, CVSS %.1f)%s\n", r.ID, r.Severity, r.CVSS, fixable)
			for _, img := range r.Images {
				fmt.Fprintf(&sb, "    - %s\n", img)
			}
		}
		output += sb.String()
	}

	if len(o.ScanErrors) > 0 {
		output += "\nWARNING: The following images had scan errors (no scan data available):\n"
		for _, img := range o.ScanErrors {
			output += fmt.Sprintf("  - %s\n", img)
		}
	}
	return output
}

type cveRow struct {
	ID       string   `json:"id" yaml:"id"`
	Severity string   `json:"severity" yaml:"severity"`
	CVSS     float64  `json:"cvss" yaml:"cvss"`
	Fixable  bool     `json:"fixable" yaml:"fixable"`
	Images   []string `json:"images" yaml:"images"`
}

type cveAggregation struct {
	severity string
	cvss     float64
	fixable  bool
	imageSet map[string]bool
}

func aggregateCVEs(entries []github.ImageScanEntry, severities map[string]bool, fixableOnly bool, cveFilter map[string]bool) []cveRow {
	cves := make(map[string]*cveAggregation)
	for _, entry := range entries {
		for _, v := range entry.Vulnerabilities {
			if !severities[v.CVESeverity] {
				continue
			}
			if fixableOnly && v.ComponentFixedVersion == "" {
				continue
			}
			if len(cveFilter) > 0 && !cveFilter[v.CVEId] {
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
			imageName := entry.Source + ": " + entry.ImageName
			if entry.Conclusion == "failure" {
				imageName += " (failed)"
			}
			agg.imageSet[imageName] = true
			if v.ComponentFixedVersion != "" {
				agg.fixable = true
			}
		}
	}

	var rows []cveRow
	for id, agg := range cves {
		var images []string
		for img := range agg.imageSet {
			images = append(images, img)
		}
		sort.Strings(images)
		rows = append(rows, cveRow{ID: id, Severity: agg.severity, CVSS: agg.cvss, Fixable: agg.fixable, Images: images})
	}
	sort.Slice(rows, func(i, j int) bool {
		if severityRank(rows[i].Severity) != severityRank(rows[j].Severity) {
			return severityRank(rows[i].Severity) > severityRank(rows[j].Severity)
		}
		return rows[i].ID < rows[j].ID
	})
	return rows
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

var listVulnsByCVECmd = &cobra.Command{
	Use:   "list-vulns-by-cve",
	Short: "List CVEs from image vulnerability scans for a given Git tag",
	Long:  `List individual CVEs from the GitHub Actions scan workflow, grouped by CVE with affected images`,
	RunE:  common.RunCommand(RunListScanCVEs),
}

func init() {
	listVulnsByCVECmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listVulnsByCVECmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	listVulnsByCVECmd.Flags().String("tag", "", "Git tag (required)")
	listVulnsByCVECmd.Flags().StringSlice("severity", []string{"CRITICAL", "IMPORTANT"}, "Severities to include")
	listVulnsByCVECmd.Flags().Bool("fixable", false, "Show only fixable CVEs")
	listVulnsByCVECmd.Flags().StringSlice("cve", nil, "Filter by specific CVE IDs")

	//nolint:errcheck
	listVulnsByCVECmd.MarkFlagRequired("tag")
}

func RunListScanCVEs(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	severityList, _ := cmd.Flags().GetStringSlice("severity")
	fixableOnly, _ := cmd.Flags().GetBool("fixable")
	cveIDs, _ := cmd.Flags().GetStringSlice("cve")

	entries, err := github.ListScanVulnerabilities(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to list scan vulnerabilities: %w", err)
	}

	severities := make(map[string]bool)
	for _, s := range severityList {
		severities[strings.ToUpper(s)] = true
	}

	cveFilter := make(map[string]bool)
	for _, id := range cveIDs {
		cveFilter[id] = true
	}

	rows := aggregateCVEs(entries, severities, fixableOnly, cveFilter)

	var scanErrors []string
	for _, e := range entries {
		if e.Conclusion == "scan error" {
			scanErrors = append(scanErrors, e.Source+": "+e.ImageName)
		}
	}

	return ScanCVEsOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		CVEs:         rows,
		ScanErrors:   scanErrors,
	}, nil
}
