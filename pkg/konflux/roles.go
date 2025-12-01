package konflux

import (
	"encoding/json"
	"fmt"
)

type RoleBindingList struct {
	Items []RoleBinding `json:"items"`
}

type RoleBinding struct {
	Metadata RoleBindingMetadata `json:"metadata"`
	RoleRef  RoleRef             `json:"roleRef"`
	Subjects []Subject           `json:"subjects"`
}

type RoleBindingMetadata struct {
	Name string `json:"name"`
}

type RoleRef struct {
	Name string `json:"name"`
}

type Subject struct {
	Name string `json:"name"`
}

func HasKonfluxRole(clusterRole, username, namespace string) (bool, error) {
	cmd, err := getK8sCommand("get", "rolebinding", "-n", namespace, "-o", "json")
	if err != nil {
		return false, fmt.Errorf("failed to get rolebindings: %w", err)
	}
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get rolebindings: %w", err)
	}

	// Parse the output JSON list of rolebindings and keep the name, roleRef.name and subject.name
	var rolebindings RoleBindingList
	err = json.Unmarshal(output, &rolebindings)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal rolebindings: %w", err)
	}

	// Check if there is a rolebinding with the username in the subject and the clusterRole in the roleRef
	for _, rolebinding := range rolebindings.Items {
		if rolebinding.Subjects[0].Name == username && rolebinding.RoleRef.Name == clusterRole {
			return true, nil
		}
	}
	return false, fmt.Errorf("no rolebinding found with the username %s and the clusterRole %s", username, clusterRole)
}

func HasKonfluxAdminRole(username, namespace string) (bool, error) {
	return HasKonfluxRole("konflux-admin-user-actions", username, namespace)
}

func HasKonfluxMaintainerRole(username, namespace string) (bool, error) {
	return HasKonfluxRole("konflux-maintainer-user-actions", username, namespace)
}
