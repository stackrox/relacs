package konflux

import (
	"fmt"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
)

func (c *Client) GetPodLogs(podName, containerName string) (string, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return "", err
	}
	logs, err := ka.GetPodLogs(podName, kubearchive.LogOptions{Container: containerName})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve pod logs: %w", err)
	}
	return logs, nil
}
