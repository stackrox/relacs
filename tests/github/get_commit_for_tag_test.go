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

func TestGetCommitForTagTestSuite(t *testing.T) {
	suite.Run(t, new(GetCommitForTagTestSuite))
}

type GetCommitForTagTestSuite struct {
	suite.Suite
}

func (s *GetCommitForTagTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

// createCommand creates and configures a mock command with the given flags
func (s *GetCommitForTagTestSuite) createCommand(org, repo, tag, outputFormat string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	s.Require().NoError(helpers.SetFlag(cmd, "tag", tag))
	s.Require().NoError(helpers.SetFlag(cmd, "output", outputFormat))
	return cmd
}

// verifyCommitOutput verifies the commit output structure
func (s *GetCommitForTagTestSuite) verifyCommitOutput(output any, org, repo, tag, commitSHA string) {
	commitOutput, ok := output.(github.CommitOutput)
	s.True(ok)
	s.Equal(org, commitOutput.Organization)
	s.Equal(repo, commitOutput.Repository)
	s.Equal(tag, commitOutput.Tag)
	s.Equal(commitSHA, commitOutput.Commit.SHA)
	s.Equal(fmt.Sprintf("https://github.com/%s/%s/commit/%s", org, repo, commitSHA), commitOutput.URL)
}

func (s *GetCommitForTagTestSuite) TestGetCommitForTag_LightweightTag() {
	org := "testorg"
	repo := "testrepo"
	tag := "v1.0.0"
	commitSHA := "abc123def456"

	helpers.MockLightweightTag(org, repo, tag, commitSHA)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, tag, "text")

	output, err := github.RunGetCommitForTag(cmd, []string{}, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyCommitOutput(output, org, repo, tag, commitSHA)
	s.True(gock.IsDone())
}

func (s *GetCommitForTagTestSuite) TestGetCommitForTag_AnnotatedTag() {
	org := "testorg"
	repo := "testrepo"
	tag := "v1.0.0"
	tagObjectSHA := "tag123"
	commitSHA := "abc123def456"

	helpers.MockAnnotatedTag(org, repo, tag, tagObjectSHA, commitSHA)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, tag, "text")

	output, err := github.RunGetCommitForTag(cmd, []string{}, testClient)

	s.NoError(err)
	s.NotNil(output)
	s.verifyCommitOutput(output, org, repo, tag, commitSHA)
	s.True(gock.IsDone())
}

func (s *GetCommitForTagTestSuite) TestGetCommitForTag_TagNotFound() {
	org := "testorg"
	repo := "testrepo"
	tag := "nonexistent"

	helpers.MockTagNotFound(org, repo, tag)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, tag, "text")

	output, err := github.RunGetCommitForTag(cmd, []string{}, testClient)

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get commit for tag")
	s.True(gock.IsDone())
}

func (s *GetCommitForTagTestSuite) TestGetCommitForTag_InvalidRepository() {
	org := "nonexistent"
	repo := "nonexistent"
	tag := "v1.0.0"

	helpers.MockTagNotFound(org, repo, tag)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, tag, "text")

	output, err := github.RunGetCommitForTag(cmd, []string{}, testClient)

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get commit for tag")
	s.True(gock.IsDone())
}
