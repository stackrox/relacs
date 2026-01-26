package utils

import (
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/config"
)

type Application struct {
	Name string
}

func (a Application) WithVersion(tag string) string {
	// For nightly builds, return just the application name
	if strings.Contains(tag, "nightly") {
		return a.Name
	}

	// Extract major.minor version from tag
	parts := strings.Split(tag, ".")
	if len(parts) >= 2 {
		major := parts[0]
		minor := parts[1]
		return fmt.Sprintf("%s-%s-%s", a.Name, major, minor)
	}

	// Fallback to just application name if tag format is unexpected
	return a.Name
}

func (a Application) GetVersion() (Version, error) {
	parts := strings.Split(a.Name, "-")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid application name: %s. Expected format: <application>-<major>-<minor>", a.Name)
	}
	versionObj, err := convertToVersion(parts[1], parts[2], "0")
	if err != nil {
		return Version{}, fmt.Errorf("failed to convert version: %w", err)
	}
	return versionObj, nil
}

func (a Application) GetResourceWithVersion(resource string) (string, error) {
	// This is the case for nightly and untagged builds.
	if a.Name == config.KonfluxApplicationPrefix {
		return resource, nil
	}
	version, err := a.GetVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get version from application name %s: %w", a.Name, err)
	}
	return fmt.Sprintf("%s-%d-%d", resource, version.Major, version.Minor), nil
}

func (a Application) GetPipelineRunURL(pipelineRunName string) string {
	// Example: https://konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com/ns/rh-acs-tenant/applications/acs-4-9/pipelineruns/collector-on-push-skx9p
	return fmt.Sprintf("https://%s/ns/%s/applications/%s/pipelineruns/%s", config.KonfluxHost, config.KonfluxTenant, a.Name, pipelineRunName)
}
