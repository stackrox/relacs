package konflux

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestLogsTestSuite(t *testing.T) {
	suite.Run(t, new(LogsTestSuite))
}

type LogsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *LogsTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *LogsTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *LogsTestSuite) TestGetPodLogs_FromK8sAPI() {
	podName := "my-pod-abc123"
	container := "step-build"
	logOutput := "Building image...\nBuild completed successfully."

	helpers.MockGetPodLogs(helpers.TestK8sAPIServerURL, config.KonfluxTenant, podName, container, logOutput)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(s.ctx, podName, container)

	s.NoError(err)
	s.Equal("Building image...\nBuild completed successfully.", logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_FallsBackToKubeArchive() {
	podName := "archived-pod"
	container := "step-build"
	logOutput := "Archived log output"

	// K8s API returns 404
	helpers.MockGetPodLogsNotFound(helpers.TestK8sAPIServerURL, config.KonfluxTenant, podName)

	// KubeArchive has the logs
	helpers.MockGetPodLogs(config.KubearchiveAPIURL, config.KonfluxTenant, podName, container, logOutput)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(s.ctx, podName, container)

	s.NoError(err)
	s.Equal("Archived log output", logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_TrimsWhitespace() {
	podName := "my-pod"
	container := "step-test"
	logOutput := "  \n  test output  \n  "

	helpers.MockGetPodLogs(helpers.TestK8sAPIServerURL, config.KonfluxTenant, podName, container, logOutput)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(s.ctx, podName, container)

	s.NoError(err)
	s.Equal("test output", logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_EmptyLogs() {
	podName := "my-pod"
	container := "step-build"

	helpers.MockGetPodLogs(helpers.TestK8sAPIServerURL, config.KonfluxTenant, podName, container, "")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(s.ctx, podName, container)

	s.NoError(err)
	s.Empty(logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_NotFoundAnywhere() {
	podName := "nonexistent-pod"

	helpers.MockGetPodLogsNotFound(helpers.TestK8sAPIServerURL, config.KonfluxTenant, podName)
	helpers.MockGetPodLogsNotFound(config.KubearchiveAPIURL, config.KonfluxTenant, podName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	_, err = testClient.GetKonflux().GetPodLogs(s.ctx, podName, "container")

	s.Error(err)
	s.Contains(err.Error(), "not found")
	s.True(gock.IsDone())
}
