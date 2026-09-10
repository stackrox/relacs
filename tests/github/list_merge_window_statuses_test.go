package github

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/github"
	"github.com/stackrox/relacs/pkg/client/konflux"
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

// mustParseProductStream parses a version string into a ProductStream, failing the test if parsing fails
func (s *ListMergeWindowStatusesTestSuite) mustParseProductStream(version string) konflux.ProductStream {
	ps, err := konflux.NewProductStream(version)
	s.Require().NoError(err)
	return ps
}

// verifyListMergeWindowsOutput verifies the output structure
func (s *ListMergeWindowStatusesTestSuite) verifyListMergeWindowsOutput(
	output any,
	expectedStreams []konflux.ProductStream,
	expectedStatuses map[konflux.ProductStream]string,
) {
	listOutput, ok := output.(github.ListMergeWindowsOutput)
	s.True(ok, "Output should be of type ListMergeWindowsOutput")
	s.Equal(expectedStreams, listOutput.SupportedProductStreams)
	s.Equal(expectedStatuses, listOutput.MergeWindowStatuses)
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MultipleSupportedVersions() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications with multiple supported versions
	apps := []string{"acs-4-4", "acs-4-5", "acs-4-6", "acs-4-7", "other-app"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW file for each supported version (sorted order)
	helpers.MockMergeWindowFile(org, repo, "4.4", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.5", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.6", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.7", "OPEN")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// Verify output - versions should be sorted
	expectedStreams := []konflux.ProductStream{
		s.mustParseProductStream("4.4"),
		s.mustParseProductStream("4.5"),
		s.mustParseProductStream("4.6"),
		s.mustParseProductStream("4.7"),
	}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.4"): "CLOSED",
		s.mustParseProductStream("4.5"): "OPEN",
		s.mustParseProductStream("4.6"): "CLOSED",
		s.mustParseProductStream("4.7"): "OPEN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_SingleVersion() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications with single supported version
	apps := []string{"acs-4-6"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW file
	helpers.MockMergeWindowFile(org, repo, "4.6", "OPEN")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []konflux.ProductStream{s.mustParseProductStream("4.6")}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.6"): "OPEN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_NoApplications() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux with no acs applications
	apps := []string{"other-app", "another-app"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// No acs-* applications found, so empty list
	expectedStreams := []konflux.ProductStream{}
	expectedStatuses := map[konflux.ProductStream]string{}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	// No gock mocks since no MERGE_WINDOW files should be fetched
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MergeWindowFileNotFound_ReturnsUnknown() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications
	apps := []string{"acs-4-6"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW file not found (404) - should return UNKNOWN
	helpers.MockMergeWindowFileNotFound(org, repo, "4.6")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err) // Should not error, should return UNKNOWN
	s.NotNil(output)

	expectedStreams := []konflux.ProductStream{s.mustParseProductStream("4.6")}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.6"): "UNKNOWN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_MixedStatuses() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications with multiple versions
	apps := []string{"acs-4-5", "acs-4-6", "acs-4-7"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW files with mixed statuses
	helpers.MockMergeWindowFile(org, repo, "4.5", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.6", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.7", "CLOSED")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []konflux.ProductStream{
		s.mustParseProductStream("4.5"),
		s.mustParseProductStream("4.6"),
		s.mustParseProductStream("4.7"),
	}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.5"): "CLOSED",
		s.mustParseProductStream("4.6"): "OPEN",
		s.mustParseProductStream("4.7"): "CLOSED",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_FiltersByApplicationPattern() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications with various patterns
	// Only acs-<major>-<minor> should be matched
	apps := []string{
		"acs-4-6",
		"acs-4-5",
		"acs-5-0",      // Future version
		"acs",          // Should not match (no version)
		"acs-scanner",  // Should not match (not version pattern)
		"other-4-6",    // Should not match (not acs)
		"acs-4-6-test", // Should not match (extra suffix)
	}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW files only for valid acs applications
	helpers.MockMergeWindowFile(org, repo, "4.5", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.6", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "5.0", "OPEN")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// Only acs-4-5, acs-4-6, acs-5-0 should be included (sorted)
	expectedStreams := []konflux.ProductStream{
		s.mustParseProductStream("4.5"),
		s.mustParseProductStream("4.6"),
		s.mustParseProductStream("5.0"),
	}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.5"): "CLOSED",
		s.mustParseProductStream("4.6"): "OPEN",
		s.mustParseProductStream("5.0"): "OPEN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_VersionSorting() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications in random order
	apps := []string{"acs-4-10", "acs-5-0", "acs-4-5", "acs-4-9", "acs-3-74"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW files
	helpers.MockMergeWindowFile(org, repo, "3.74", "CLOSED")
	helpers.MockMergeWindowFile(org, repo, "4.5", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.9", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "4.10", "OPEN")
	helpers.MockMergeWindowFile(org, repo, "5.0", "OPEN")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	// Should be sorted by semantic version (3.74, 4.5, 4.9, 4.10, 5.0)
	expectedStreams := []konflux.ProductStream{
		s.mustParseProductStream("3.74"),
		s.mustParseProductStream("4.5"),
		s.mustParseProductStream("4.9"),
		s.mustParseProductStream("4.10"),
		s.mustParseProductStream("5.0"),
	}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("3.74"): "CLOSED",
		s.mustParseProductStream("4.5"):  "OPEN",
		s.mustParseProductStream("4.9"):  "OPEN",
		s.mustParseProductStream("4.10"): "OPEN",
		s.mustParseProductStream("5.0"):  "OPEN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_SomeMissingFiles() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications
	apps := []string{"acs-4-5", "acs-4-6", "acs-5-0"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW files - some exist, some don't
	helpers.MockMergeWindowFile(org, repo, "4.5", "CLOSED")
	helpers.MockMergeWindowFileNotFound(org, repo, "4.6") // Returns UNKNOWN
	helpers.MockMergeWindowFile(org, repo, "5.0", "OPEN")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []konflux.ProductStream{
		s.mustParseProductStream("4.5"),
		s.mustParseProductStream("4.6"),
		s.mustParseProductStream("5.0"),
	}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.5"): "CLOSED",
		s.mustParseProductStream("4.6"): "UNKNOWN",
		s.mustParseProductStream("5.0"): "OPEN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_CustomOrgRepo() {
	org := "custom-org"
	repo := "custom-repo"

	// Mock Konflux applications
	apps := []string{"acs-1-0"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// Mock MERGE_WINDOW file
	helpers.MockMergeWindowFile(org, repo, "1.0", "OPEN")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []konflux.ProductStream{s.mustParseProductStream("1.0")}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("1.0"): "OPEN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}

func (s *ListMergeWindowStatusesTestSuite) TestListMergeWindowStatuses_AllUnknown() {
	org := "stackrox"
	repo := "stackrox"

	// Mock Konflux applications
	apps := []string{"acs-4-5", "acs-4-6"}
	testClient, err := helpers.SetupTestClientWithKonfluxApps(apps)
	s.Require().NoError(err)

	// All MERGE_WINDOW files missing
	helpers.MockMergeWindowFileNotFound(org, repo, "4.5")
	helpers.MockMergeWindowFileNotFound(org, repo, "4.6")

	cmd := s.createCommand(org, repo)

	output, err := github.RunListMergeWindowStatuses(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)

	expectedStreams := []konflux.ProductStream{
		s.mustParseProductStream("4.5"),
		s.mustParseProductStream("4.6"),
	}
	expectedStatuses := map[konflux.ProductStream]string{
		s.mustParseProductStream("4.5"): "UNKNOWN",
		s.mustParseProductStream("4.6"): "UNKNOWN",
	}
	s.verifyListMergeWindowsOutput(output, expectedStreams, expectedStatuses)
	s.True(gock.IsDone())
}
