package konflux

import (
	"encoding/json"
	"fmt"
	"strings"

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
	query := strings.Join(labels, ",")
	cmd, err := c.k8sCommand("get", "snapshot", "-l", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve matching snapshots: %w", err)
	}

	snapshotList := SnapshotList{}
	err = json.Unmarshal(output, &snapshotList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot list: %w", err)
	}
	return snapshotList.Items, nil
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


