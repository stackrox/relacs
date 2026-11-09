package github

import (
	"bytes"
	"testing"

	"github.com/h2non/gock"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

const (
	org  = "stackrox"
	repo = "stackrox"
)

func TestReleasePRCheckTestSuite(t *testing.T) {
	suite.Run(t, new(ReleasePRCheckTestSuite))
}

type ReleasePRCheckTestSuite struct {
	suite.Suite
}

func (s *ReleasePRCheckTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *ReleasePRCheckTestSuite) createCommand(pullRequestNumbers ...int) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().LessOrEqual(len(pullRequestNumbers), 1, "at most one pull request number can be provided")
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	if len(pullRequestNumbers) == 1 {
		s.Require().NoError(helpers.SetIntFlag(cmd, "pull-request", pullRequestNumbers[0]))
	}
	return cmd
}

func (s *ReleasePRCheckTestSuite) verifyOutput(output github.ReleasePRCheckOutput, org, repo string, pullRequestNumber int, valid bool, skipReviewReasonEmpty bool, errorSubstrings ...string) {
	s.Equal(org, output.Org)
	s.Equal(repo, output.Repo)
	s.Equal(pullRequestNumber, output.PullRequestNumber)
	s.Equal(valid, output.Valid)
	s.Len(output.Errors, len(errorSubstrings))
	for i, substring := range errorSubstrings {
		s.Contains(output.Errors[i], substring)
	}
	s.Equal(skipReviewReasonEmpty, output.SkipReviewReason == "")
}

func (s *ReleasePRCheckTestSuite) verifySkipReviewReason(output github.ReleasePRCheckOutput, skipReviewReason string) {
	s.Equal(skipReviewReason, output.SkipReviewReason)
}

func (s *ReleasePRCheckTestSuite) mockValidJiraFlow(issueKey, status, resolution, expectedFixVersion string, unreleasedVersions ...string) {
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions(unreleasedVersions...)
	helpers.MockJiraIssue(issueKey, status, resolution, expectedFixVersion)
}

func (s *ReleasePRCheckTestSuite) mockListingReviews(prNumber int, reviews []helpers.ReviewData) {
	helpers.MockGetCurrentUser()
	helpers.MockListReviews(org, repo, prNumber, reviews)
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_Success() {
	prNumber := 123
	title := "ROX-12345: backport fix"
	author := "developer"

	helpers.MockGetPullRequest(org, repo, prNumber, title, author, "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output[0])
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_Konflux() {
	prNumber := 125

	helpers.MockGetPullRequest(org, repo, prNumber, "Automated update", "red-hat-konflux[bot]", "release-4.8")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_ChorePrefix() {
	prNumber := 138

	helpers.MockGetPullRequest(org, repo, prNumber, "chore(deps): update dependencies", "rhacs-bot", "release-4.8")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_NoChorePrefix() {
	prNumber := 139

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "rhacs-bot", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_PullRequestNotFound() {
	prNumber := 999

	helpers.MockGetPullRequestNotFound(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, false, "failed to get pull request")
	s.verifySkipReviewReason(output[0], "the PR could not be found")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_InvalidReleaseBranch() {
	prNumber := 126

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-1: fix", "developer", "main")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, false, "does not target a valid release branch")
	s.verifySkipReviewReason(output[0], "the check should not have run for this PR")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_NoUnreleasedFixVersion() {
	prNumber := 128

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.9.0", "5.0.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, true, "no unreleased fixVersion found for 4.8")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MissingJiraKeyInTitle() {
	prNumber := 129

	helpers.MockGetPullRequest(org, repo, prNumber, "backport without jira key", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, true, "does not mention at least one Jira issue key")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_WrongJiraResolution() {
	prNumber := 130

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "In Progress", "", "4.8.0", "4.8.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, true, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MissingFixVersion() {
	prNumber := 131

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.1", "4.8.2")
	helpers.MockJiraIssue("ROX-12345", "Release Pending", "", "4.8.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, true, "does not have the expected fixVersion '4.8.1'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_JiraIssueNotFound() {
	prNumber := 132

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-99999: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssueNotFound("ROX-99999")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, false, true, "failed to get Jira issue")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_JiraIssueServerError() {
	prNumber := 144

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssueServerError("ROX-12345")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "failed to get Jira issue ROX-12345")
	s.Nil(output)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MultipleJiraKeys_Success() {
	prNumber := 133
	title := "ROX-111, ROX-222: backport fixes"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssue("ROX-111", "Release Pending", "", "4.8.0")
	helpers.MockJiraIssue("ROX-222", "Release Pending", "", "4.8.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_JiraTokenInvalid() {
	prNumber := 134

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenInvalid()

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "jira token validation failed")
	s.Nil(output)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_DoneResolution() {
	prNumber := 136
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Closed", "Done", "4.8.1", "4.8.1")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MultipleJiraKeys_PartialFailure() {
	prNumber := 135
	title := "ROX-111, ROX-222, ROX-333: backport fixes"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssue("ROX-111", "Release Pending", "", "4.8.0")
	helpers.MockJiraIssue("ROX-222", "In Progress", "", "4.8.0")
	helpers.MockJiraIssue("ROX-333", "Release Pending", "", "4.9.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "one or more release branch PRs do not satisfy the criteria")
	s.Len(output, 1)
	// Should report both errors: wrong status for ROX-222 and wrong fixVersion for ROX-333
	s.verifyOutput(output[0], org, repo, prNumber, false, true, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'", "does not have the expected fixVersion")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_UpdateExistingReview() {
	prNumber := 138
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")
	s.mockListingReviews(prNumber, []helpers.ReviewData{
		{ID: 1, User: "test-user", State: "COMMENTED", Review: "<!-- relacs:check-release-pr --> previous review"},
		{ID: 2, User: "test-user", State: "COMMENTED", Review: "<!-- relacs:check-release-pr --> latest review"},
		{ID: 3, User: "other-user", State: "COMMENTED", Review: "comment"},
	})
	helpers.MockUpdateReview(org, repo, prNumber, 2)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)
	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, prNumber, true, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_DraftPR_SinglePRMode() {
	prNumber := 141

	helpers.MockGetDraftPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "is a draft")
	s.Nil(output)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_DraftPR_BatchMode() {
	draftPR := 142
	validPR := 143

	helpers.MockListOpenReleasePRs(org, repo, []int{draftPR, validPR}, "release-4.8")
	helpers.MockGetDraftPullRequest(org, repo, draftPR, "ROX-11111: draft fix", "developer", "release-4.8")
	helpers.MockGetPullRequest(org, repo, validPR, "ROX-22222: backport fix", "developer", "release-4.8")
	s.mockValidJiraFlow("ROX-22222", "Release Pending", "", "4.8.1", "4.8.1")
	s.mockListingReviews(validPR, []helpers.ReviewData{})
	helpers.MockPostCommentToPR(org, repo, validPR)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand()

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, 1)
	s.verifyOutput(output[0], org, repo, validPR, true, true)
	s.Contains(stderr.String(), "stackrox/stackrox#142: skipping because the PR is a draft")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestMultiplePRs_Success() {
	prNumbers := []int{139, 140}
	pullRequests := []helpers.PRData{
		{Number: prNumbers[0], Title: "ROX-12345: backport fix", Author: "developer", BaseBranch: "release-4.8"},
		{Number: prNumbers[1], Title: "ROX-23456: backport fix", Author: "developer", BaseBranch: "release-4.8"},
	}

	helpers.MockListOpenReleasePRs(org, repo, prNumbers, "release-4.8")
	for _, pullRequest := range pullRequests {
		helpers.MockGetPullRequest(org, repo, pullRequest.Number, pullRequest.Title, pullRequest.Author, pullRequest.BaseBranch)
		s.mockListingReviews(pullRequest.Number, []helpers.ReviewData{})
		helpers.MockPostCommentToPR(org, repo, pullRequest.Number)
	}
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")
	helpers.MockUnreleasedProjectVersions("4.8.2", "4.8.1", "4.9.0")
	helpers.MockJiraIssue("ROX-23456", "Release Pending", "", "4.8.1")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand()
	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.Len(output, len(prNumbers))
	for i, pullRequest := range pullRequests {
		s.verifyOutput(output[i], org, repo, pullRequest.Number, true, true)
	}

	s.True(gock.IsDone())
}
