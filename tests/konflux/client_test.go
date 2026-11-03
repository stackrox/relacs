package konflux

import (
	"context"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type ClientTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *ClientTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ClientTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *ClientTestSuite) TestNewClient_HealthCheckSuccess() {
	helpers.MockKubearchiveHealthCheckOK(config.KubearchiveAPIURL)

	httpClient := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(httpClient)

	client, err := kubearchive.NewClientWithHTTPClient(httpClient, config.KubearchiveAPIURL, "test-token", config.KonfluxTenant)

	s.NoError(err)
	s.NotNil(client)
	s.True(gock.IsDone())
}

func (s *ClientTestSuite) TestNewClient_HealthCheckFailure() {
	helpers.MockKubearchiveHealthCheckFailure(config.KubearchiveAPIURL)

	httpClient := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(httpClient)

	client, err := kubearchive.NewClientWithHTTPClient(httpClient, config.KubearchiveAPIURL, "test-token", config.KonfluxTenant)

	s.Error(err)
	s.Nil(client)
	s.Contains(err.Error(), "KubeArchive API unreachable")
	s.Contains(err.Error(), config.KubearchiveAPIURL)
	s.True(gock.IsDone())
}

func (s *ClientTestSuite) TestAuthError_401OnQuery() {
	labels := "app=test"

	// K8s API returns 401
	helpers.MockQueryResourcesAuthFailure(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	// KubeArchive also returns 401
	helpers.MockQueryResourcesAuthFailure(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	_, err = testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.Error(err)
	s.Contains(err.Error(), "Unauthorized")
	s.Contains(err.Error(), "oc login")
	s.True(gock.IsDone())
}

func (s *ClientTestSuite) TestAuthError_403OnQuery() {
	labels := "app=test"

	// K8s API returns 403
	helpers.MockQueryResourcesForbidden(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	// KubeArchive also returns 403
	helpers.MockQueryResourcesForbidden(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	_, err = testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.Error(err)
	s.Contains(err.Error(), "Forbidden")
	s.True(gock.IsDone())
}

func (s *ClientTestSuite) TestAuthError_K8sFails_KubeArchiveSucceeds() {
	labels := "app=test"

	// K8s API returns 401
	helpers.MockQueryResourcesAuthFailure(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	// KubeArchive succeeds
	helpers.MockKubearchiveReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "archived-release",
					"uid":  "uid-1",
				},
				"spec": map[string]interface{}{
					"releasePlan": "plan",
					"snapshot":    "snap",
				},
			},
		},
	})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	releases, err := testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(releases, 1)
	s.Equal("archived-release", releases[0].Metadata.Name)
	s.True(gock.IsDone())
}

func (s *ClientTestSuite) TestAuthError_K8sSucceeds_KubeArchiveFails() {
	labels := "app=test"

	// K8s API succeeds
	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "live-release",
					"uid":  "uid-1",
				},
				"spec": map[string]interface{}{
					"releasePlan": "plan",
					"snapshot":    "snap",
				},
			},
		},
	})

	// KubeArchive returns 401
	helpers.MockQueryResourcesAuthFailure(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	releases, err := testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(releases, 1)
	s.Equal("live-release", releases[0].Metadata.Name)
	s.True(gock.IsDone())
}
