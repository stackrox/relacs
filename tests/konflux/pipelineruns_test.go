package konflux

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestPipelineRunsTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineRunsTestSuite))
}

type PipelineRunsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *PipelineRunsTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *PipelineRunsTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *PipelineRunsTestSuite) TestGetPipelineRuns_FromBothSources() {
	labels := "pac.test.appstudio.openshift.io/sha=abc123"

	helpers.MockK8sPipelineRuns(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "run-live",
					"uid":  "uid-live",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Succeeded", "status": "True", "reason": "Completed"},
					},
				},
			},
		},
	})

	helpers.MockKubearchivePipelineRuns(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "run-archived",
					"uid":  "uid-archived",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Succeeded", "status": "True", "reason": "Completed"},
					},
				},
			},
		},
	})

	testClient := helpers.SetupTestClient()

	runs, err := testClient.GetKonflux().GetPipelineRuns(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(runs, 2)
	s.Equal("run-live", runs[0].Metadata.Name)
	s.Equal("run-archived", runs[1].Metadata.Name)
	s.True(gock.IsDone())
}

func (s *PipelineRunsTestSuite) TestGetPipelineRuns_DeduplicatesByUID() {
	labels := "app=test"

	helpers.MockK8sPipelineRuns(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "run-1",
					"uid":  "same-uid",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{},
				},
			},
		},
	})

	helpers.MockKubearchivePipelineRuns(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "run-1",
					"uid":  "same-uid",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{},
				},
			},
		},
	})

	testClient := helpers.SetupTestClient()

	runs, err := testClient.GetKonflux().GetPipelineRuns(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(runs, 1)
	s.True(gock.IsDone())
}

func (s *PipelineRunsTestSuite) TestGetPipelineRun_ExactlyOne() {
	labels := "app=test"

	helpers.MockK8sPipelineRuns(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "single-run",
					"uid":  "uid-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Succeeded", "status": "True", "reason": "Completed"},
					},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	run, err := testClient.GetKonflux().GetPipelineRun(s.ctx, []string{labels})

	s.NoError(err)
	s.Equal("single-run", run.Metadata.Name)
	s.True(gock.IsDone())
}

func (s *PipelineRunsTestSuite) TestGetPipelineRun_MoreThanOne_Fails() {
	labels := "app=test"

	helpers.MockK8sPipelineRuns(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "run-1",
					"uid":  "uid-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{},
				},
			},
			{
				"metadata": map[string]interface{}{
					"name": "run-2",
					"uid":  "uid-2",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	_, err := testClient.GetKonflux().GetPipelineRun(s.ctx, []string{labels})

	s.Error(err)
	s.Contains(err.Error(), "more than 1 pipeline runs")
	s.True(gock.IsDone())
}

func (s *PipelineRunsTestSuite) TestGetPipelineRuns_NoResults() {
	labels := "app=nonexistent"

	helpers.MockQueryResourcesEmpty(helpers.TestK8sAPIServerURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant)
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "tekton.dev", "v1", "pipelineruns", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	_, err := testClient.GetKonflux().GetPipelineRuns(s.ctx, []string{labels})

	s.Error(err)
	s.Contains(err.Error(), "no pipeline runs found")
	s.True(gock.IsDone())
}
