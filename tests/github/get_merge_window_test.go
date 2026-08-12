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

func TestGetMergeWindowTestSuite(t *testing.T) {
	suite.Run(t, new(GetMergeWindowTestSuite))
}

type GetMergeWindowTestSuite struct {
	suite.Suite
}

func (s *GetMergeWindowTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

// createCommand creates and configures a mock command with the given flags
func (s *GetMergeWindowTestSuite) createCommand(org, repo, productStream string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	s.Require().NoError(helpers.SetFlag(cmd, "product-stream", productStream))
	return cmd
}

// verifyMergeWindowOutput verifies the merge window output structure
func (s *GetMergeWindowTestSuite) verifyMergeWindowOutput(output any, productStream, expectedStatus string) {
	mergeWindowOutput, ok := output.(github.GetMergeWindowOutput)
	s.True(ok, "Output should be of type GetMergeWindowOutput")
	s.Equal(productStream, mergeWindowOutput.ProductStream)
	s.Equal(expectedStatus, mergeWindowOutput.Status)
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_OpenStatus() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.6"
	status := "OPEN"

	helpers.MockMergeWindowFile(org, repo, productStream, status)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyMergeWindowOutput(output, productStream, status)
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_ClosedStatus() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.5"
	status := "CLOSED"

	helpers.MockMergeWindowFile(org, repo, productStream, status)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyMergeWindowOutput(output, productStream, status)
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_FileNotFound() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.99"

	helpers.MockMergeWindowFileNotFound(org, repo, productStream)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get contents of")
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_InvalidStatus() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.6"
	invalidStatus := "INVALID_STATUS"

	helpers.MockMergeWindowFile(org, repo, productStream, invalidStatus)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "not a valid status")
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_CustomOrgRepo() {
	org := "custom-org"
	repo := "custom-repo"
	productStream := "4.7"
	status := "OPEN"

	helpers.MockMergeWindowFile(org, repo, productStream, status)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyMergeWindowOutput(output, productStream, status)
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_EmptyFile() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.6"

	helpers.MockMergeWindowFileEmpty(org, repo, productStream)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "not a valid status")
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_WhitespaceHandling() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.6"
	status := "OPEN"

	// Mock file with extra whitespace around status
	helpers.MockMergeWindowFileWithWhitespace(org, repo, productStream, status)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	// Should succeed because code trims whitespace
	s.verifyMergeWindowOutput(output, productStream, status)
	s.True(gock.IsDone())
}

func (s *GetMergeWindowTestSuite) TestGetMergeWindow_CaseSensitive() {
	org := "stackrox"
	repo := "stackrox"
	productStream := "4.6"
	invalidStatus := "open" // lowercase should be invalid

	helpers.MockMergeWindowFile(org, repo, productStream, invalidStatus)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, productStream)

	output, err := github.RunGetMergeWindow(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "not a valid status")
	s.True(gock.IsDone())
}
