package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/stackrox/relacs/pkg/utils"
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

type CommitInfo struct {
	SHA string `json:"sha" yaml:"sha"`
}

type TeamMembershipInfo struct {
	Username string `json:"username" yaml:"username"`
	Team     string `json:"team" yaml:"team"`
	Org      string `json:"org" yaml:"org"`
	Role     string `json:"role" yaml:"role"`
	State    string `json:"state" yaml:"state"`
}

// GetCommitForTag retrieves the commit SHA for a given tag
func GetCommitForTag(ctx context.Context, gh *github.Client, org, repo, tag string) (CommitInfo, error) {
	// TODO: allow using an ACS version to retrieve a commit SHA (ie tags + make tag).
	tagRef, _, err := gh.Git.GetRef(ctx, org, repo, "tags/"+tag)
	if err != nil {
		return CommitInfo{}, err
	}

	var commitSHA string
	switch *tagRef.Object.Type {
	case "commit":
		commitSHA = *tagRef.Object.SHA
	case "tag":
		tagDetails, _, err := gh.Git.GetTag(ctx, org, repo, *tagRef.Object.SHA)
		if err != nil {
			return CommitInfo{}, err
		}
		commitSHA = *tagDetails.Object.SHA
	default:
		return CommitInfo{}, fmt.Errorf("invalid tag type: %s", *tagRef.Object.Type)
	}

	commit, _, err := gh.Git.GetCommit(ctx, org, repo, commitSHA)
	if err != nil {
		return CommitInfo{}, err
	}
	return CommitInfo{
		SHA: *commit.SHA,
	}, nil
}

// GetCommitFromACSVersion retrieves the commit SHA for a given ACS version
func GetCommitFromACSVersion(ctx context.Context, gh *github.Client, org, repo, acsVersion string) (CommitInfo, error) {
	if utils.IsACSVersionTagged(acsVersion) {
		return GetCommitForTag(ctx, gh, org, repo, acsVersion)
	}

	shortSHA, err := utils.GetCommitSHAFromGitDescribeVersion(acsVersion)
	if err != nil {
		return CommitInfo{}, fmt.Errorf("failed to get commit SHA from git-describe version: %w", err)
	}
	commitObj, _, err := gh.Repositories.GetCommit(ctx, org, repo, shortSHA, nil)
	if err != nil {
		return CommitInfo{}, fmt.Errorf("failed to get commit: %w", err)
	}
	return CommitInfo{
		SHA: *commitObj.SHA,
	}, nil
}

func ListPullRequests(ctx context.Context, gh *github.Client, org, repo, baseBranch string, includeDrafts bool) ([]PullRequestInfo, error) {
	prs, err := retrievePullRequests(ctx, gh, org, repo, baseBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pull requests: %w", err)
	}

	pullRequestInfos := []PullRequestInfo{}
	for _, pr := range prs {
		if !includeDrafts && *pr.Draft {
			continue
		}
		combinedChecksStatus, err := GetCombinedChecksStatus(ctx, gh, org, repo, *pr.Head.SHA)
		if err != nil {
			return nil, fmt.Errorf("failed to get combined checks status: %w", err)
		}

		latestReviewStatus, err := GetLatestReviewStatus(ctx, gh, org, repo, *pr.Number)
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

func retrievePullRequests(ctx context.Context, gh *github.Client, org, repo, baseBranch string) ([]*github.PullRequest, error) {
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
		prs, resp, err := gh.PullRequests.List(ctx, org, repo, options)
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
func GetLatestReviewStatus(ctx context.Context, gh *github.Client, org, repo string, prNumber int) (string, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	var allReviews []*github.PullRequestReview
	for {
		reviews, resp, err := gh.PullRequests.ListReviews(ctx, org, repo, prNumber, opts)
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

	latestReview := allReviews[len(allReviews)-1]
	state := latestReview.GetState()
	state = strings.ReplaceAll(state, "_", " ")
	return cases.Title(language.English).String(state), nil
}

// AddTeamMember adds or invites a user to a team in a GitHub organization.
func AddTeamMember(ctx context.Context, gh *github.Client, org, team, user, role string) (TeamMembershipInfo, error) {
	opts := &github.TeamAddTeamMembershipOptions{
		Role: role,
	}
	membership, _, err := gh.Teams.AddTeamMembershipBySlug(ctx, org, team, user, opts)
	if err != nil {
		return TeamMembershipInfo{}, fmt.Errorf("failed to add user %q to team %q: %w", user, team, err)
	}
	return TeamMembershipInfo{
		Username: user,
		Team:     team,
		Org:      org,
		Role:     membership.GetRole(),
		State:    membership.GetState(),
	}, nil
}

// GetCombinedChecksStatus returns the combined status of checks for a commit
func GetCombinedChecksStatus(ctx context.Context, gh *github.Client, org, repo, sha string) (string, error) {
	combinedStatus, _, err := gh.Repositories.GetCombinedStatus(ctx, org, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get combined status: %w", err)
	}

	checkRuns, _, err := gh.Checks.ListCheckRunsForRef(ctx, org, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get check runs: %w", err)
	}

	hasFailure := false
	hasPending := false
	hasSuccess := false

	combinedStatusState := combinedStatus.GetState()
	switch combinedStatusState {
	case "failure":
		hasFailure = true
	case "pending":
		hasPending = true
	case "success":
		hasSuccess = true
	}

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

	if !hasFailure && !hasPending && !hasSuccess && len(checkRuns.CheckRuns) == 0 && combinedStatus.GetTotalCount() == 0 {
		return "No Checks", nil
	}

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

func GetPullRequestFromURL(ctx context.Context, gh *github.Client, url string) (*github.PullRequest, error) {
	org, repo, prNumber, err := parsePullRequestURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request URL: %w", err)
	}
	pr, _, err := gh.PullRequests.Get(ctx, org, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}
	return pr, nil
}

func parsePullRequestURL(url string) (string, string, int, error) {
	// Expected format: https://github.com/org/repo/pull/123
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	if len(parts) < 7 {
		return "", "", 0, fmt.Errorf("invalid pull request URL: %s (expected format: https://github.com/org/repo/pull/123)", url)
	}

	if parts[5] != "pull" {
		return "", "", 0, fmt.Errorf("invalid pull request URL: %s (expected 'pull' in URL)", url)
	}

	prNumber, err := strconv.Atoi(parts[6])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid PR number in URL %s: %w", url, err)
	}

	return parts[3], parts[4], prNumber, nil
}
