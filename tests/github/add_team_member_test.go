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

func TestAddTeamMemberTestSuite(t *testing.T) {
	suite.Run(t, new(AddTeamMemberTestSuite))
}

type AddTeamMemberTestSuite struct {
	suite.Suite
}

func (s *AddTeamMemberTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *AddTeamMemberTestSuite) createCommand(org, team, user, role, outputFormat string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "team", team))
	s.Require().NoError(helpers.SetFlag(cmd, "user", user))
	s.Require().NoError(helpers.SetFlag(cmd, "role", role))
	s.Require().NoError(helpers.SetFlag(cmd, "output", outputFormat))
	return cmd
}

func (s *AddTeamMemberTestSuite) TestAddTeamMember_Success() {
	org := "stackrox"
	team := "my-team"
	user := "octocat"

	helpers.MockAddTeamMembership(org, team, user, "member", "active")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, team, user, "member", "text")

	output, err := github.RunAddTeamMember(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	result, ok := output.(github.AddTeamMemberOutput)
	s.True(ok)
	s.Equal(user, result.Username)
	s.Equal(team, result.Team)
	s.Equal(org, result.Org)
	s.Equal("member", result.Role)
	s.Equal("active", result.State)
	s.True(gock.IsDone())
}

func (s *AddTeamMemberTestSuite) TestAddTeamMember_Pending() {
	org := "stackrox"
	team := "my-team"
	user := "newuser"

	helpers.MockAddTeamMembership(org, team, user, "member", "pending")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, team, user, "member", "text")

	output, err := github.RunAddTeamMember(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	result := output.(github.AddTeamMemberOutput)
	s.Equal("pending", result.State)
	s.True(gock.IsDone())
}

func (s *AddTeamMemberTestSuite) TestAddTeamMember_Maintainer() {
	org := "stackrox"
	team := "my-team"
	user := "octocat"

	helpers.MockAddTeamMembership(org, team, user, "maintainer", "active")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, team, user, "maintainer", "text")

	output, err := github.RunAddTeamMember(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	result := output.(github.AddTeamMemberOutput)
	s.Equal("maintainer", result.Role)
	s.True(gock.IsDone())
}

func (s *AddTeamMemberTestSuite) TestAddTeamMember_TeamNotFound() {
	org := "stackrox"
	team := "nonexistent-team"
	user := "octocat"

	helpers.MockAddTeamMembershipNotFound(org, team, user)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, team, user, "member", "text")

	output, err := github.RunAddTeamMember(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to add user")
	s.True(gock.IsDone())
}

func (s *AddTeamMemberTestSuite) TestAddTeamMember_Unprocessable() {
	org := "stackrox"
	team := "my-team"
	user := "invalid-user"

	helpers.MockAddTeamMembershipUnprocessable(org, team, user)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, team, user, "member", "text")

	output, err := github.RunAddTeamMember(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to add user")
	s.True(gock.IsDone())
}
