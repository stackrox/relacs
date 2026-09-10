package github

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
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

func (s *ReleasePRCheckTestSuite) createCommand(pullRequestURL string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "pull-request", pullRequestURL))
	return cmd
}

func (s *ReleasePRCheckTestSuite) verifyOutput(output any, pullRequestURL string, valid bool, errorSubstrings ...string) {
	result, ok := output.(github.ReleasePRCheckOutput)
	s.True(ok, "output should be ReleasePRCheckOutput")
	s.Equal(pullRequestURL, result.PullRequestURL)
	s.Equal(valid, result.Valid)
	s.Len(result.Errors, len(errorSubstrings))
	for i, substring := range errorSubstrings {
		s.Contains(result.Errors[i], substring)
	}
}

func (s *ReleasePRCheckTestSuite) mockValidJiraFlow(issueKey, status, resolution, expectedFixVersion string, unreleasedVersions ...string) {
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions(unreleasedVersions...)
	helpers.MockJiraIssue(issueKey, status, resolution, expectedFixVersion)
}

func (s *ReleasePRCheckTestSuite) mockMergeWindow(org, repo, productStream, status string) {
	helpers.MockMergeWindowFile(org, repo, productStream, status)
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_Success() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 123
	prURL := "https://github.com/stackrox/stackrox/pull/123"
	title := "ROX-12345: backport fix"
	author := "developer"

	helpers.MockGetPullRequest(org, repo, prNumber, title, author, "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_Konflux() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 125
	prURL := "https://github.com/stackrox/stackrox/pull/125"

	helpers.MockGetPullRequest(org, repo, prNumber, "Automated update", "red-hat-konflux[bot]", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_ChorePrefix() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 138
	prURL := "https://github.com/stackrox/stackrox/pull/138"

	helpers.MockGetPullRequest(org, repo, prNumber, "chore(deps): update dependencies", "rhacs-bot", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_NoChorePrefix() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 139
	prURL := "https://github.com/stackrox/stackrox/pull/139"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "rhacs-bot", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_NoChorePrefix_InvalidJira() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 140
	prURL := "https://github.com/stackrox/stackrox/pull/140"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "rhacs-bot", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	s.mockValidJiraFlow("ROX-12345", "In Progress", "", "4.8.0", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_PullRequestNotFound() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 999
	prURL := "https://github.com/stackrox/stackrox/pull/999"

	helpers.MockGetPullRequestNotFound(org, repo, prNumber)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "failed to get pull request")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_InvalidReleaseBranch() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 126
	prURL := "https://github.com/stackrox/stackrox/pull/126"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-1: fix", "developer", "main")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "does not target a valid release branch")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MalformedReleaseBranch() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 127
	prURL := "https://github.com/stackrox/stackrox/pull/127"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-1: fix", "developer", "release-foo")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "does not target a valid release branch")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_NoUnreleasedFixVersion() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 128
	prURL := "https://github.com/stackrox/stackrox/pull/128"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.9.0", "5.0.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "no unreleased fixVersion found for 4.8")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MissingJiraKeyInTitle() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 129
	prURL := "https://github.com/stackrox/stackrox/pull/129"

	helpers.MockGetPullRequest(org, repo, prNumber, "backport without jira key", "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "does not mention at least one Jira issue key")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_WrongJiraResolution() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 130
	prURL := "https://github.com/stackrox/stackrox/pull/130"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	s.mockValidJiraFlow("ROX-12345", "In Progress", "", "4.8.0", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MissingFixVersion() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 131
	prURL := "https://github.com/stackrox/stackrox/pull/131"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.1", "4.8.2")
	helpers.MockJiraIssue("ROX-12345", "Release Pending", "", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "does not have the expected fixVersion '4.8.1'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_JiraIssueNotFound() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 132
	prURL := "https://github.com/stackrox/stackrox/pull/132"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-99999: backport fix", "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssueNotFound("ROX-99999")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "failed to get Jira issue")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MultipleJiraKeys_Success() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 133
	prURL := "https://github.com/stackrox/stackrox/pull/133"
	title := "ROX-111, ROX-222: backport fixes"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssue("ROX-111", "Release Pending", "", "4.8.0")
	helpers.MockJiraIssue("ROX-222", "Release Pending", "", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_JiraTokenInvalid() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 134
	prURL := "https://github.com/stackrox/stackrox/pull/134"

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenInvalid()

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "jira token validation failed")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_DoneResolution() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 136
	prURL := "https://github.com/stackrox/stackrox/pull/136"
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	s.mockValidJiraFlow("ROX-12345", "Closed", "Done", "4.8.1", "4.8.1")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MixedValidResolutions() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 137
	prURL := "https://github.com/stackrox/stackrox/pull/137"
	title := "ROX-111, ROX-222: backport fixes"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssue("ROX-111", "Release Pending", "", "4.8.0")
	helpers.MockJiraIssue("ROX-222", "Closed", "Done", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MultipleJiraKeys_PartialFailure() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 135
	prURL := "https://github.com/stackrox/stackrox/pull/135"
	title := "ROX-111, ROX-222, ROX-333: backport fixes"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "OPEN")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssue("ROX-111", "Release Pending", "", "4.8.0")
	helpers.MockJiraIssue("ROX-222", "In Progress", "", "4.8.0")
	helpers.MockJiraIssue("ROX-333", "Release Pending", "", "4.9.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	// Should report both errors: wrong status for ROX-222 and wrong fixVersion for ROX-333
	s.verifyOutput(output, prURL, false, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'", "does not have the expected fixVersion")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MergeWindowClosed() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 141
	prURL := "https://github.com/stackrox/stackrox/pull/141"
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "CLOSED")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.1")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "merge window is CLOSED (expected OPEN)")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MergeWindowUnknown() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 142
	prURL := "https://github.com/stackrox/stackrox/pull/142"
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	helpers.MockMergeWindowFileNotFound(org, repo, "4.8")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.1")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "merge window is UNKNOWN (expected OPEN)")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_Konflux_MergeWindowClosed() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 143
	prURL := "https://github.com/stackrox/stackrox/pull/143"

	helpers.MockGetPullRequest(org, repo, prNumber, "Automated update", "red-hat-konflux[bot]", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "CLOSED")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "merge window is CLOSED (expected OPEN)")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_ChorePrefix_MergeWindowClosed() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 144
	prURL := "https://github.com/stackrox/stackrox/pull/144"

	helpers.MockGetPullRequest(org, repo, prNumber, "chore(deps): update dependencies", "rhacs-bot", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "CLOSED")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	s.verifyOutput(output, prURL, false, "merge window is CLOSED (expected OPEN)")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MergeWindowClosedWithJiraErrors() {
	org := "stackrox"
	repo := "stackrox"
	prNumber := 145
	prURL := "https://github.com/stackrox/stackrox/pull/145"
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockMergeWindow(org, repo, "4.8", "CLOSED")
	s.mockValidJiraFlow("ROX-12345", "In Progress", "", "4.8.0", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prURL)

	output, err := github.RunReleasePRCheck(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.NotNil(output)
	// Should report both merge window error AND Jira status error
	result, ok := output.(github.ReleasePRCheckOutput)
	s.True(ok, "output should be ReleasePRCheckOutput")
	s.Equal(prURL, result.PullRequestURL)
	s.False(result.Valid)
	s.Len(result.Errors, 2)
	s.Contains(result.Errors[0], "merge window is CLOSED (expected OPEN)")
	s.Contains(result.Errors[1], "should have status 'Release Pending' or status 'Closed' with resolution 'Done'")
	s.True(gock.IsDone())
}
