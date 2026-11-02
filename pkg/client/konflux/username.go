package konflux

import (
	"context"
	"fmt"
)

func (c *Client) GetCurrentUsername(ctx context.Context) (string, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return "", err
	}
	username, err := ka.GetCurrentUsername(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current username: %w", err)
	}
	return username, nil
}
