package konflux

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/config"
)

func (c *Client) GetReleases(labels []string) ([]Release, error) {
	query := strings.Join(labels, ",")
	cmd, err := c.k8sCommand("get", "releases.appstudio.redhat.com", "-l", query)
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve releases: %w", err)
	}

	releaseList := ReleaseList{}
	err = json.Unmarshal(output, &releaseList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal release list: %w", err)
	}
	return releaseList.Items, nil
}

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

func (r Release) GetState() string {
	if r.Status == nil {
		return "unknown"
	}
	for _, c := range r.Status.Conditions {
		if c.Type == "Released" {
			return strings.ToLower(c.Reason)
		}
	}
	return "unknown"
}

func (r Release) GetTarget() string {
	plan := r.Spec.ReleasePlan
	if strings.Contains(plan, "prod") {
		return "prod"
	}
	if strings.Contains(plan, "stage") {
		return "stage"
	}
	return "unknown"
}

func (r Release) GetURL() (string, error) {
	application, ok := r.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return "", fmt.Errorf("application not found in release labels")
	}
	return fmt.Sprintf("https://%s/ns/%s/applications/%s/releases/%s", config.KonfluxHost, config.KonfluxTenant, application, r.Metadata.Name), nil
}
