package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/rel-acs/pkg/client"
)

// GetAuthenticatedUser returns the login name of the authenticated user
func GetAuthenticatedUser(ctx context.Context, c *client.Client) (string, error) {
	user, _, err := c.GetGitHub().Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	if user.Login == nil {
		return "", fmt.Errorf("user login is nil")
	}
	return *user.Login, nil
}

// IsTeamMember checks if a user is a member of a team in an organization
func IsTeamMember(ctx context.Context, c *client.Client, org, teamSlug, username string) (bool, error) {
	membership, _, err := c.GetGitHub().Teams.GetTeamMembershipBySlug(ctx, org, teamSlug, username)
	if err != nil {
		// 404 means the user is not a member
		if resp, ok := err.(*github.ErrorResponse); ok && resp.Response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to get team membership: %w", err)
	}

	// Check if the membership is active
	if membership.State != nil && *membership.State == "active" {
		return true, nil
	}

	return false, nil
}

// GetCommitForTag retrieves the commit SHA for a given tag
func GetCommitForTag(ctx context.Context, c *client.Client, org, repo, tag string) (string, error) {
	tagRef, _, err := c.GetGitHub().Git.GetRef(ctx, org, repo, "tags/"+tag)
	if err != nil {
		return "", err
	}
	tagSHA := *tagRef.Object.SHA

	tagDetails, _, err := c.GetGitHub().Git.GetTag(ctx, org, repo, tagSHA)
	if err != nil {
		return "", err
	}
	return *tagDetails.Object.SHA, nil
}

// PullRequestInfo contains information about a pull request
type PullRequestInfo struct {
	Number       int    `json:"number" yaml:"number"`
	Title        string `json:"title" yaml:"title"`
	Author       string `json:"author" yaml:"author"`
	ReviewStatus string `json:"review_status" yaml:"review_status"`
	ChecksStatus string `json:"checks_status" yaml:"checks_status"`
	IsDraft      bool   `json:"is_draft" yaml:"is_draft"`
	URL          string `json:"url" yaml:"url"`
}

// ListPullRequestsForBranch retrieves all open pull requests targeting a specific branch
func ListPullRequestsForBranch(ctx context.Context, c *client.Client, org, repo, branch string) ([]PullRequestInfo, error) {
	// List open pull requests for the target branch
	opts := &github.PullRequestListOptions{
		State: "open",
		Base:  branch,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []*github.PullRequest
	for {
		prs, resp, err := c.GetGitHub().PullRequests.List(ctx, org, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Get review status and checks status for each PR
	result := make([]PullRequestInfo, 0, len(allPRs))
	for _, pr := range allPRs {
		reviewStatus, err := getLatestReviewStatus(ctx, c, org, repo, pr.GetNumber())
		if err != nil {
			return nil, fmt.Errorf("failed to get review status for PR #%d: %w", pr.GetNumber(), err)
		}

		checksStatus, err := getChecksStatus(ctx, c, org, repo, pr.GetHead().GetSHA())
		if err != nil {
			return nil, fmt.Errorf("failed to get checks status for PR #%d: %w", pr.GetNumber(), err)
		}

		result = append(result, PullRequestInfo{
			Number:       pr.GetNumber(),
			Title:        pr.GetTitle(),
			Author:       pr.GetUser().GetLogin(),
			ReviewStatus: reviewStatus,
			ChecksStatus: checksStatus,
			IsDraft:      pr.GetDraft(),
			URL:          pr.GetHTMLURL(),
		})
	}

	return result, nil
}

// getLatestReviewStatus returns the latest review status for a pull request
func getLatestReviewStatus(ctx context.Context, c *client.Client, org, repo string, prNumber int) (string, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	var allReviews []*github.PullRequestReview
	for {
		reviews, resp, err := c.GetGitHub().PullRequests.ListReviews(ctx, org, repo, prNumber, opts)
		if err != nil {
			return "", fmt.Errorf("failed to list reviews: %w", err)
		}
		allReviews = append(allReviews, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if len(allReviews) == 0 {
		return "No Reviews", nil
	}

	// Find the most recent review (reviews are returned in chronological order)
	// We want the last one
	latestReview := allReviews[len(allReviews)-1]
	state := latestReview.GetState()

	switch state {
	case "APPROVED":
		return "Approved", nil
	case "CHANGES_REQUESTED":
		return "Changes Requested", nil
	case "COMMENTED":
		return "Commented", nil
	default:
		return "No Reviews", nil
	}
}

// getChecksStatus returns the combined status of checks for a commit
func getChecksStatus(ctx context.Context, c *client.Client, org, repo, sha string) (string, error) {
	// Get combined status (for status API)
	combinedStatus, _, err := c.GetGitHub().Repositories.GetCombinedStatus(ctx, org, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get combined status: %w", err)
	}

	// Get check runs (for checks API)
	checkRuns, _, err := c.GetGitHub().Checks.ListCheckRunsForRef(ctx, org, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get check runs: %w", err)
	}

	// Determine overall status
	// Priority: failure > pending > success
	hasFailure := false
	hasPending := false
	hasSuccess := false

	// Check combined status
	if combinedStatus.GetState() == "failure" || combinedStatus.GetState() == "error" {
		hasFailure = true
	} else if combinedStatus.GetState() == "pending" {
		hasPending = true
	} else if combinedStatus.GetState() == "success" {
		hasSuccess = true
	}

	// Check check runs
	for _, run := range checkRuns.CheckRuns {
		conclusion := run.GetConclusion()
		status := run.GetStatus()

		if conclusion == "failure" || conclusion == "timed_out" || conclusion == "action_required" {
			hasFailure = true
		} else if status == "in_progress" || status == "queued" || conclusion == "" {
			hasPending = true
		} else if conclusion == "success" {
			hasSuccess = true
		}
	}

	// No checks at all
	if !hasFailure && !hasPending && !hasSuccess && len(checkRuns.CheckRuns) == 0 && combinedStatus.GetTotalCount() == 0 {
		return "No Checks", nil
	}

	// Return status based on priority
	if hasFailure {
		return "Failed", nil
	}
	if hasPending {
		return "Pending", nil
	}
	if hasSuccess {
		return "Passed", nil
	}

	return "Unknown", nil
}
