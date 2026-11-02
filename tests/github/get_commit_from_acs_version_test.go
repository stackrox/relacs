package github

import (
	"fmt"
	"testing"

	"github.com/h2non/gock"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestGetCommitFromACSVersionTestSuite(t *testing.T) {
	suite.Run(t, new(GetCommitFromACSVersionTestSuite))
}

type GetCommitFromACSVersionTestSuite struct {
	suite.Suite
}

func (s *GetCommitFromACSVersionTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

// createCommand creates and configures a mock command with the given flags
func (s *GetCommitFromACSVersionTestSuite) createCommand(org, repo, acsVersion, outputFormat string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	s.Require().NoError(helpers.SetFlag(cmd, "acs-version", acsVersion))
	s.Require().NoError(helpers.SetFlag(cmd, "output", outputFormat))
	return cmd
}

// verifyCommitOutput verifies the commit output structure
func (s *GetCommitFromACSVersionTestSuite) verifyCommitOutput(output any, org, repo, acsVersion, commitSHA string) {
	commitOutput, ok := output.(github.CommitFromACSVersionOutput)
	s.True(ok)
	s.Equal(org, commitOutput.Organization)
	s.Equal(repo, commitOutput.Repository)
	s.Equal(acsVersion, commitOutput.ACSVersion)
	s.Equal(commitSHA, commitOutput.Commit.SHA)
	s.Equal(fmt.Sprintf("https://github.com/%s/%s/commit/%s", org, repo, commitSHA), commitOutput.URL)
}

func (s *GetCommitFromACSVersionTestSuite) TestGetCommitFromACSVersion_TaggedVersion() {
	org := "testorg"
	repo := "testrepo"
	acsVersion := "1.2.3"
	commitSHA := "abc123def456"

	helpers.MockLightweightTag(org, repo, acsVersion, commitSHA)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, acsVersion, "text")

	output, err := github.RunGetCommitFromACSVersion(cmd, []string{}, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyCommitOutput(output, org, repo, acsVersion, commitSHA)
	s.True(gock.IsDone())
}

func (s *GetCommitFromACSVersionTestSuite) TestGetCommitFromACSVersion_GitDescribeVersion() {
	org := "testorg"
	repo := "testrepo"
	acsVersion := "4.10.x-827-gf388f07b07"
	shortSHA := "f388f07b07"
	fullSHA := "f388f07b07abcdef1234567890abcdef12345678"

	helpers.MockGitCommitBySHA(org, repo, shortSHA, fullSHA)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, acsVersion, "text")

	output, err := github.RunGetCommitFromACSVersion(cmd, []string{}, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyCommitOutput(output, org, repo, acsVersion, fullSHA)
	s.True(gock.IsDone())
}

func (s *GetCommitFromACSVersionTestSuite) TestGetCommitFromACSVersion_NightlyTag() {
	org := "testorg"
	repo := "testrepo"
	acsVersion := "4.10.x-nightly-20260120"
	commitSHA := "def456ghi789"

	helpers.MockLightweightTag(org, repo, acsVersion, commitSHA)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, acsVersion, "text")

	output, err := github.RunGetCommitFromACSVersion(cmd, []string{}, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyCommitOutput(output, org, repo, acsVersion, commitSHA)
	s.True(gock.IsDone())
}

func (s *GetCommitFromACSVersionTestSuite) TestGetCommitFromACSVersion_VersionNotFound() {
	org := "testorg"
	repo := "testrepo"
	acsVersion := "nonexistent-version"

	helpers.MockTagNotFound(org, repo, acsVersion)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, acsVersion, "text")

	output, err := github.RunGetCommitFromACSVersion(cmd, []string{}, testClient)

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get commit from ACS version")
	s.True(gock.IsDone())
}

func (s *GetCommitFromACSVersionTestSuite) TestGetCommitFromACSVersion_InvalidGitDescribe() {
	org := "testorg"
	repo := "testrepo"
	acsVersion := "4.10.x-invalid-format"

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, acsVersion, "text")

	output, err := github.RunGetCommitFromACSVersion(cmd, []string{}, testClient)

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get commit from ACS version")
}

func (s *GetCommitFromACSVersionTestSuite) TestGetCommitFromACSVersion_InvalidRepository() {
	org := "nonexistent"
	repo := "nonexistent"
	acsVersion := "1.2.3"

	helpers.MockTagNotFound(org, repo, acsVersion)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, acsVersion, "text")

	output, err := github.RunGetCommitFromACSVersion(cmd, []string{}, testClient)

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get commit from ACS version")
	s.True(gock.IsDone())
}
