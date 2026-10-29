package github

import (
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

func (s *ReleasePRCheckTestSuite) createCommand(pullRequestNumber int) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	s.Require().NoError(helpers.SetIntFlag(cmd, "pull-request", pullRequestNumber))
	return cmd
}

func (s *ReleasePRCheckTestSuite) verifyOutput(output any, org, repo string, pullRequestNumber int, valid bool, errorSubstrings ...string) {
	result, ok := output.(github.ReleasePRCheckOutput)
	s.True(ok, "output should be ReleasePRCheckOutput")
	s.Equal(github.BuildPullRequestURL(org, repo, pullRequestNumber), result.PullRequestURL)
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

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_Success() {
	prNumber := 123
	title := "ROX-12345: backport fix"
	author := "developer"

	helpers.MockGetPullRequest(org, repo, prNumber, title, author, "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_Konflux() {
	prNumber := 125

	helpers.MockGetPullRequest(org, repo, prNumber, "Automated update", "red-hat-konflux[bot]", "release-4.8")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_ChorePrefix() {
	prNumber := 138

	helpers.MockGetPullRequest(org, repo, prNumber, "chore(deps): update dependencies", "rhacs-bot", "release-4.8")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_NoChorePrefix() {
	prNumber := 139

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "rhacs-bot", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Release Pending", "", "4.8.1", "4.8.2", "4.8.1", "4.9.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_BotAuthor_RhacsBot_NoChorePrefix_InvalidJira() {
	prNumber := 140

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "rhacs-bot", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "In Progress", "", "4.8.0", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'")
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
	s.ErrorContains(err, "failed to get pull request")
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_InvalidReleaseBranch() {
	prNumber := 126

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-1: fix", "developer", "main")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "does not target a valid release branch")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MalformedReleaseBranch() {
	prNumber := 127

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-1: fix", "developer", "release-foo")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "does not target a valid release branch")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_NoUnreleasedFixVersion() {
	prNumber := 128

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.9.0", "5.0.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.Error(err)
	s.ErrorContains(err, "no unreleased fixVersion found for 4.8")
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MissingJiraKeyInTitle() {
	prNumber := 129

	helpers.MockGetPullRequest(org, repo, prNumber, "backport without jira key", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "does not mention at least one Jira issue key")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_WrongJiraResolution() {
	prNumber := 130

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "In Progress", "", "4.8.0", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MissingFixVersion() {
	prNumber := 131

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-12345: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.1", "4.8.2")
	helpers.MockJiraIssue("ROX-12345", "Release Pending", "", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "does not have the expected fixVersion '4.8.1'")
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_JiraIssueNotFound() {
	prNumber := 132

	helpers.MockGetPullRequest(org, repo, prNumber, "ROX-99999: backport fix", "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssueNotFound("ROX-99999")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false, "failed to get Jira issue")
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

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
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
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, false)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_DoneResolution() {
	prNumber := 136
	title := "ROX-12345: backport fix"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	s.mockValidJiraFlow("ROX-12345", "Closed", "Done", "4.8.1", "4.8.1")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
	s.True(gock.IsDone())
}

func (s *ReleasePRCheckTestSuite) TestReleasePRCheck_MixedValidResolutions() {
	prNumber := 137
	title := "ROX-111, ROX-222: backport fixes"

	helpers.MockGetPullRequest(org, repo, prNumber, title, "developer", "release-4.8")
	helpers.MockJiraTokenValid()
	helpers.MockUnreleasedProjectVersions("4.8.0")
	helpers.MockJiraIssue("ROX-111", "Release Pending", "", "4.8.0")
	helpers.MockJiraIssue("ROX-222", "Closed", "Done", "4.8.0")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyOutput(output, org, repo, prNumber, true)
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

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(prNumber)

	output, err := github.RunReleasePRCheck(cmd, testClient)

	s.NoError(err)
	s.NotNil(output)
	// Should report both errors: wrong status for ROX-222 and wrong fixVersion for ROX-333
	s.verifyOutput(output, org, repo, prNumber, false, "should have status 'Release Pending' or status 'Closed' with resolution 'Done'", "does not have the expected fixVersion")
	s.True(gock.IsDone())
}
