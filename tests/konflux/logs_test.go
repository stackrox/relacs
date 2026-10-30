package konflux

import (
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
}

func (s *LogsTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *LogsTestSuite) TestGetPodLogs_Success() {
	podName := "my-pod-abc123"
	container := "step-build"
	logOutput := "Building image...\nBuild completed successfully."

	helpers.MockGetPodLogs(config.KubearchiveAPIURL, config.KonfluxTenant, podName, container, logOutput)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(podName, container)

	s.NoError(err)
	s.Equal("Building image...\nBuild completed successfully.", logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_TrimsWhitespace() {
	podName := "my-pod"
	container := "step-test"
	logOutput := "  \n  test output  \n  "

	helpers.MockGetPodLogs(config.KubearchiveAPIURL, config.KonfluxTenant, podName, container, logOutput)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(podName, container)

	s.NoError(err)
	s.Equal("test output", logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_EmptyLogs() {
	podName := "my-pod"
	container := "step-build"

	helpers.MockGetPodLogs(config.KubearchiveAPIURL, config.KonfluxTenant, podName, container, "")

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	logs, err := testClient.GetKonflux().GetPodLogs(podName, container)

	s.NoError(err)
	s.Empty(logs)
	s.True(gock.IsDone())
}

func (s *LogsTestSuite) TestGetPodLogs_NotFound() {
	podName := "nonexistent-pod"

	helpers.MockGetPodLogsNotFound(config.KubearchiveAPIURL, config.KonfluxTenant, podName)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	_, err = testClient.GetKonflux().GetPodLogs(podName, "container")

	s.Error(err)
	s.True(gock.IsDone())
}
