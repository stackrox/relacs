package konflux

import (
	"context"
	"fmt"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
)

func (c *Client) GetPodLogs(ctx context.Context, podName, containerName string) (string, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return "", err
	}
	logs, err := ka.GetPodLogs(ctx, podName, kubearchive.LogOptions{Container: containerName})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve pod logs: %w", err)
	}
	return logs, nil
}
