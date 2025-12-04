package konflux

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetRelease(name string) (Release, error) {
	cmd, err := c.k8sCommand("get", "release", name)
	if err != nil {
		return Release{}, fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return Release{}, fmt.Errorf("failed to retrieve release: %w, %s", err, string(output))
	}

	releaseList := ReleaseList{}
	err = json.Unmarshal(output, &releaseList)
	if err != nil {
		return Release{}, fmt.Errorf("failed to unmarshal release list: %w", err)
	}
	if len(releaseList.Items) == 0 {
		return Release{}, fmt.Errorf("release %s not found", name)
	}
	return releaseList.Items[0], nil
}

func (r Release) GetAdvisoryInternalURL() string {
	return r.Status.Artifacts.Advisory.InternalURL
}

func (r Release) GetAdvisoryPublicURL() string {
	return r.Status.Artifacts.Advisory.URL
}

func (r Release) IsCompleted() bool {
	return r.Status.CompletionTime != nil
}
