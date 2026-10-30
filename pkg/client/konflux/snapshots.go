package konflux

import (
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/utils"
)

func (s Snapshot) GetURL() (string, error) {
	application, ok := s.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return "", fmt.Errorf("application not found in snapshot labels")
	}
	return fmt.Sprintf("https://%s/ns/%s/applications/%s/snapshots/%s", config.KonfluxHost, config.KonfluxTenant, application, s.Metadata.Name), nil
}

func (c *Client) GetSnapshots(labels []string) ([]Snapshot, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return nil, err
	}
	opts := kubearchive.QueryOptions{
		LabelSelector: strings.Join(labels, ","),
	}
	snapshots, err := kubearchive.QueryResources[Snapshot](ka, "appstudio.redhat.com", "v1alpha1", "snapshots", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve matching snapshots: %w", err)
	}
	return snapshots, nil
}

func (s Snapshot) GetOperatorBundleImage(tag string) (string, error) {
	application, ok := s.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return "", fmt.Errorf("application not found in snapshot labels")
	}
	operatorBundleComponentName, err := utils.Application{Name: application}.GetResourceWithVersion("operator-bundle")
	if err != nil {
		return "", fmt.Errorf("failed to get operator bundle component name: %w", err)
	}
	for _, component := range s.Spec.Components {
		if component.Name == operatorBundleComponentName {
			return component.ContainerImage, nil
		}
	}
	return "", fmt.Errorf("operator bundle image not found in snapshot")
}
