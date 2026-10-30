package konflux

import "fmt"

func (c *Client) GetCurrentUsername() (string, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return "", err
	}
	username, err := ka.GetCurrentUsername()
	if err != nil {
		return "", fmt.Errorf("failed to get current username: %w", err)
	}
	return username, nil
}
