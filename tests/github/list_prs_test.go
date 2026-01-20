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

func TestListPRsTestSuite(t *testing.T) {
	suite.Run(t, new(ListPRsTestSuite))
}

type ListPRsTestSuite struct {
	suite.Suite
}

func (s *ListPRsTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

// createCommand creates and configures a mock command with the given flags
func (s *ListPRsTestSuite) createCommand(org, repo, baseBranch string, includeDrafts bool, outputFormat string) *cobra.Command {
	cmd := helpers.NewMockCommand()
	s.Require().NoError(helpers.SetFlag(cmd, "org", org))
	s.Require().NoError(helpers.SetFlag(cmd, "repo", repo))
	s.Require().NoError(helpers.SetFlag(cmd, "base-branch", baseBranch))
	s.Require().NoError(helpers.SetFlag(cmd, "output", outputFormat))
	s.Require().NoError(helpers.SetBoolFlag(cmd, "include-drafts", includeDrafts))
	return cmd
}

// verifyPRsOutput verifies the PRs output structure
func (s *ListPRsTestSuite) verifyPRsOutput(output any, org, repo, baseBranch string, expectedPRCount int, includeDrafts bool) {
	prsOutput, ok := output.(github.PRsOutput)
	s.True(ok)
	s.Equal(org, prsOutput.Organization)
	s.Equal(repo, prsOutput.Repository)
	s.Equal(baseBranch, prsOutput.BaseBranch)
	s.Equal(includeDrafts, prsOutput.IncludeDrafts)
	s.Len(prsOutput.PRs, expectedPRCount)
}

// PRStatusConfig configures the status, checks, and reviews for a PR
type PRStatusConfig struct {
	CombinedStatus string
	CheckRuns      []helpers.CheckRunData
	Reviews        []helpers.ReviewData
}

// mockPRWithStatus mocks all dependencies for a single PR (status, checks, reviews)
func (s *ListPRsTestSuite) mockPRWithStatus(org, repo string, pr helpers.PRData, status PRStatusConfig) {
	helpers.MockCombinedStatus(org, repo, pr.HeadSHA, status.CombinedStatus)
	helpers.MockCheckRuns(org, repo, pr.HeadSHA, status.CheckRuns)
	helpers.MockListReviews(org, repo, pr.Number, status.Reviews)
}

// mockPRsWithStatuses mocks PR list and all dependencies for multiple PRs
// statuses should be provided for all PRs that will be processed (non-drafts when includeDrafts=false, all when includeDrafts=true)
func (s *ListPRsTestSuite) mockPRsWithStatuses(org, repo, baseBranch string, prs []helpers.PRData, statuses []PRStatusConfig) {
	helpers.MockListPullRequests(org, repo, baseBranch, prs)
	for i, pr := range prs {
		if i < len(statuses) {
			s.mockPRWithStatus(org, repo, pr, statuses[i])
		}
	}
}

// defaultPRStatus returns a default status config (passed checks, approved review)
func (s *ListPRsTestSuite) defaultPRStatus() PRStatusConfig {
	return PRStatusConfig{
		CombinedStatus: "success",
		CheckRuns:      []helpers.CheckRunData{{Conclusion: "success", Status: "completed"}},
		Reviews:        []helpers.ReviewData{{State: "APPROVED"}},
	}
}

func (s *ListPRsTestSuite) TestListPRs_Success() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	prs := []helpers.PRData{
		{Number: 1, Title: "Fix bug", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
		{Number: 2, Title: "Add feature", Author: "user2", HeadSHA: "sha2", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/2"},
	}

	statuses := []PRStatusConfig{
		s.defaultPRStatus(),
		{
			CombinedStatus: "pending",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "", Status: "in_progress"}},
			Reviews:        []helpers.ReviewData{{State: "COMMENTED"}},
		},
	}

	s.mockPRsWithStatuses(org, repo, baseBranch, prs, statuses)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyPRsOutput(output, org, repo, baseBranch, 2, false)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_ExcludingDrafts() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	prs := []helpers.PRData{
		{Number: 1, Title: "Fix bug", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
		{Number: 2, Title: "Draft PR", Author: "user2", HeadSHA: "sha2", IsDraft: true, HTMLURL: "https://github.com/testorg/testrepo/pull/2"},
		{Number: 3, Title: "Another fix", Author: "user3", HeadSHA: "sha3", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/3"},
	}

	helpers.MockListPullRequests(org, repo, baseBranch, prs)
	s.mockPRWithStatus(org, repo, prs[0], s.defaultPRStatus()) // PR 0 (non-draft)
	// Skip mock for PR 1 because it is a draft and will be filtered out, so we don't need to mock its status
	s.mockPRWithStatus(org, repo, prs[2], s.defaultPRStatus()) // PR 2 (non-draft)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	// Should only have 2 PRs (draft filtered out)
	s.verifyPRsOutput(output, org, repo, baseBranch, 2, false)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_IncludingDrafts() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	prs := []helpers.PRData{
		{Number: 1, Title: "Fix bug", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
		{Number: 2, Title: "Draft PR", Author: "user2", HeadSHA: "sha2", IsDraft: true, HTMLURL: "https://github.com/testorg/testrepo/pull/2"},
	}

	statuses := []PRStatusConfig{
		s.defaultPRStatus(),
		s.defaultPRStatus(),
		{
			CombinedStatus: "pending",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "", Status: "in_progress"}},
			Reviews:        []helpers.ReviewData{},
		},
	}

	s.mockPRsWithStatuses(org, repo, baseBranch, prs, statuses)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, true, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	// Should have both PRs including draft
	s.verifyPRsOutput(output, org, repo, baseBranch, 2, true)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_EmptyResult() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	helpers.MockListPullRequests(org, repo, baseBranch, []helpers.PRData{})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyPRsOutput(output, org, repo, baseBranch, 0, false)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_SinglePR() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	prs := []helpers.PRData{
		{Number: 1, Title: "Single PR", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
	}

	s.mockPRsWithStatuses(org, repo, baseBranch, prs, []PRStatusConfig{s.defaultPRStatus()})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyPRsOutput(output, org, repo, baseBranch, 1, false)

	prsOutput := output.(github.PRsOutput)
	s.Equal(1, prsOutput.PRs[0].Number)
	s.Equal("Single PR", prsOutput.PRs[0].Title)
	s.Equal("user1", prsOutput.PRs[0].Author)
	s.Equal("Approved", prsOutput.PRs[0].ReviewStatus)
	s.Equal("Passed", prsOutput.PRs[0].ChecksStatus)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_CheckStatuses() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	prs := []helpers.PRData{
		{Number: 1, Title: "Passed checks", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
		{Number: 2, Title: "Failed checks", Author: "user2", HeadSHA: "sha2", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/2"},
		{Number: 3, Title: "Pending checks", Author: "user3", HeadSHA: "sha3", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/3"},
		{Number: 4, Title: "No checks", Author: "user4", HeadSHA: "sha4", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/4"},
	}

	statuses := []PRStatusConfig{
		// PR 1: Passed
		s.defaultPRStatus(),
		// PR 2: Failed
		{
			CombinedStatus: "failure",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "failure", Status: "completed"}},
			Reviews:        []helpers.ReviewData{{State: "CHANGES_REQUESTED"}},
		},
		// PR 3: Pending
		{
			CombinedStatus: "pending",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "", Status: "in_progress"}},
			Reviews:        []helpers.ReviewData{{State: "COMMENTED"}},
		},
		// PR 4: No checks
		{
			CombinedStatus: "",
			CheckRuns:      []helpers.CheckRunData{},
			Reviews:        []helpers.ReviewData{},
		},
	}

	s.mockPRsWithStatuses(org, repo, baseBranch, prs, statuses)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyPRsOutput(output, org, repo, baseBranch, 4, false)

	prsOutput := output.(github.PRsOutput)
	s.Equal("Passed", prsOutput.PRs[0].ChecksStatus)
	s.Equal("Failed", prsOutput.PRs[1].ChecksStatus)
	s.Equal("Pending", prsOutput.PRs[2].ChecksStatus)
	s.Equal("No Checks", prsOutput.PRs[3].ChecksStatus)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_ReviewStatuses() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	prs := []helpers.PRData{
		{Number: 1, Title: "Approved", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
		{Number: 2, Title: "Changes requested", Author: "user2", HeadSHA: "sha2", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/2"},
		{Number: 3, Title: "Commented", Author: "user3", HeadSHA: "sha3", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/3"},
		{Number: 4, Title: "No reviews", Author: "user4", HeadSHA: "sha4", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/4"},
	}

	statuses := []PRStatusConfig{
		// PR 1: Approved
		{
			CombinedStatus: "success",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "success", Status: "completed"}},
			Reviews:        []helpers.ReviewData{{State: "APPROVED"}},
		},
		// PR 2: Changes requested
		{
			CombinedStatus: "success",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "success", Status: "completed"}},
			Reviews:        []helpers.ReviewData{{State: "CHANGES_REQUESTED"}},
		},
		// PR 3: Commented
		{
			CombinedStatus: "success",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "success", Status: "completed"}},
			Reviews:        []helpers.ReviewData{{State: "COMMENTED"}},
		},
		// PR 4: No reviews
		{
			CombinedStatus: "success",
			CheckRuns:      []helpers.CheckRunData{{Conclusion: "success", Status: "completed"}},
			Reviews:        []helpers.ReviewData{},
		},
	}

	s.mockPRsWithStatuses(org, repo, baseBranch, prs, statuses)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyPRsOutput(output, org, repo, baseBranch, 4, false)

	prsOutput := output.(github.PRsOutput)
	s.Equal("Approved", prsOutput.PRs[0].ReviewStatus)
	s.Equal("Changes Requested", prsOutput.PRs[1].ReviewStatus)
	s.Equal("Commented", prsOutput.PRs[2].ReviewStatus)
	s.Equal("No Reviews", prsOutput.PRs[3].ReviewStatus)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_Pagination() {
	org := "testorg"
	repo := "testrepo"
	baseBranch := "main"

	page1PRs := []helpers.PRData{
		{Number: 1, Title: "PR 1", Author: "user1", HeadSHA: "sha1", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/1"},
	}
	page2PRs := []helpers.PRData{
		{Number: 2, Title: "PR 2", Author: "user2", HeadSHA: "sha2", IsDraft: false, HTMLURL: "https://github.com/testorg/testrepo/pull/2"},
	}

	helpers.MockListPullRequestsPaginated(org, repo, baseBranch, page1PRs, page2PRs)

	// Mock statuses for both PRs
	s.mockPRWithStatus(org, repo, page1PRs[0], s.defaultPRStatus())
	s.mockPRWithStatus(org, repo, page2PRs[0], s.defaultPRStatus())

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.NoError(err)
	s.NotNil(output)
	s.verifyPRsOutput(output, org, repo, baseBranch, 2, false)
	s.True(gock.IsDone())
}

func (s *ListPRsTestSuite) TestListPRs_InvalidRepository() {
	org := "nonexistent"
	repo := "nonexistent"
	baseBranch := "main"

	helpers.MockRepositoryNotFound(org, repo)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	cmd := s.createCommand(org, repo, baseBranch, false, "text")

	output, err := github.RunListPRs(cmd, []string{}, testClient, context.Background())

	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to list pull requests")
	s.True(gock.IsDone())
}
