package konflux

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetCurrentUsername() (string, error) {
	cmd, err := getDirectK8sCommandWithJSONOutput("auth", "whoami")
	if err != nil {
		return "", fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run kubectl command: %w", err)
	}

	userInfo := UserInfo{}
	err = json.Unmarshal(output, &userInfo)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal user info: %w", err)
	}
	return userInfo.Status.UserInfo.Username, nil
}
