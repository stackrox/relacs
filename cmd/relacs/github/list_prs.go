package github

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

type PRsOutput struct {
	Organization  string                   `json:"organization" yaml:"organization"`
	Repository    string                   `json:"repository" yaml:"repository"`
	BaseBranch    string                   `json:"base-branch" yaml:"base-branch"`
	IncludeDrafts bool                     `json:"include-drafts" yaml:"include-drafts"`
	PRs           []github.PullRequestInfo `json:"prs" yaml:"prs"`
}

func (c PRsOutput) String() string {
	output := fmt.Sprintf("%s/%s@%s\n", c.Organization, c.Repository, c.BaseBranch)
	output += fmt.Sprintln("Open pull requests:")
	output += renderPRTable(c.PRs, c.IncludeDrafts)
	return output
}

var listPRsCmd = &cobra.Command{
	Use:   "list-prs",
	Short: "List the pull requests for a given repository and base branch",
	Long:  `Retrieve the pull requests associated with a specific repository and base branch in a GitHub repository`,
	RunE:  common.RunCommand(RunListPRs),
}

func init() {
	listPRsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listPRsCmd.Flags().String("repo", "", "GitHub repository (required)")
	listPRsCmd.Flags().String("base-branch", "", "Base branch (required)")
	listPRsCmd.Flags().Bool("include-drafts", false, "Include draft PRs (optional)")

	//nolint:errcheck
	listPRsCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	listPRsCmd.MarkFlagRequired("base-branch")
}

// RunListPRs is the main function of this command and exported for testing purposes.
func RunListPRs(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	baseBranch, _ := cmd.Flags().GetString("base-branch")
	includeDrafts, _ := cmd.Flags().GetBool("include-drafts")

	gh, err := c.GetGitHub()
	if err != nil {
		return nil, err
	}

	prs, err := github.ListPullRequests(ctx, gh, org, repo, baseBranch, includeDrafts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	return PRsOutput{
		Organization:  org,
		Repository:    repo,
		BaseBranch:    baseBranch,
		IncludeDrafts: includeDrafts,
		PRs:           prs,
	}, nil
}

func renderPRTable(prs []github.PullRequestInfo, includeDrafts bool) string {
	if len(prs) == 0 {
		return "  No open pull requests found."
	}

	var buf bytes.Buffer
	columnHeaders := getColumnHeaders(includeDrafts)
	table := tablewriter.NewTable(&buf, tablewriter.WithHeader(columnHeaders))

	for _, pr := range prs {
		row := getTableRow(pr, includeDrafts)
		//nolint:errcheck
		table.Append(row)
	}

	//nolint:errcheck
	table.Render()
	return buf.String()
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

func getChecksSymbol(status string) string {
	switch status {
	case "Passed":
		return "✓"
	case "Failed":
		return "✗"
	case "Pending":
		return "⋯"
	case "No Checks":
		return "-"
	default:
		return "?"
	}
}

func getColumnHeaders(includeDrafts bool) []string {
	headers := []string{"#", "Title", "Author", "Review", "Checks"}
	if includeDrafts {
		headers = append(headers, "Draft")
	}
	headers = append(headers, "URL")
	return headers
}

func getTableRow(pr github.PullRequestInfo, includeDrafts bool) []string {
	row := []string{
		strconv.Itoa(pr.Number),
		truncateString(pr.Title, 50),
		pr.Author,
		pr.ReviewStatus,
		getChecksSymbol(pr.ChecksStatus),
	}
	if includeDrafts {
		draft := ""
		if pr.IsDraft {
			draft = "yes"
		}
		row = append(row, draft)
	}
	return append(row, pr.URL)
}
