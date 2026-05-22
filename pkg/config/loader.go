package config

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// fetchConfigMap retrieves a ConfigMap from the Konflux cluster using kubectl
func fetchConfigMap(name, namespace string) (map[string]string, error) {
	cmd := exec.Command("kubectl", "get", "configmap", name, "-n", namespace, "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if kubectl is available
		if _, pathErr := exec.LookPath("kubectl"); pathErr != nil {
			return nil, fmt.Errorf("kubectl is not available in PATH: %w", pathErr)
		}

		return nil, fmt.Errorf("failed to fetch ConfigMap %s/%s: %w\nOutput: %s", namespace, name, err, string(output))
	}

	var configMap struct {
		Data map[string]string `json:"data"`
	}

	if err := json.Unmarshal(output, &configMap); err != nil {
		return nil, fmt.Errorf("failed to parse ConfigMap JSON: %w", err)
	}

	if configMap.Data == nil {
		return nil, fmt.Errorf("ConfigMap %s/%s has no data field", namespace, name)
	}

	return configMap.Data, nil
}
