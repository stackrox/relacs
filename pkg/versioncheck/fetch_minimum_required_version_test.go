package versioncheck

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

const (
	testOrg     = "stackrox"
	testRepo    = "relacs"
	testBranch  = "main"
	testPath    = "MINIMUM_REQUIRED_VERSION"
	testVersion = "v0.0.9"
)

func TestFetchMinimumRequiredVersionTestSuite(t *testing.T) {
	suite.Run(t, new(FetchMinimumRequiredVersionTestSuite))
}

type FetchMinimumRequiredVersionTestSuite struct {
	suite.Suite
}

func (s *FetchMinimumRequiredVersionTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *FetchMinimumRequiredVersionTestSuite) TestFetchMinimumRequiredVersion_Success() {
	// Base64 encode the version content (GitHub API returns base64 encoded content)
	encodedContent := base64.StdEncoding.EncodeToString([]byte(testVersion))

	helpers.MockGetFileContents(testOrg, testRepo, testPath, testBranch, encodedContent)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	fetchedVersion, err := fetchMinimumRequiredVersion(context.Background(), testClient)

	s.NoError(err)
	s.Equal(testVersion, fetchedVersion)
	s.True(gock.IsDone())
}

func (s *FetchMinimumRequiredVersionTestSuite) TestFetchMinimumRequiredVersion_WithWhitespace() {
	versionWithWhitespace := "  " + testVersion + "\n  "

	// Base64 encode the version content with whitespace
	encodedContent := base64.StdEncoding.EncodeToString([]byte(versionWithWhitespace))

	helpers.MockGetFileContents(testOrg, testRepo, testPath, testBranch, encodedContent)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	fetchedVersion, err := fetchMinimumRequiredVersion(context.Background(), testClient)

	s.NoError(err)
	s.Equal(testVersion, fetchedVersion) // Should be trimmed
	s.True(gock.IsDone())
}

func (s *FetchMinimumRequiredVersionTestSuite) TestFetchMinimumRequiredVersion_FileNotFound() {
	helpers.MockFileNotFound(testOrg, testRepo, testPath, testBranch)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	fetchedVersion, err := fetchMinimumRequiredVersion(context.Background(), testClient)

	s.Error(err)
	s.Contains(err.Error(), "failed to fetch")
	s.Empty(fetchedVersion)
	s.True(gock.IsDone())
}

func (s *FetchMinimumRequiredVersionTestSuite) TestFetchMinimumRequiredVersion_EmptyFile() {
	// Base64 encode empty content
	encodedContent := base64.StdEncoding.EncodeToString([]byte(""))

	helpers.MockGetFileContents(testOrg, testRepo, testPath, testBranch, encodedContent)
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	fetchedVersion, err := fetchMinimumRequiredVersion(context.Background(), testClient)

	s.Error(err)
	s.Contains(err.Error(), "file is empty")
	s.Empty(fetchedVersion)
	s.True(gock.IsDone())
}
