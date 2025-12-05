package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/client"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type PullRequestInfo struct {
	Number       int    `json:"number" yaml:"number"`
	Title        string `json:"title" yaml:"title"`
	Author       string `json:"author" yaml:"author"`
	ReviewStatus string `json:"review-status" yaml:"review-status"`
	ChecksStatus string `json:"checks-status" yaml:"checks-status"`
	IsDraft      bool   `json:"is-draft" yaml:"is-draft"`
	URL          string `json:"url" yaml:"url"`
}

// GetCommitForTag retrieves the commit SHA for a given tag
func GetCommitForTag(ctx context.Context, c *client.Client, org, repo, tag string) (github.GitObject, error) {
	tagRef, _, err := c.GetGitHub().Git.GetRef(ctx, org, repo, "tags/"+tag)
	if err != nil {
		return github.GitObject{}, err
	}
	tagSHA := *tagRef.Object.SHA

	tagDetails, _, err := c.GetGitHub().Git.GetTag(ctx, org, repo, tagSHA)
	if err != nil {
		return github.GitObject{}, err
	}
	return *tagDetails.Object, nil
}

func ListPullRequests(ctx context.Context, c *client.Client, org, repo, baseBranch string, includeDrafts bool) ([]PullRequestInfo, error) {
	prs, err := retrievePullRequests(ctx, c, org, repo, baseBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pull requests: %w", err)
	}

	pullRequestInfos := []PullRequestInfo{}
	for _, pr := range prs {
		if !includeDrafts && *pr.Draft {
			continue
		}
		combinedChecksStatus, err := GetCombinedChecksStatus(ctx, c, org, repo, *pr.Head.SHA)
		if err != nil {
			return nil, fmt.Errorf("failed to get combined checks status: %w", err)
		}

		latestReviewStatus, err := GetLatestReviewStatus(ctx, c, org, repo, *pr.Number)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest review status: %w", err)
		}
		pullRequestInfos = append(pullRequestInfos, PullRequestInfo{
			Number:       *pr.Number,
			Title:        *pr.Title,
			Author:       *pr.User.Login,
			ReviewStatus: latestReviewStatus,
			ChecksStatus: combinedChecksStatus,
			IsDraft:      *pr.Draft,
			URL:          *pr.HTMLURL,
		})
	}
	return pullRequestInfos, nil
}

func retrievePullRequests(ctx context.Context, c *client.Client, org, repo, baseBranch string) ([]*github.PullRequest, error) {
	options := &github.PullRequestListOptions{
		Base:      baseBranch,
		State:     "open",
		Direction: "asc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []*github.PullRequest
	for {
		prs, resp, err := c.GetGitHub().PullRequests.List(ctx, org, repo, options)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return allPRs, nil
}

// GetLatestReviewStatus returns the latest review status for a pull request
func GetLatestReviewStatus(ctx context.Context, c *client.Client, org, repo string, prNumber int) (string, error) {
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
	// We want the last one.
	latestReview := allReviews[len(allReviews)-1]
	state := latestReview.GetState()
	state = strings.ReplaceAll(state, "_", " ")
	return cases.Title(language.English).String(state), nil
}

// GetCombinedChecksStatus returns the combined status of checks for a commit
func GetCombinedChecksStatus(ctx context.Context, c *client.Client, org, repo, sha string) (string, error) {
	// Get combined status (from status API)
	combinedStatus, _, err := c.GetGitHub().Repositories.GetCombinedStatus(ctx, org, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get combined status: %w", err)
	}

	// Get check runs (from checks API)
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
	combinedStatusState := combinedStatus.GetState()
	switch combinedStatusState {
	case "failure":
		hasFailure = true
	case "pending":
		hasPending = true
	case "success":
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

	// No checks found
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
