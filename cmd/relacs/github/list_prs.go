package github

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/stackrox/rel-acs/pkg/client"
	ghpkg "github.com/stackrox/rel-acs/pkg/github"
)

var listReleasePRsCmd = &cobra.Command{
	Use:   "list-prs",
	Short: "List open pull requests targeting a release branch",
	Long:  `Retrieve and display all open pull requests targeting a specific release branch with their review status`,
	RunE:  runListReleasePRs,
}

func init() {
	listReleasePRsCmd.Flags().String("org", "", "GitHub organization (required)")
	listReleasePRsCmd.Flags().String("repo", "", "GitHub repository (required)")
	listReleasePRsCmd.Flags().String("branch", "", "Target branch name (required)")
	listReleasePRsCmd.Flags().Bool("include-drafts", false, "Include draft pull requests in the output")

	//nolint:errcheck
	listReleasePRsCmd.MarkFlagRequired("org")
	//nolint:errcheck
	listReleasePRsCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	listReleasePRsCmd.MarkFlagRequired("branch")
}

type PRListOutput struct {
	Organization string                  `json:"organization" yaml:"organization"`
	Repository   string                  `json:"repository" yaml:"repository"`
	Branch       string                  `json:"branch" yaml:"branch"`
	PullRequests []ghpkg.PullRequestInfo `json:"pull_requests" yaml:"pull_requests"`
}

func runListReleasePRs(cmd *cobra.Command, args []string) error {
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	branch, _ := cmd.Flags().GetString("branch")
	includeDrafts, _ := cmd.Flags().GetBool("include-drafts")

	outputFormat, err := getOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	// Get pull requests
	prs, err := ghpkg.ListPullRequestsForBranch(context.Background(), c, org, repo, branch)
	if err != nil {
		return fmt.Errorf("failed to list pull requests: %w", err)
	}

	// Filter out drafts if not requested
	if !includeDrafts {
		filtered := make([]ghpkg.PullRequestInfo, 0)
		for _, pr := range prs {
			if !pr.IsDraft {
				filtered = append(filtered, pr)
			}
		}
		prs = filtered
	}

	// Sort by PR number ascending
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].Number < prs[j].Number
	})

	// Handle empty results
	if len(prs) == 0 {
		if outputFormat == "text" {
			fmt.Printf("No open pull requests found targeting branch '%s'\n", branch)
			return nil
		}
		// For JSON/YAML, still output the structure with empty array
		return formatOutput(outputFormat, PRListOutput{
			Organization: org,
			Repository:   repo,
			Branch:       branch,
			PullRequests: []ghpkg.PullRequestInfo{},
		})
	}

	// Format output
	if outputFormat == "text" {
		printPRTable(prs, includeDrafts)
		return nil
	}

	return formatOutput(outputFormat, PRListOutput{
		Organization: org,
		Repository:   repo,
		Branch:       branch,
		PullRequests: prs,
	})
}

func printPRTable(prs []ghpkg.PullRequestInfo, showDraftColumn bool) {
	// Calculate column widths
	const (
		minTitleWidth  = 50
		minAuthorWidth = 15
		minReviewWidth = 18
		minChecksWidth = 6
		minDraftWidth  = 5
	)

	titleWidth := minTitleWidth
	authorWidth := minAuthorWidth
	reviewWidth := minReviewWidth
	checksWidth := minChecksWidth

	for _, pr := range prs {
		if len(pr.Title) > titleWidth {
			titleWidth = len(pr.Title)
		}
		if len(pr.Author) > authorWidth {
			authorWidth = len(pr.Author)
		}
		if len(pr.ReviewStatus) > reviewWidth {
			reviewWidth = len(pr.ReviewStatus)
		}
		if len(pr.ChecksStatus) > checksWidth {
			checksWidth = len(pr.ChecksStatus)
		}
	}

	// Limit maximum widths for readability
	if titleWidth > 80 {
		titleWidth = 80
	}

	// Print header
	if showDraftColumn {
		headerFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n", titleWidth, authorWidth, reviewWidth, checksWidth, minDraftWidth)
		fmt.Printf(headerFormat, "Title", "Author", "Review Status", "Checks", "Draft", "URL")
	} else {
		headerFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n", titleWidth, authorWidth, reviewWidth, checksWidth)
		fmt.Printf(headerFormat, "Title", "Author", "Review Status", "Checks", "URL")
	}

	// Print separator
	separator := ""
	separatorLength := titleWidth + authorWidth + reviewWidth + checksWidth + 60
	if showDraftColumn {
		separatorLength += minDraftWidth
	}
	for i := 0; i < separatorLength; i++ {
		separator += "-"
	}
	fmt.Println(separator)

	// Print rows
	for _, pr := range prs {
		title := pr.Title
		if len(title) > titleWidth {
			title = title[:titleWidth-3] + "..."
		}

		// Convert checks status to symbols for text output
		checksSymbol := getChecksSymbol(pr.ChecksStatus)

		if showDraftColumn {
			draft := "no"
			if pr.IsDraft {
				draft = "yes"
			}
			rowFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n", titleWidth, authorWidth, reviewWidth, checksWidth, minDraftWidth)
			fmt.Printf(rowFormat, title, pr.Author, pr.ReviewStatus, checksSymbol, draft, pr.URL)
		} else {
			rowFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n", titleWidth, authorWidth, reviewWidth, checksWidth)
			fmt.Printf(rowFormat, title, pr.Author, pr.ReviewStatus, checksSymbol, pr.URL)
		}
	}
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
