package konflux

import (
	"encoding/json"
	"fmt"

	"github.com/stackrox/relacs/pkg/config"
)

func (c *Client) GetRelease(name string) (Release, error) {
	cmd, err := c.k8sCommand("get", "release", name)
	if err != nil {
		return Release{}, fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return Release{}, fmt.Errorf("failed to retrieve release: %w", err)
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
	if r.Status == nil {
		return ""
	}
	return r.Status.Artifacts.Advisory.InternalURL
}

func (r Release) GetAdvisoryPublicURL() string {
	if r.Status == nil {
		return ""
	}
	return r.Status.Artifacts.Advisory.URL
}

func (r Release) IsCompleted() bool {
	if r.Status == nil {
		return false
	}
	return r.Status.CompletionTime != nil
}

func (r Release) IsSucceeded() bool {
	if r.Status == nil {
		return false
	}
	for _, c := range r.Status.Conditions {
		if c.Type == "Released" {
			return c.Status == "True"
		}
	}
	return false
}

func (r Release) GetURL(cfg *config.Config) (string, error) {
	application, ok := r.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return "", fmt.Errorf("application not found in release labels")
	}
	return fmt.Sprintf("https://%s/ns/%s/applications/%s/releases/%s", cfg.KonfluxHost, cfg.KonfluxTenant, application, r.Metadata.Name), nil
}
