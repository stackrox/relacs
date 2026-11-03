package konflux

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestReleasesTestSuite(t *testing.T) {
	suite.Run(t, new(ReleasesTestSuite))
}

type ReleasesTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *ReleasesTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ReleasesTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *ReleasesTestSuite) TestGetReleases_FromBothSources() {
	labels := "pac.test.appstudio.openshift.io/sha=abc123"

	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "release-live",
					"uid":  "uid-1",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-stage-rp",
					"snapshot":    "snap-1",
				},
			},
		},
	})

	helpers.MockKubearchiveReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "release-archived",
					"uid":  "uid-2",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-prod-rp",
					"snapshot":    "snap-2",
				},
			},
		},
	})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	releases, err := testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(releases, 2)
	s.Equal("release-live", releases[0].Metadata.Name)
	s.Equal("release-archived", releases[1].Metadata.Name)
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetReleases_DeduplicatesByUID() {
	labels := "app=test"

	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "release-1",
					"uid":  "same-uid",
				},
				"spec": map[string]interface{}{
					"releasePlan": "plan-1",
					"snapshot":    "snap-1",
				},
			},
		},
	})

	helpers.MockKubearchiveReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "release-1",
					"uid":  "same-uid",
				},
				"spec": map[string]interface{}{
					"releasePlan": "plan-1",
					"snapshot":    "snap-1",
				},
			},
		},
	})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	releases, err := testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(releases, 1)
	s.Equal("release-1", releases[0].Metadata.Name)
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetReleases_OnlyKubeArchive() {
	labels := "app=test"

	// K8s API returns empty
	helpers.MockQueryResourcesEmpty(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	helpers.MockKubearchiveReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "archived-release",
					"uid":  "uid-archived",
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

func (s *ReleasesTestSuite) TestGetReleases_OnlyK8sAPI() {
	labels := "app=test"

	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "live-release",
					"uid":  "uid-live",
				},
				"spec": map[string]interface{}{
					"releasePlan": "plan",
					"snapshot":    "snap",
				},
			},
		},
	})

	// KubeArchive returns empty
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	releases, err := testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(releases, 1)
	s.Equal("live-release", releases[0].Metadata.Name)
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetReleases_EmptyResults() {
	labels := "app=nonexistent"

	helpers.MockQueryResourcesEmpty(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	releases, err := testClient.GetKonflux().GetReleases(s.ctx, []string{labels})

	s.NoError(err)
	s.Empty(releases)
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetRelease_FromK8sAPI() {
	name := "my-release"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": name,
			"uid":  "uid-1",
			"labels": map[string]interface{}{
				"appstudio.openshift.io/application": "acs-4-11",
			},
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-stage-rp",
			"snapshot":    "snap-1",
		},
		"status": map[string]interface{}{
			"conditions": []map[string]interface{}{
				{"type": "Released", "status": "True", "reason": "Succeeded"},
			},
		},
	})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, name)

	s.NoError(err)
	s.Equal(name, release.Metadata.Name)
	s.True(release.IsSucceeded())
	s.Equal("succeeded", release.GetState())
	s.Equal("stage", release.GetTarget())
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetRelease_FallsBackToKubeArchive() {
	name := "archived-release"

	// K8s API returns 404
	helpers.MockQueryResourceNotFound(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name)

	// KubeArchive has it
	helpers.MockQueryResource(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": name,
			"uid":  "uid-archived",
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-prod-rp",
			"snapshot":    "snap-1",
		},
	})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, name)

	s.NoError(err)
	s.Equal(name, release.Metadata.Name)
	s.Equal("prod", release.GetTarget())
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetRelease_NotFoundAnywhere() {
	name := "nonexistent"

	helpers.MockQueryResourceNotFound(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name)
	helpers.MockQueryResourceNotFound(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	_, err = testClient.GetKonflux().GetRelease(s.ctx, name)

	s.Error(err)
	s.Contains(err.Error(), "not found")
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetRelease_HelperMethods() {
	name := "test-release"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": name,
			"labels": map[string]interface{}{
				"appstudio.openshift.io/application": "acs-4-11",
			},
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-prod-rp",
			"snapshot":    "snap-1",
		},
		"status": map[string]interface{}{
			"completionTime": "2026-10-29T12:00:00Z",
			"conditions": []map[string]interface{}{
				{"type": "Released", "status": "True", "reason": "Succeeded"},
			},
			"artifacts": map[string]interface{}{
				"advisory": map[string]interface{}{
					"internal_url": "https://internal.example.com/advisory",
					"url":          "https://public.example.com/advisory",
				},
			},
		},
	})

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, name)

	s.NoError(err)
	s.True(release.IsCompleted())
	s.True(release.IsSucceeded())
	s.Equal("succeeded", release.GetState())
	s.Equal("prod", release.GetTarget())
	s.Equal("https://internal.example.com/advisory", release.GetAdvisoryInternalURL())
	s.Equal("https://public.example.com/advisory", release.GetAdvisoryPublicURL())

	url, err := release.GetURL()
	s.NoError(err)
	s.Contains(url, "acs-4-11")
	s.Contains(url, name)
	s.True(gock.IsDone())
}

func (s *ReleasesTestSuite) TestGetReleases_PaginatedResults() {
	labels := "app=test"

	// K8s API returns empty (single page, no continue)
	helpers.MockQueryResourcesEmpty(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	// KubeArchive page 1: has continue token
	helpers.MockKubearchiveReleases(labels, map[string]interface{}{
		"metadata": map[string]interface{}{
			"continue": "token-page2",
		},
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "release-page1",
					"uid":  "uid-1",
				},
				"spec": map[string]interface{}{
					"releasePlan": "plan",
					"snapshot":    "snap",
				},
			},
		},
	})

	// KubeArchive page 2: no continue token
	helpers.MockQueryResourcesWithContinue(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, labels, "token-page2", map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "release-page2",
					"uid":  "uid-2",
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
	s.Len(releases, 2)
	s.Equal("release-page1", releases[0].Metadata.Name)
	s.Equal("release-page2", releases[1].Metadata.Name)
	s.True(gock.IsDone())
}
