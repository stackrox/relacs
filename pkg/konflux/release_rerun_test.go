package konflux

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestReleaseRerunTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseRerunTestSuite))
}

type ReleaseRerunTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *ReleaseRerunTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ReleaseRerunTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *ReleaseRerunTestSuite) TestIsReleaseOrRerunSucceeded_OriginalSucceeded() {
	name := "acs-4-11--4-11-7--cdn-stage-s390x-1"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, name, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": name,
			"uid":  "uid-1",
			"labels": map[string]interface{}{
				"acs-release-version": "4.11.7",
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
	s.Require().NoError(err)

	succeeded, successfulName, err := isReleaseOrRerunSucceeded(s.ctx, testClient, release)

	s.NoError(err)
	s.True(succeeded)
	s.Equal(name, successfulName)
	s.True(gock.IsDone())
}

func (s *ReleaseRerunTestSuite) TestIsReleaseOrRerunSucceeded_RerunSucceeded() {
	originalName := "acs-4-11--4-11-7--cdn-stage-s390x-1"
	rerunName := "acs-4-11--4-11-7--cdn-stage-s390x-1-rerun-wtz8k"
	version := "4.11.7"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, originalName, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": originalName,
			"uid":  "uid-1",
			"labels": map[string]interface{}{
				"acs-release-version": version,
			},
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-stage-rp",
			"snapshot":    "snap-1",
		},
		"status": map[string]interface{}{
			"conditions": []map[string]interface{}{
				{"type": "Released", "status": "False", "reason": "Failed"},
			},
		},
	})

	labels := "acs-release-version=" + version
	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": originalName,
					"uid":  "uid-1",
					"labels": map[string]interface{}{
						"acs-release-version": version,
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-stage-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "False", "reason": "Failed"},
					},
				},
			},
			{
				"metadata": map[string]interface{}{
					"name": rerunName,
					"uid":  "uid-2",
					"labels": map[string]interface{}{
						"acs-release-version": version,
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
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, originalName)
	s.Require().NoError(err)

	succeeded, successfulName, err := isReleaseOrRerunSucceeded(s.ctx, testClient, release)

	s.NoError(err)
	s.True(succeeded)
	s.Equal(rerunName, successfulName)
	s.True(gock.IsDone())
}

func (s *ReleaseRerunTestSuite) TestIsReleaseOrRerunSucceeded_NoSuccessfulRelease() {
	originalName := "acs-4-11--4-11-7--cdn-stage-s390x-1"
	rerunName := "acs-4-11--4-11-7--cdn-stage-s390x-1-rerun-wtz8k"
	version := "4.11.7"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, originalName, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": originalName,
			"uid":  "uid-1",
			"labels": map[string]interface{}{
				"acs-release-version": version,
			},
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-stage-rp",
			"snapshot":    "snap-1",
		},
		"status": map[string]interface{}{
			"conditions": []map[string]interface{}{
				{"type": "Released", "status": "False", "reason": "Failed"},
			},
		},
	})

	labels := "acs-release-version=" + version
	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": originalName,
					"uid":  "uid-1",
					"labels": map[string]interface{}{
						"acs-release-version": version,
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-stage-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "False", "reason": "Failed"},
					},
				},
			},
			{
				"metadata": map[string]interface{}{
					"name": rerunName,
					"uid":  "uid-2",
					"labels": map[string]interface{}{
						"acs-release-version": version,
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-stage-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "False", "reason": "Failed"},
					},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, originalName)
	s.Require().NoError(err)

	succeeded, successfulName, err := isReleaseOrRerunSucceeded(s.ctx, testClient, release)

	s.NoError(err)
	s.False(succeeded)
	s.Empty(successfulName)
	s.True(gock.IsDone())
}

func (s *ReleaseRerunTestSuite) TestIsReleaseOrRerunSucceeded_CDNReleaseWithArch() {
	originalName := "acs-4-11--4-11-7--cdn-stage-s390x-1"
	rerunName := "acs-4-11--4-11-7--cdn-stage-s390x-1-rerun-h6l25"
	version := "4.11.7"
	arch := "s390x"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, originalName, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": originalName,
			"uid":  "uid-1",
			"labels": map[string]interface{}{
				"acs-release-version": version,
				"acs-cdn-arch":        arch,
			},
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-cdn-stage-rp",
			"snapshot":    "snap-1",
		},
		"status": map[string]interface{}{
			"conditions": []map[string]interface{}{
				{"type": "Released", "status": "False", "reason": "Failed"},
			},
		},
	})

	labels := "acs-release-version=" + version + ",acs-cdn-arch=" + arch
	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": originalName,
					"uid":  "uid-1",
					"labels": map[string]interface{}{
						"acs-release-version": version,
						"acs-cdn-arch":        arch,
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-cdn-stage-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "False", "reason": "Failed"},
					},
				},
			},
			{
				"metadata": map[string]interface{}{
					"name": rerunName,
					"uid":  "uid-2",
					"labels": map[string]interface{}{
						"acs-release-version": version,
						"acs-cdn-arch":        arch,
					},
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-cdn-stage-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "True", "reason": "Succeeded"},
					},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, originalName)
	s.Require().NoError(err)

	succeeded, successfulName, err := isReleaseOrRerunSucceeded(s.ctx, testClient, release)

	s.NoError(err)
	s.True(succeeded)
	s.Equal(rerunName, successfulName)
	s.True(gock.IsDone())
}

func (s *ReleaseRerunTestSuite) TestIsReleaseOrRerunSucceeded_MultipleReruns_PicksMostRecent() {
	originalName := "acs-5-0--5-0-0--cdn-prod-x86-64-1"
	firstRerunName := "acs-5-0--5-0-0--cdn-prod-x86-64-1-rerun-abc12"
	secondRerunName := "acs-5-0--5-0-0--cdn-prod-x86-64-1-rerun-xyz99"
	version := "5.0.0"
	arch := "x86-64"

	helpers.MockQueryResource(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant, originalName, map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": originalName,
			"uid":  "uid-1",
			"labels": map[string]interface{}{
				"acs-release-version": version,
				"acs-cdn-arch":        arch,
			},
			"creationTimestamp": "2026-11-10T10:00:00Z",
		},
		"spec": map[string]interface{}{
			"releasePlan": "acs-cdn-prod-rp",
			"snapshot":    "snap-1",
		},
		"status": map[string]interface{}{
			"conditions": []map[string]interface{}{
				{"type": "Released", "status": "False", "reason": "Failed"},
			},
		},
	})

	labels := "acs-release-version=" + version + ",acs-cdn-arch=" + arch
	helpers.MockK8sReleases(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": originalName,
					"uid":  "uid-1",
					"labels": map[string]interface{}{
						"acs-release-version": version,
						"acs-cdn-arch":        arch,
					},
					"creationTimestamp": "2026-11-10T10:00:00Z",
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-cdn-prod-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "False", "reason": "Failed"},
					},
				},
			},
			{
				"metadata": map[string]interface{}{
					"name": firstRerunName,
					"uid":  "uid-2",
					"labels": map[string]interface{}{
						"acs-release-version": version,
						"acs-cdn-arch":        arch,
					},
					"creationTimestamp": "2026-11-10T11:00:00Z", // Earlier rerun
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-cdn-prod-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "True", "reason": "Succeeded"},
					},
				},
			},
			{
				"metadata": map[string]interface{}{
					"name": secondRerunName,
					"uid":  "uid-3",
					"labels": map[string]interface{}{
						"acs-release-version": version,
						"acs-cdn-arch":        arch,
					},
					"creationTimestamp": "2026-11-10T12:00:00Z", // Later rerun - should be picked
				},
				"spec": map[string]interface{}{
					"releasePlan": "acs-cdn-prod-rp",
					"snapshot":    "snap-1",
				},
				"status": map[string]interface{}{
					"conditions": []map[string]interface{}{
						{"type": "Released", "status": "True", "reason": "Succeeded"},
					},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "releases", config.KonfluxTenant)

	testClient, err := helpers.SetupTestClient()
	s.Require().NoError(err)

	release, err := testClient.GetKonflux().GetRelease(s.ctx, originalName)
	s.Require().NoError(err)

	succeeded, successfulName, err := isReleaseOrRerunSucceeded(s.ctx, testClient, release)

	s.NoError(err)
	s.True(succeeded)
	s.Equal(secondRerunName, successfulName, "Should pick the most recent successful rerun")
	s.True(gock.IsDone())
}
