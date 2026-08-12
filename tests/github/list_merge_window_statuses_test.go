package github

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestListMergeWindowStatusesTestSuite(t *testing.T) {
	suite.Run(t, new(ListMergeWindowStatusesTestSuite))
}

type ListMergeWindowStatusesTestSuite struct {
	suite.Suite
}

func (s *ListMergeWindowStatusesTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

// createCommand creates and configures a mock command with the given flags
func (s *ListMergeWindowStatusesTestSuite) createCommand(org, repo string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	return cmd
}

// verifyListMergeWindowStatusesOutput verifies the output structure
func (s *ListMergeWindowStatusesTestSuite) verifyListMergeWindowStatusesOutput(
	output any,
	expectedStreams []string,
	expectedStatuses map[string]string,
) {
	listOutput, ok := output.(github.ListMergeWindowStatusesOutput)
	s.True(ok, "Output should be of type ListMergeWindowStatusesOutput")
	s.Equal(expectedStreams, listOutput.SupportedProductStreams)
	s.Equal(expectedStatuses, listOutput.MergeWindowStatuses)
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MultipleSupportedVersions() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with multiple supported versions
	versions := []helpers.ProductVersionData{
		{Name: "4.4", Type: "Full Support"},
		{Name: "4.5", Type: "Maintenance Support"},
		{Name: "4.6", Type: "Extended Support"},
		{Name: "4.7", Type: "Full Support"},
		{Name: "3.9", Type: "End of Life"}, // Should be filtered out
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	// Mock MERGE_WINDOW file for each supported version (sorted order)
	helpers.MockMergeWindowFile(org, repo, "4.4", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.5", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.6", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.7", "OPEN")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// Verify output - versions should be sorted
	expectedStreams := []string{"4.4", "4.5", "4.6", "4.7"}
	expectedStatuses := map[string]string{
		"4.4": "CLOSED",
		"4.5": "OPEN",
		"4.6": "CLOSED",
		"4.7": "OPEN",
	}
	s.verifyListMergeWindowStatusesOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_SingleVersion() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with single supported version
	versions := []helpers.ProductVersionData{
		{Name: "4.6", Type: "Full Support"},
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	// Mock MERGE_WINDOW file
	helpers.MockMergeWindowFile(org, repo, "4.6", "OPEN")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []string{"4.6"}
	expectedStatuses := map[string]string{
		"4.6": "OPEN",
	}
	s.verifyListMergeWindowStatusesOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_ProductLifecycleAPIError() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API error
	helpers.MockProductLifecycleAPIError(config.ProductName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "unexpected status code")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_NoProductFound() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with no products
	helpers.MockProductLifecycleAPINotFound(config.ProductName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "expected exactly 1 product")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MergeWindowFileNotFound() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with supported version
	versions := []helpers.ProductVersionData{
		{Name: "4.6", Type: "Full Support"},
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	// Mock MERGE_WINDOW file not found
	helpers.MockMergeWindowFileNotFound(org, repo, "4.6")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get contents of")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MixedStatuses() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with multiple versions
	versions := []helpers.ProductVersionData{
		{Name: "4.5", Type: "Maintenance Support"},
		{Name: "4.6", Type: "Full Support"},
		{Name: "4.7", Type: "Extended Support"},
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	// Mock MERGE_WINDOW files with mixed statuses
	helpers.MockMergeWindowFile(org, repo, "4.5", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.6", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.7", "CLOSED")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []string{"4.5", "4.6", "4.7"}
	expectedStatuses := map[string]string{
		"4.5": "CLOSED",
		"4.6": "OPEN",
		"4.7": "CLOSED",
	}
	s.verifyListMergeWindowStatusesOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_FiltersByVersionType() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with various version types
	versions := []helpers.ProductVersionData{
		{Name: "4.6", Type: "Full Support"},
		{Name: "4.5", Type: "Maintenance Support"},
		{Name: "4.4", Type: "Extended Support"},
		{Name: "4.3", Type: "End of Life"},        // Should be filtered out
		{Name: "4.2", Type: "Extended Life"},      // Should be filtered out
		{Name: "4.1", Type: "Product End of Life"}, // Should be filtered out
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	// Mock MERGE_WINDOW files only for supported versions
	helpers.MockMergeWindowFile(org, repo, "4.4", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.5", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.6", "OPEN")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// Only Full Support, Maintenance Support, and Extended Support should be included
	expectedStreams := []string{"4.4", "4.5", "4.6"}
	expectedStatuses := map[string]string{
		"4.4": "CLOSED",
		"4.5": "OPEN",
		"4.6": "OPEN",
	}
	s.verifyListMergeWindowStatusesOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_CustomOrgRepo() {
	org := "custom-org"
	repo := "custom-repo"

	// Mock Product Lifecycle API
	versions := []helpers.ProductVersionData{
		{Name: "1.0", Type: "Full Support"},
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	// Mock MERGE_WINDOW file
	helpers.MockMergeWindowFile(org, repo, "1.0", "OPEN")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []string{"1.0"}
	expectedStatuses := map[string]string{
		"1.0": "OPEN",
	}
	s.verifyListMergeWindowStatusesOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MultipleProductsReturned() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API returning multiple products (ambiguous)
	versions := []helpers.ProductVersionData{
		{Name: "4.6", Type: "Full Support"},
	}
	helpers.MockProductLifecycleAPIMultipleProducts(config.ProductName, versions)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "expected exactly 1 product, got 2")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MissingDataField() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with missing data field entirely
	helpers.MockProductLifecycleAPIMissingDataField(config.ProductName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "expected exactly 1 product, got 0")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_NullDataField() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with null data field
	helpers.MockProductLifecycleAPINullData(config.ProductName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "expected exactly 1 product, got 0")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MalformedJSON() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with malformed JSON
	helpers.MockProductLifecycleAPIMalformedJSON(config.ProductName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to unmarshal")
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_AllVersionsEOL() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Product Lifecycle API with only EOL versions (all filtered out)
	versions := []helpers.ProductVersionData{
		{Name: "4.1", Type: "End of Life"},
		{Name: "4.2", Type: "Product End of Life"},
		{Name: "4.3", Type: "Extended Life"},
	}
	helpers.MockProductLifecycleAPI(config.ProductName, versions)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// No versions should be returned as all are EOL
	expectedStreams := []string{}
	expectedStatuses := map[string]string{}
	s.verifyListMergeWindowStatusesOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}
