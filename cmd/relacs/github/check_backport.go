package github

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/github"
)

var releaseBranchPattern = regexp.MustCompile(`^release-(\d+)\.(\d+)$`)

const (
	konfluxBotUsername   = "red-hat-konflux[bot]"
	rhacsBotUsername     = "rhacs-bot"
	releasePendingStatus = "Release Pending"
)

type BackportCheckOutput struct {
	PullRequestURL string   `json:"pull-request-url" yaml:"pull-request-url"`
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors"`
}

func (o BackportCheckOutput) String() string {
	if o.Valid {
		return fmt.Sprintf("The PR '%s' satisfies the backport criteria.", o.PullRequestURL)
	}

	out := fmt.Sprintf("The PR '%s' does not satisfy the backport criteria:\n", o.PullRequestURL)
	for i, error := range o.Errors {
		out += fmt.Sprintf("  %d. %s\n", i+1, error)
	}
	return out
}

// BackportCheck represents the backport-check command
var backportCheckCmd = &cobra.Command{
	Use:   "backport-check",
	Short: "Check if a backport PR satisfies the criteria",
	Long:  `Check if a PR to a release branch links to a Jira with correct fixVersion and state.`,
	RunE:  common.RunCommand(RunBackportCheck),
}

func init() {
	backportCheckCmd.Flags().String("pull-request", "", "Pull Request URL to check (required)")

	//nolint:errcheck
	backportCheckCmd.MarkFlagRequired("pull-request")
}

func RunBackportCheck(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	//nolint:errcheck
	pullRequest, _ := cmd.Flags().GetString("pull-request")
	out := BackportCheckOutput{
		PullRequestURL: pullRequest,
		Valid:          false,
		Errors:         []string{},
	}

	pr, err := github.GetPullRequestFromURL(ctx, c, pullRequest)
	if err != nil {
		out.Errors = append(out.Errors, fmt.Sprintf("failed to get pull request: %v", err))
		return out, nil
	}

	// Skip checks for PRs by red-hat-konflux[bot] or rhacs-bot
	getPRAuthor := pr.GetUser().GetLogin()
	if getPRAuthor == konfluxBotUsername || getPRAuthor == rhacsBotUsername {
		out.Valid = true
		return out, nil
	}

	// Check if the PR targets a release branch (base is release-<major>.<minor>)
	baseBranch := pr.GetBase().GetRef()
	matches := releaseBranchPattern.FindStringSubmatch(baseBranch)
	if matches == nil {
		out.Errors = append(out.Errors, fmt.Sprintf("the PR does not target a valid release branch (expected release-<major>.<minor>): %s", baseBranch))
		return out, nil
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])

	// Check if Jira client is configured
	jiraClient := c.GetJira()
	if jiraClient == nil {
		out.Errors = append(out.Errors, "Jira client is not configured (JIRA_TOKEN environment variable is required)")
		return out, nil
	}

	expectedFixVersion, err := jiraClient.GetLowestUnreleasedFixVersion(major, minor)
	if err != nil {
		out.Errors = append(out.Errors, err.Error())
		return out, nil
	}

	prTitle := pr.GetTitle()
	// Extract all Jira issue keys from the PR title (format: "PROJ-123, PROJ-456: description" or "PROJ-123: description")
	issueKeyPattern := regexp.MustCompile(`[A-Z]+-[0-9]+`)
	issueKeys := issueKeyPattern.FindAllString(prTitle, -1)

	if len(issueKeys) == 0 {
		out.Errors = append(out.Errors, fmt.Sprintf("the PR title does not mention at least one Jira issue key: %s", prTitle))
		return out, nil
	}

	for _, issueKey := range issueKeys {
		issue, err := jiraClient.GetIssueFromKey(issueKey)
		if err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("failed to get Jira issue: %v", err))
			continue
		}

		status := issue.Fields.Status.Name
		if status != releasePendingStatus {
			out.Errors = append(out.Errors, fmt.Sprintf("the Jira issue %s is not in state '%s': %s", issueKey, releasePendingStatus, status))
		}

		// Check if the linked Jira all have the expected fixVersion
		fixVersions := issue.Fields.FixVersions
		found := false
		for _, fixVersion := range fixVersions {
			if fixVersion.Name == expectedFixVersion {
				found = true
				break
			}
		}
		if !found {
			out.Errors = append(out.Errors, fmt.Sprintf("the Jira issue %s does not have the expected fixVersion: %s", issueKey, expectedFixVersion))
		}
	}

	if len(out.Errors) > 0 {
		return out, nil
	}

	out.Valid = true
	return out, nil
}
