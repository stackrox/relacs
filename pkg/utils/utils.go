package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/stackrox/relacs/pkg/config"
)

type Version struct {
	Major int
	Minor int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func GetApplicationWithVersionName(application, tag string) string {
	// For nightly builds, return just the application name
	if strings.Contains(tag, "nightly") {
		return application
	}

	// Extract major.minor version from tag
	parts := strings.Split(tag, ".")
	if len(parts) >= 2 {
		major := parts[0]
		minor := parts[1]
		return fmt.Sprintf("%s-%s-%s", application, major, minor)
	}

	// Fallback to just application name if tag format is unexpected
	return application
}

func GetVersionFromApplicationName(application string) (Version, error) {
	parts := strings.Split(application, "-")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid application name: %s. Expected format: <application>-<major>-<minor>", application)
	}
	return convertToVersion(parts[1], parts[2])
}

func convertToVersion(major, minor string) (Version, error) {
	majorInt, err := strconv.Atoi(major)
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	minorInt, err := strconv.Atoi(minor)
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	return Version{Major: majorInt, Minor: minorInt}, nil
}

func GetResourceWithVersion(resource, application string) (string, error) {
	// This is the case for nightly and untagged builds.
	if application == config.KonfluxApplicationPrefix {
		return resource, nil
	}
	version, err := GetVersionFromApplicationName(application)
	if err != nil {
		return "", fmt.Errorf("failed to get version from application name %s: %w", application, err)
	}
	return fmt.Sprintf("%s-%d-%d", resource, version.Major, version.Minor), nil
}

func GetPipelineRunURL(application, pipelineRunName string) string {
	// Example: https://konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com/ns/rh-acs-tenant/applications/acs-4-9/pipelineruns/collector-on-push-skx9p
	return fmt.Sprintf("https://%s/ns/%s/applications/%s/pipelineruns/%s", config.KonfluxHost, config.KonfluxTenant, application, pipelineRunName)
}

// gitDescribePattern matches the git-describe suffix: -<commits>-g<sha>
// Example: 4.10.x-423-g6e85db6baa matches because it ends with -423-g6e85db6baa
var gitDescribePattern = regexp.MustCompile(`-\d+-g[a-f0-9]+$`)
var nightlyTagPattern = regexp.MustCompile(`^\d+\.\d+\.x-nightly-\d{8}$`)
var nightlyTagPatternWithGroups = regexp.MustCompile(`^(\d+)\.(\d+)\.x-nightly-(\d{8})$`)

func IsACSVersionTagged(acsVersion string) bool {
	if nightlyTagPattern.MatchString(acsVersion) {
		return true
	}
	if gitDescribePattern.MatchString(acsVersion) {
		return false
	}
	return true
}

func GetCommitSHAFromGitDescribeVersion(version string) (string, error) {
	// Example: 4.10.x-423-g6e85db6baa --> 6e85db6baa
	parts := strings.Split(version, "-")
	if len(parts) >= 3 && strings.HasPrefix(parts[len(parts)-1], "g") {
		return parts[len(parts)-1][1:], nil // skip the "g"
	}
	return "", fmt.Errorf("invalid git-describe version: %s. Expected format: <major>.<minor>.x-<commits>-g<sha>", version)
}

func GetProductVersionFromTag(tag string) string {
	matches := nightlyTagPatternWithGroups.FindStringSubmatch(tag)
	if len(matches) == 4 {
		return fmt.Sprintf("%s.%s.0-nightly-%s-%s", matches[1], matches[2], matches[3], config.ImageTagSuffix)
	}
	return tag
}
