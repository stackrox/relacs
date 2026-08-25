package konflux

import (
	"encoding/json"
	"errors"
	"fmt"
)

// userInfo is the data returned from the kubectl auth whoami command.
type userInfo struct {
	Status struct {
		UserInfo struct {
			Username string `json:"username"`
		} `json:"userInfo"`
	} `json:"status"`
}

func (c *Client) GetCurrentUsername() (string, error) {
	cmd, err := getDirectK8sCommandWithJSONOutput("auth", "whoami")
	if err != nil {
		return "", fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run kubectl command: %w", err)
	}

	info := userInfo{}
	err = json.Unmarshal(output, &info)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	username := info.Status.UserInfo.Username
	if username == "" {
		return "", errors.New("empty username returned from kubectl")
	}
	return username, nil
}
