package github

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

const (
	konfluxBotUsername   = "red-hat-konflux[bot]"
	rhacsBotUsername     = "rhacs-bot"
	resolutionDone       = "Done"
	statusClosed         = "Closed"
	statusReleasePending = "Release Pending"
	unrecoverableError   = "release branch PR check failed"
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
	checkReleasePRCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization to use")
	checkReleasePRCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository to use")
	checkReleasePRCmd.Flags().Int("pull-request", 0, "Pull Request number to check (required)")

	//nolint:errcheck
	checkReleasePRCmd.MarkFlagRequired("pull-request")
}

func runCheckReleasePRCommand(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	outputFormat, err := common.GetOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	output, err := RunReleasePRCheck(cmd, c)
	if err != nil {
		return err
	}

	// Always format output (even on error) since it contains validation errors
	formatErr := common.FormatOutput(outputFormat, output)
	if formatErr != nil {
		return formatErr
	}

	// TODO: approve or request changes and handle any errors from GH

	// If there was a validation error, we've already shown the errors
	// in the structured output. Silence cobra's error message but still
	// return an error to get exit code 1.
	if !output.Valid {
		cmd.SilenceErrors = true
		return errors.New(unrecoverableError)
	}

	return nil
}

func RunReleasePRCheck(cmd *cobra.Command, c *client.Client) (ReleasePRCheckOutput, error) {
	//nolint:errcheck
	pullRequestNumber, _ := cmd.Flags().GetInt("pull-request")

	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	return validatePullRequest(cmd, c, org, repo, pullRequestNumber)
}

// validatePullRequest runs PR checks and returns the structured output including validation errors.
// It only returns an error if the validation cannot be performed.
// The validation result can be found in the Valid field of the output.
// The validation errors can be found in the Errors field of the output.
func validatePullRequest(cmd *cobra.Command, client *client.Client, org, repo string, pullRequestNumber int) (ReleasePRCheckOutput, error) {

	pullRequestURL := BuildPullRequestURL(org, repo, pullRequestNumber)
	out := ReleasePRCheckOutput{
		PullRequestURL: pullRequestURL,
		Valid:          false,
		Errors:         []string{},
	}

	if pullRequestNumber <= 0 {
		return out, fmt.Errorf("%s: pull request number must be set and greater than 0", unrecoverableError)
	}

	pr, err := github.GetPullRequestFromURL(cmd.Context(), client, pullRequestURL)
	if err != nil {
		return out, fmt.Errorf("%s: %v", unrecoverableError, err)
	}

	// Skip checks for PRs by red-hat-konflux[bot]
	if pr.GetUser().GetLogin() == konfluxBotUsername {
		out.Valid = true
		return out, nil
	}

	// For rhacs-bot PRs with 'chore' prefix, print warning and skip checks
	prTitle := pr.GetTitle()
	if pr.GetUser().GetLogin() == rhacsBotUsername && strings.HasPrefix(prTitle, "chore") {
		//nolint:errcheck
		fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Skipping release PR checks for rhacs-bot PR with 'chore' prefix")
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
	jiraClient := client.GetJira()
	if jiraClient == nil {
		return out, fmt.Errorf("%s: Jira client is not configured (JIRA_TOKEN environment variable is required)", unrecoverableError)
	}

	expectedFixVersion, err := jiraClient.GetLowestUnreleasedFixVersion(major, minor)
	if err != nil {
		return out, fmt.Errorf("%s: failed to get lowest unreleased fix version: %v", unrecoverableError, err)
	}

	// Extract all Jira issue keys from the PR title (format: "PROJ-123, PROJ-456: description" or "PROJ-123: description")
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

	if len(out.Errors) == 0 {
		out.Valid = true
	}

	return out, nil
}

func BuildPullRequestURL(org, repo string, pullRequestNumber int) string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", org, repo, pullRequestNumber)
}
