package konflux

import (
	"fmt"
	"strings"
)

func fetchPodLogs(namespace, podName, container string) (string, error) {
	cmd, err := getK8sCommand("ka", "logs", podName, "-c", container, "-n", namespace)
	if err != nil {
		return "", fmt.Errorf("failed to fetch pod logs: %w", err)
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch pod logs: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
