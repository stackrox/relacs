package konflux

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestSnapshotsTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotsTestSuite))
}

type SnapshotsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *SnapshotsTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *SnapshotsTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *SnapshotsTestSuite) TestGetSnapshots_FromBothSources() {
	labels := "acs-style-snapshot=true,appstudio.openshift.io/application=acs-4-11"

	helpers.MockK8sSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "snapshot-live",
					"uid":  "uid-live",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{},
				},
			},
		},
	})

	helpers.MockKubearchiveSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "snapshot-archived",
					"uid":  "uid-archived",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{},
				},
			},
		},
	})

	testClient := helpers.SetupTestClient()

	snapshots, err := testClient.GetKonflux().GetSnapshots(s.ctx, []string{
		"acs-style-snapshot=true",
		"appstudio.openshift.io/application=acs-4-11",
	})

	s.NoError(err)
	s.Len(snapshots, 2)
	s.Equal("snapshot-live", snapshots[0].Metadata.Name)
	s.Equal("snapshot-archived", snapshots[1].Metadata.Name)
	s.True(gock.IsDone())
}

func (s *SnapshotsTestSuite) TestGetSnapshots_DeduplicatesByUID() {
	labels := "app=test"

	helpers.MockK8sSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "snapshot-1",
					"uid":  "same-uid",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{},
				},
			},
		},
	})

	helpers.MockKubearchiveSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "snapshot-1",
					"uid":  "same-uid",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{},
				},
			},
		},
	})

	testClient := helpers.SetupTestClient()

	snapshots, err := testClient.GetKonflux().GetSnapshots(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(snapshots, 1)
	s.True(gock.IsDone())
}

func (s *SnapshotsTestSuite) TestGetSnapshots_OnlyFromLiveCluster() {
	labels := "acs-style-snapshot=true"

	helpers.MockK8sSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "live-only-snapshot",
					"uid":  "uid-live",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{
						{
							"name":           "main",
							"containerImage": "quay.io/rhacs-eng/main@sha256:abc123",
						},
					},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	snapshots, err := testClient.GetKonflux().GetSnapshots(s.ctx, []string{labels})

	s.NoError(err)
	s.Len(snapshots, 1)
	s.Equal("live-only-snapshot", snapshots[0].Metadata.Name)
	s.Len(snapshots[0].Spec.Components, 1)
	s.Equal("main", snapshots[0].Spec.Components[0].Name)
	s.True(gock.IsDone())
}

func (s *SnapshotsTestSuite) TestGetSnapshots_EmptyResults() {
	labels := "nonexistent=true"

	helpers.MockQueryResourcesEmpty(helpers.TestK8sAPIServerURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant)
	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	snapshots, err := testClient.GetKonflux().GetSnapshots(s.ctx, []string{labels})

	s.NoError(err)
	s.Empty(snapshots)
	s.True(gock.IsDone())
}

func (s *SnapshotsTestSuite) TestSnapshot_GetURL() {
	labels := "app=test"

	helpers.MockK8sSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "my-snapshot",
					"uid":  "uid-1",
					"labels": map[string]interface{}{
						"appstudio.openshift.io/application": "acs-4-11",
					},
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	snapshots, err := testClient.GetKonflux().GetSnapshots(s.ctx, []string{labels})

	s.Require().NoError(err)
	s.Require().Len(snapshots, 1)

	url, err := snapshots[0].GetURL()
	s.NoError(err)
	s.Contains(url, config.KonfluxHost)
	s.Contains(url, "acs-4-11")
	s.Contains(url, "my-snapshot")
	s.True(gock.IsDone())
}

func (s *SnapshotsTestSuite) TestSnapshot_GetURL_MissingApplicationLabel() {
	labels := "app=test"

	helpers.MockK8sSnapshots(labels, map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name": "no-app-label-snapshot",
					"uid":  "uid-1",
				},
				"spec": map[string]interface{}{
					"components": []map[string]interface{}{},
				},
			},
		},
	})

	helpers.MockQueryResourcesEmpty(config.KubearchiveAPIURL, "appstudio.redhat.com", "v1alpha1", "snapshots", config.KonfluxTenant)

	testClient := helpers.SetupTestClient()

	snapshots, err := testClient.GetKonflux().GetSnapshots(s.ctx, []string{labels})

	s.Require().NoError(err)
	s.Require().Len(snapshots, 1)

	_, err = snapshots[0].GetURL()
	s.Error(err)
	s.Contains(err.Error(), "application not found")
	s.True(gock.IsDone())
}
