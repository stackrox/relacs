package konflux

import (
	"fmt"
	"strings"
)

func (c *Client) GetPodLogs(podName, containerName string) (string, error) {
	cmd, err := c.k8sCommandPlainOutput("logs", podName, "-c", containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve pod logs: %w, %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}
