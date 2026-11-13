package konflux

import (
	"context"
	"strings"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
	relacskonflux "github.com/stackrox/relacs/pkg/konflux"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

const (
	conformaTestTag = "4.10.0"
	conformaTestSHA = "abc123def456"
)

func TestConformaViolationsTestSuite(t *testing.T) {
	suite.Run(t, new(ConformaViolationsTestSuite))
}

type ConformaViolationsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *ConformaViolationsTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ConformaViolationsTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

// conformaPipelineRun builds the JSON of a conforma pipeline run for the test
// tag. Passing nil conditions means the run has not finished yet.
func (s *ConformaViolationsTestSuite) conformaPipelineRun(name, uid, component string, conditions []map[string]interface{}) map[string]interface{} {
	if conditions == nil {
		conditions = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              name,
			"uid":               uid,
			"creationTimestamp": "2025-01-01T00:00:00Z",
			"labels": map[string]interface{}{
				"appstudio.openshift.io/application":   "acs-4-10",
				"appstudio.openshift.io/component":     component,
				"test.appstudio.openshift.io/scenario": "acs-conforma-dev-4-10",
			},
			"annotations": map[string]interface{}{
				"pac.test.appstudio.openshift.io/branch": "refs/tags/" + conformaTestTag,
			},
		},
		"status": map[string]interface{}{
			"conditions": conditions,
		},
	}
}

func (s *ConformaViolationsTestSuite) succeededConditions() []map[string]interface{} {
	return []map[string]interface{}{{"type": "Succeeded", "status": "True", "reason": "Completed"}}
}

func (s *ConformaViolationsTestSuite) failedConditions() []map[string]interface{} {
	return []map[string]interface{}{{"type": "Succeeded", "status": "False", "reason": "Failed"}}
}

// startedConditions is how Tekton marks a pipeline run that has started but
// not finished: a Succeeded condition with unknown status.
func (s *ConformaViolationsTestSuite) startedConditions() []map[string]interface{} {
	return []map[string]interface{}{{"type": "Succeeded", "status": "Unknown", "reason": "Started"}}
}

func (s *ConformaViolationsTestSuite) mockConformaPipelineRuns(runs ...map[string]interface{}) {
	helpers.MockLightweightTag(config.GitHubOrganization, config.MainGitHubRepository, conformaTestTag, conformaTestSHA)

	labels := strings.Join([]string{
		"pac.test.appstudio.openshift.io/sha=" + conformaTestSHA,
		"pac.test.appstudio.openshift.io/url-org=" + config.GitHubOrganization,
		"pac.test.appstudio.openshift.io/url-repository=" + config.MainGitHubRepository,
		"pipelines.appstudio.openshift.io/type=test",
		"build.appstudio.redhat.com/pipeline=enterprise-contract",
	}, ",")

	helpers.MockK8sPipelineRuns(labels, map[string]interface{}{
		"items": runs,
	})
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant)
}

func (s *ConformaViolationsTestSuite) mockConformaReport(runName string) {
	helpers.MockGetPodLogs(helpers.TestK8sAPIServerURL, config.KonfluxTenant, runName+"-verify-pod", "step-report-json",
		`{"components": [{"name": "main", "violations": [{"msg": "violation one"}]}]}`)
}

func (s *ConformaViolationsTestSuite) mockConformaReportUnavailable(runName string) {
	helpers.MockGetPodLogsNotFound(helpers.TestK8sAPIServerURL, config.KonfluxTenant, runName+"-verify-pod")
	helpers.MockGetPodLogsNotFound(config.KubearchiveAPIURL, config.KonfluxTenant, runName+"-verify-pod")
}

func (s *ConformaViolationsTestSuite) listConformaViolations() (relacskonflux.ConformaViolations, error) {
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	return relacskonflux.ListConformaViolations(s.ctx, testClient, config.GitHubOrganization, config.MainGitHubRepository, conformaTestTag, relacskonflux.ConformaScenarioDev)
}

// mockProdLikeCustomSnapshot mocks the custom snapshot for the test tag and
// returns its name. The snapshot's application label is versioned, from which
// the versioned prod-like scenario name is derived.
func (s *ConformaViolationsTestSuite) mockProdLikeCustomSnapshot() string {
	helpers.MockLightweightTag(config.GitHubOrganization, config.MainGitHubRepository, conformaTestTag, conformaTestSHA)

	labels := strings.Join([]string{
		"pac.test.appstudio.openshift.io/url-org=" + config.GitHubOrganization,
		"pac.test.appstudio.openshift.io/url-repository=" + config.MainGitHubRepository,
		"pac.test.appstudio.openshift.io/sha=" + conformaTestSHA,
		"acs-style-snapshot=true",
		"acs-product-version=" + conformaTestTag,
		"appstudio.openshift.io/application=acs-4-10",
	}, ",")

	helpers.MockK8sSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name":              "acs-4-10-custom-snapshot",
					"uid":               "uid-snapshot",
					"creationTimestamp": "2025-01-01T00:00:00Z",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-10",
					},
				},
			},
		},
	})
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant)
	return "acs-4-10-custom-snapshot"
}

// mockProdLikeConformaRuns mocks the prod-like conforma pipeline runs attached
// to the given custom snapshot.
func (s *ConformaViolationsTestSuite) mockProdLikeConformaRuns(snapshotName string, runs ...map[string]interface{}) {
	labels := strings.Join([]string{
		"appstudio.openshift.io/snapshot=" + snapshotName,
		"test.appstudio.openshift.io/scenario=acs-conforma-prod-like-4-10",
	}, ",")

	helpers.MockK8sPipelineRuns(labels, map[string]interface{}{
		"items": runs,
	})
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant)
}

func (s *ConformaViolationsTestSuite) prodLikeConformaPipelineRun(name, uid, component string, conditions []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              name,
			"uid":               uid,
			"creationTimestamp": "2025-01-01T00:00:00Z",
			"labels": map[string]interface{}{
				"appstudio.openshift.io/component":     component,
				"test.appstudio.openshift.io/scenario": "acs-conforma-prod-like-4-10",
			},
		},
		"status": map[string]interface{}{
			"conditions": conditions,
		},
	}
}

func (s *ConformaViolationsTestSuite) listConformaViolationsForScenario(repo, scenario string) (relacskonflux.ConformaViolations, error) {
	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	return relacskonflux.ListConformaViolations(s.ctx, testClient, config.GitHubOrganization, repo, conformaTestTag, scenario)
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_ProdLike() {
	snapshotName := s.mockProdLikeCustomSnapshot()
	run := s.prodLikeConformaPipelineRun("acs-conforma-prod-like-4-10-abc12", "uid-prod-like", "main-4-10", s.succeededConditions())
	s.mockProdLikeConformaRuns(snapshotName, run)
	s.mockConformaReport("acs-conforma-prod-like-4-10-abc12")

	result, err := s.listConformaViolationsForScenario(config.MainGitHubRepository, relacskonflux.ConformaScenarioProdLike)

	s.Require().NoError(err)
	s.Len(result.Violations.Violations, 1)
	s.Equal("violation one", result.Violations.Violations[0].Violation)
	s.Empty(result.SkippedRuns)
	s.Empty(result.InProgressRuns)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_ProdLikeNonMainRepo() {
	_, err := s.listConformaViolationsForScenario("collector", relacskonflux.ConformaScenarioProdLike)

	s.Require().Error(err)
	s.Contains(err.Error(), "only supported for the stackrox repository")
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_UnknownScenario() {
	_, err := s.listConformaViolationsForScenario(config.MainGitHubRepository, "acs-conforma-other")

	s.Require().Error(err)
	s.Contains(err.Error(), "unknown conforma scenario")
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaRuns_ProdLike() {
	snapshotName := s.mockProdLikeCustomSnapshot()
	run := s.prodLikeConformaPipelineRun("acs-conforma-prod-like-4-10-abc12", "uid-prod-like", "main-4-10", s.succeededConditions())
	s.mockProdLikeConformaRuns(snapshotName, run)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)
	entries, err := relacskonflux.ListConformaRuns(s.ctx, testClient, config.GitHubOrganization, config.MainGitHubRepository, conformaTestTag, relacskonflux.ConformaScenarioProdLike)

	s.Require().NoError(err)
	s.Require().Len(entries, 1)
	s.Equal("acs-conforma-prod-like-4-10-abc12", entries[0].Name)
	s.Equal("main-4-10", entries[0].Component)
	s.Equal("acs-conforma-prod-like-4-10", entries[0].Scenario)
	s.Equal("Succeeded", entries[0].Status)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_InProgress() {
	// Tekton reports a started run as a Succeeded condition with unknown
	// status, not as an absent condition.
	started := s.conformaPipelineRun("acs-conforma-dev-4-10-started", "uid-started", "main-4-10", s.startedConditions())
	s.mockConformaPipelineRuns(started)

	result, err := s.listConformaViolations()

	s.Require().Error(err)
	s.Contains(err.Error(), "still in progress")
	s.Empty(result.Violations.Violations)
	s.Empty(result.SkippedRuns)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_PartialInProgress() {
	succeeded := s.conformaPipelineRun("acs-conforma-dev-4-10-done", "uid-done", "main-4-10", s.succeededConditions())
	running := s.conformaPipelineRun("acs-conforma-dev-4-10-running", "uid-running", "scanner-4-10", nil)
	s.mockConformaPipelineRuns(succeeded, running)
	s.mockConformaReport("acs-conforma-dev-4-10-done")

	result, err := s.listConformaViolations()

	s.Require().NoError(err)
	s.Len(result.Violations.Violations, 1)
	s.Equal("violation one", result.Violations.Violations[0].Violation)
	s.Empty(result.SkippedRuns)
	s.Equal([]string{"acs-conforma-dev-4-10-running"}, result.InProgressRuns)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_PartialSkippedAndInProgress() {
	succeeded := s.conformaPipelineRun("acs-conforma-dev-4-10-done", "uid-done", "main-4-10", s.succeededConditions())
	failed := s.conformaPipelineRun("acs-conforma-dev-4-10-failed", "uid-failed", "scanner-4-10", s.failedConditions())
	running := s.conformaPipelineRun("acs-conforma-dev-4-10-running", "uid-running", "collector-4-10", nil)
	s.mockConformaPipelineRuns(succeeded, failed, running)
	s.mockConformaReport("acs-conforma-dev-4-10-done")
	s.mockConformaReportUnavailable("acs-conforma-dev-4-10-failed")

	result, err := s.listConformaViolations()

	s.Require().NoError(err)
	s.Len(result.Violations.Violations, 1)
	s.Equal([]string{"acs-conforma-dev-4-10-failed"}, result.SkippedRuns)
	s.Equal([]string{"acs-conforma-dev-4-10-running"}, result.InProgressRuns)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_AllSkipped() {
	failed := s.conformaPipelineRun("acs-conforma-dev-4-10-failed", "uid-failed", "main-4-10", s.failedConditions())
	s.mockConformaPipelineRuns(failed)
	s.mockConformaReportUnavailable("acs-conforma-dev-4-10-failed")

	result, err := s.listConformaViolations()

	s.Require().Error(err)
	s.Contains(err.Error(), "failed to retrieve any conforma report")
	s.Empty(result.Violations.Violations)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_UnparseableReport() {
	good := s.conformaPipelineRun("acs-conforma-dev-4-10-done", "uid-done", "main-4-10", s.succeededConditions())
	bad := s.conformaPipelineRun("acs-conforma-dev-4-10-empty", "uid-empty", "scanner-4-10", s.succeededConditions())
	s.mockConformaPipelineRuns(good, bad)
	s.mockConformaReport("acs-conforma-dev-4-10-done")
	// An empty log body parses as invalid JSON, e.g. when the report step
	// produced no output.
	helpers.MockGetPodLogs(helpers.TestK8sAPIServerURL, config.KonfluxTenant, "acs-conforma-dev-4-10-empty-verify-pod", "step-report-json", "")

	result, err := s.listConformaViolations()

	s.Require().NoError(err)
	s.Len(result.Violations.Violations, 1)
	s.Equal([]string{"acs-conforma-dev-4-10-empty"}, result.SkippedRuns)
	s.Empty(result.InProgressRuns)
	s.True(gock.IsDone())
}

func (s *ConformaViolationsTestSuite) TestListConformaViolations_AllSucceeded() {
	succeeded := s.conformaPipelineRun("acs-conforma-dev-4-10-done", "uid-done", "main-4-10", s.succeededConditions())
	s.mockConformaPipelineRuns(succeeded)
	s.mockConformaReport("acs-conforma-dev-4-10-done")

	result, err := s.listConformaViolations()

	s.Require().NoError(err)
	s.Len(result.Violations.Violations, 1)
	s.Empty(result.SkippedRuns)
	s.Empty(result.InProgressRuns)
	s.True(gock.IsDone())
}
