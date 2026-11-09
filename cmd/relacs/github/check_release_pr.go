package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	gogithub "github.com/google/go-github/v79/github"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/client/jira"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
)

const (
	konfluxBotUsername   = "red-hat-konflux[bot]"
	rhacsBotUsername     = "rhacs-bot"
	resolutionDone       = "Done"
	statusClosed         = "Closed"
	statusReleasePending = "Release Pending"

	reviewMarker = "<!-- relacs:check-release-pr -->"
)

type ReleasePRCheckOutput struct {
	Org               string   `json:"org" yaml:"org"`
	Repo              string   `json:"repo" yaml:"repo"`
	PullRequestNumber int      `json:"pull-request-number" yaml:"pull-request-number"`
	Valid             bool     `json:"valid"`
	Errors            []string `json:"errors"`
	SkipReviewReason  string   `json:"skip-review-reason"`
}

func (o ReleasePRCheckOutput) String() string {
	if o.Valid {
		return fmt.Sprintf("The PR %s/%s#%d satisfies the release branch PR criteria.", o.Org, o.Repo, o.PullRequestNumber)
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "The PR %s/%s#%d does not satisfy the release branch PR criteria:\n", o.Org, o.Repo, o.PullRequestNumber)
	for i, err := range o.Errors {
		fmt.Fprintf(&buf, "  %d. %s\n", i+1, err)
	}

	if o.SkipReviewReason != "" {
		fmt.Fprintf(&buf, "Skipped adding or updating review to this PR because %s.\n", o.SkipReviewReason)
	}
	return buf.String()
}

func prRef(org, repo string, number int) string {
	return fmt.Sprintf("%s/%s#%d", org, repo, number)
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
	checkReleasePRCmd.Flags().Int("pull-request", 0, "Pull Request number to check")
	checkReleasePRCmd.Flags().Bool("dry-run", true, "Dry run mode, do not post comments to PRs")
	checkReleasePRCmd.Flags().String("reviewer-login", "", "GitHub login to use for finding previous reviews (bypasses the /user API call, useful for GitHub App tokens)")

	// We only allow text output in this command, so have override
	// and hide the --output flag from the root command.
	checkReleasePRCmd.Flags().StringP("output", "o", "text", "")
	//nolint:errcheck
	checkReleasePRCmd.Flags().MarkHidden("output")
}

func runCheckReleasePRCommand(cmd *cobra.Command, _ []string) error {
	if cmd.Flags().Changed("output") {
		return fmt.Errorf("unknown flag: --output")
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	_, err = RunReleasePRCheck(cmd, c)
	if err != nil {
		return err
	}
	return nil
}

func RunReleasePRCheck(cmd *cobra.Command, c *client.Client) ([]ReleasePRCheckOutput, error) {
	//nolint:errcheck
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	var pullRequestNumbers []int
	if cmd.Flags().Changed("pull-request") {
		pullRequestNumber, _ := cmd.Flags().GetInt("pull-request")
		pullRequestNumbers = []int{pullRequestNumber}
	} else {
		var err error
		pullRequestNumbers, err = github.ListOpenReleasePRs(cmd.Context(), c, org, repo)
		if err != nil {
			return nil, err
		}
	}

	singlePRMode := cmd.Flags().Changed("pull-request")

	anyValidationFailed := false
	var checkOutput []ReleasePRCheckOutput
	for _, prNumber := range pullRequestNumbers {
		out, err := validatePullRequest(cmd, c, org, repo, prNumber)
		if err != nil {
			return nil, err
		}

		if out.SkipReviewReason == "the PR is a draft" {
			if singlePRMode {
				return nil, fmt.Errorf("PR %s/%s#%d is a draft", org, repo, prNumber)
			}
			//nolint:errcheck
			fmt.Fprintf(cmd.ErrOrStderr(), "%s: skipping because %s\n", prRef(org, repo, prNumber), out.SkipReviewReason)
			continue
		}

		checkOutput = append(checkOutput, out)
		if !out.Valid {
			anyValidationFailed = true
		}

		err = common.FormatOutput("text", out)
		if err != nil {
			return nil, err
		}

		if out.SkipReviewReason == "" {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if !dryRun {
				reviewerLogin, _ := cmd.Flags().GetString("reviewer-login")
				err := postVerdictToPR(cmd.Context(), c, out, reviewerLogin)
				if err != nil {
					return nil, err
				}
			} else {
				//nolint:errcheck
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: dry run mode, skipping posting comment to PR\n", prRef(org, repo, out.PullRequestNumber))
			}
		} else {
			//nolint:errcheck
			fmt.Fprintf(cmd.ErrOrStderr(), "%s: skipping review because %s\n", prRef(org, repo, out.PullRequestNumber), out.SkipReviewReason)
		}
	}

	if anyValidationFailed {
		return checkOutput, errors.New("one or more release branch PRs do not satisfy the criteria")
	}

	return checkOutput, nil
}

// validatePullRequest runs PR checks and returns the structured output including validation errors.
// It only returns an error if the validation cannot be performed.
// The validation result can be found in the Valid field of the output.
// The validation errors can be found in the Errors field of the output.
func validatePullRequest(cmd *cobra.Command, client *client.Client, org, repo string, pullRequestNumber int) (ReleasePRCheckOutput, error) {
	out := ReleasePRCheckOutput{
		Org:               org,
		Repo:              repo,
		PullRequestNumber: pullRequestNumber,
	}

	if pullRequestNumber <= 0 {
		return out, fmt.Errorf("pull request number must be set and greater than 0")
	}

	pr, err := github.GetPullRequest(cmd.Context(), client, org, repo, pullRequestNumber)
	if err != nil {
		var ghErr *gogithub.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
			out.Errors = append(out.Errors, fmt.Sprintf("failed to get pull request: %v", err))
			out.SkipReviewReason = "the PR could not be found"
			return out, nil
		}
		return out, fmt.Errorf("failed to get pull request: %w", err)
	}

	if pr.GetDraft() {
		out.SkipReviewReason = "the PR is a draft"
		return out, nil
	}

	// Check if the PR targets a release branch (base is release-<major>.<minor>)
	baseBranch := pr.GetBase().GetRef()
	matches := releaseBranchPattern.FindStringSubmatch(baseBranch)
	if matches == nil {
		out.Errors = append(out.Errors, fmt.Sprintf("the PR does not target a valid release branch (expected release-<major>.<minor>): %s", baseBranch))
		out.SkipReviewReason = "the check should not have run for this PR"
		return out, nil
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])

	// Reject PRs to release branches that are from forks.
	if !strings.EqualFold(pr.GetHead().GetRepo().GetFullName(), fmt.Sprintf("%s/%s", org, repo)) {
		out.Errors = append(out.Errors, "PRs to release branches cannot be forks. Instead, let your PR target the default branch of this repository and the team will backport it, if necessary.")
		return out, nil
	}

	// Skip checks for PRs by red-hat-konflux[bot]
	if pr.GetUser().GetLogin() == konfluxBotUsername {
		//nolint:errcheck
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: skipping release PR checks for %s\n", prRef(org, repo, pullRequestNumber), konfluxBotUsername)
		out.Valid = true
		return out, nil
	}

	// For rhacs-bot PRs with 'chore' prefix, print warning and skip checks
	prTitle := pr.GetTitle()
	if pr.GetUser().GetLogin() == rhacsBotUsername && strings.HasPrefix(prTitle, "chore") {
		//nolint:errcheck
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: skipping release PR checks for rhacs-bot PR with 'chore' prefix\n", prRef(org, repo, pullRequestNumber))
		out.Valid = true
		return out, nil
	}

	// Check if Jira client is configured
	jiraClient := client.GetJira()
	if jiraClient == nil {
		return out, fmt.Errorf("jira client is not configured (JIRA_TOKEN environment variable is required)")
	}

	expectedFixVersion, err := jiraClient.GetLowestUnreleasedFixVersion(major, minor)
	if err != nil {
		if errors.Is(err, jira.ErrNoFixVersion) {
			out.Errors = append(out.Errors, fmt.Sprintf("failed to get lowest unreleased fix version: %v", err))
			return out, nil
		}
		return out, fmt.Errorf("failed to get lowest unreleased fix version: %w", err)
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
			var apiErr *jira.APIError
			if errors.As(err, &apiErr) && !apiErr.IsServerError() {
				out.Errors = append(out.Errors, fmt.Sprintf("failed to get Jira issue: %v", err))
				continue
			}
			return out, fmt.Errorf("failed to get Jira issue %s: %w", issueKey, err)
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

func postVerdictToPR(ctx context.Context, client *client.Client, out ReleasePRCheckOutput, reviewerLogin string) error {
	// TODO: remove disclaimer once the check goes live
	previewDisclaimer := "**This is a preview of the release branch PR check. It it not a blocking review yet.**"

	event := "COMMENT"
	// TODO: uncomment this once the check goes live
	// if !out.Valid {
	// 	event = "REQUEST_CHANGES"
	// }

	// TODO: remove the Event field once the check goes live
	comment := fmt.Sprintf("%s\n%s\n\nEvent: *%s*\n\n%s", reviewMarker, previewDisclaimer, event, out.String())

	var (
		lastReview *gogithub.PullRequestReview
		err        error
	)
	if reviewerLogin != "" {
		lastReview, err = github.GetLastReviewForLogin(ctx, client, out.Org, out.Repo, out.PullRequestNumber, reviewerLogin, reviewMarker)
	} else {
		lastReview, err = github.GetLastReviewForCurrentUser(ctx, client, out.Org, out.Repo, out.PullRequestNumber, reviewMarker)
	}
	if err != nil {
		return fmt.Errorf("failed to get last review for current user: %w", err)
	}

	// TODO: update workflows based on new event once the check goes live
	if lastReview == nil {
		err = github.PostCommentToPR(ctx, client, out.Org, out.Repo, out.PullRequestNumber, event, comment)
		if err != nil {
			return fmt.Errorf("failed to post comment to PR: %w", err)
		}
	} else {
		// TODO: once the check goes live, update the logic to dismiss previous review (if has different event), then create new review
		err = github.UpdateReview(ctx, client, out.Org, out.Repo, out.PullRequestNumber, lastReview.GetID(), comment)
		if err != nil {
			return fmt.Errorf("failed to update review: %w", err)
		}
	}
	return nil
}
