package konflux

import (
	"context"
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
	"github.com/stackrox/relacs/pkg/config"
)

func (c *Client) GetReleases(ctx context.Context, labels []string) ([]Release, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return nil, err
	}
	opts := kubearchive.QueryOptions{
		LabelSelector: strings.Join(labels, ","),
	}
	releases, err := kubearchive.QueryResources[Release](ctx, ka, "appstudio.redhat.com", "v1alpha1", "releases", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve releases: %w", err)
	}
	return releases, nil
}

func (c *Client) GetRelease(ctx context.Context, name string) (Release, error) {
	ka, err := c.getKubearchive()
	if err != nil {
		return Release{}, err
	}
	release, err := kubearchive.QueryResource[Release](ctx, ka, "appstudio.redhat.com", "v1alpha1", "releases", name)
	if err != nil {
		return Release{}, fmt.Errorf("failed to retrieve release: %w", err)
	}
	return release, nil
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
