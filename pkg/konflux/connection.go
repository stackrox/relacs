package konflux

import (
	"fmt"
	"strings"
)

func GetCurrentContext() (string, error) {
	// use the kubectl config current-context command to get the current context
	// return the context name
	cmd, err := getK8sCommand("config", "current-context")
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func GetCurrentUsername() (string, error) {
	currentContext, err := GetCurrentContext()
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}
	username := strings.Split(currentContext, "/")[2]
	return username, nil
}
