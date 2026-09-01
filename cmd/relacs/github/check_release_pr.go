package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/github"
)

const (
	konfluxBotUsername   = "red-hat-konflux[bot]"
	rhacsBotUsername     = "rhacs-bot"
	resolutionDone       = "Done"
	statusClosed         = "Closed"
	statusReleasePending = "Release Pending"
)

type ReleasePRCheckOutput struct {
	PullRequestURL string   `json:"pull-request-url" yaml:"pull-request-url"`
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors"`
}

func (o ReleasePRCheckOutput) String() string {
	if o.Valid {
		return fmt.Sprintf("The PR '%s' satisfies the release branch PR criteria.", o.PullRequestURL)
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "The PR '%s' does not satisfy the release branch PR criteria:\n", o.PullRequestURL)
	for i, err := range o.Errors {
		fmt.Fprintf(&buf, "  %d. %s\n", i+1, err)
	}
	return buf.String()
}

// checkReleasePRCmd represents the check-release-pr command
var checkReleasePRCmd = &cobra.Command{
	Use:   "check-release-pr",
	Short: "Check if a release branch PR satisfies the criteria",
	Long:  `Check if a PR to a release-* branch links to a Jira with correct fixVersion and state.`,
	RunE:  runCheckReleasePRCommand,
}

func init() {
	checkReleasePRCmd.Flags().String("pull-request", "", "Pull Request URL to check (required)")

	//nolint:errcheck
	checkReleasePRCmd.MarkFlagRequired("pull-request")
}

func runCheckReleasePRCommand(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	outputFormat, err := common.GetOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	output, err := RunReleasePRCheck(cmd, args, c, ctx)

	// Always format output (even on error) since it contains validation errors
	formatErr := common.FormatOutput(outputFormat, output)
	if formatErr != nil {
		return formatErr
	}

	// If there was a validation error, we've already shown the errors
	// in the structured output. Silence cobra's error message but still
	// return an error to get exit code 1.
	if err != nil {
		cmd.SilenceErrors = true
		return err
	}

	return nil
}

func RunReleasePRCheck(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	//nolint:errcheck
	pullRequest, _ := cmd.Flags().GetString("pull-request")
	out := ReleasePRCheckOutput{
		PullRequestURL: pullRequest,
		Valid:          false,
		Errors:         []string{},
	}

	pr, err := github.GetPullRequestFromURL(ctx, c, pullRequest)
	if err != nil {
		out.Errors = append(out.Errors, fmt.Sprintf("failed to get pull request: %v", err))
		return out, fmt.Errorf("release branch PR check failed")
	}

	// Skip checks for PRs by red-hat-konflux[bot]
	// For rhacs-bot, only skip if the PR doesn't have a "backport" label
	getPRAuthor := pr.GetUser().GetLogin()
	if getPRAuthor == konfluxBotUsername {
		out.Valid = true
		return out, nil
	}
	if getPRAuthor == rhacsBotUsername {
		hasBackportLabel := false
		for _, label := range pr.Labels {
			if label.GetName() == "backport" {
				hasBackportLabel = true
				break
			}
		}
		if !hasBackportLabel {
			out.Valid = true
			return out, nil
		}
		// If rhacs-bot PR has "backport" label, continue with normal checks
	}

	// Check if the PR targets a release branch (base is release-<major>.<minor>)
	baseBranch := pr.GetBase().GetRef()
	matches := releaseBranchPattern.FindStringSubmatch(baseBranch)
	if matches == nil {
		out.Errors = append(out.Errors, fmt.Sprintf("the PR does not target a valid release branch (expected release-<major>.<minor>): %s", baseBranch))
		return out, fmt.Errorf("release branch PR check failed")
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])

	// Check if Jira client is configured
	jiraClient := c.GetJira()
	if jiraClient == nil {
		out.Errors = append(out.Errors, "Jira client is not configured (JIRA_TOKEN environment variable is required)")
		return out, fmt.Errorf("release branch PR check failed")
	}

	expectedFixVersion, err := jiraClient.GetLowestUnreleasedFixVersion(major, minor)
	if err != nil {
		out.Errors = append(out.Errors, err.Error())
		return out, fmt.Errorf("release branch PR check failed")
	}

	prTitle := pr.GetTitle()
	// Extract all Jira issue keys from the PR title (format: "PROJ-123, PROJ-456: description" or "PROJ-123: description")
	issueKeys := issueKeyPattern.FindAllString(prTitle, -1)

	if len(issueKeys) == 0 {
		out.Errors = append(out.Errors, fmt.Sprintf("the PR title does not mention at least one Jira issue key: %s", prTitle))
		return out, fmt.Errorf("release branch PR check failed")
	}

	for _, issueKey := range issueKeys {
		issue, err := jiraClient.GetIssueFromKey(issueKey)
		if err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("failed to get Jira issue: %v", err))
			continue
		}

		status := issue.Fields.Status.Name
		resolution := issue.Fields.Resolution.Name
		if status != statusReleasePending && (status != statusClosed || resolution != resolutionDone) {
			out.Errors = append(out.Errors,
				fmt.Sprintf(
					"the Jira issue %s should have status '%s' or status '%s' with resolution '%s', but it has status '%s' and resolution '%s'",
					issueKey, statusReleasePending, statusClosed, resolutionDone, status, resolution,
				),
			)
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
			out.Errors = append(out.Errors, fmt.Sprintf("the Jira issue %s does not have the expected fixVersion '%s'", issueKey, expectedFixVersion))
		}
	}

	if len(out.Errors) > 0 {
		return out, fmt.Errorf("release branch PR check failed")
	}

	out.Valid = true
	return out, nil
}
