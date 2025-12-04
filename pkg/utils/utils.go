package utils

import (
	"fmt"
	"strconv"
	"strings"
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

func GetVersionFromTag(tag string) (Version, error) {
	if strings.Contains(tag, "nightly") {
		return Version{Major: 0, Minor: 0}, fmt.Errorf("nightly tags are not supported by this method")
	}

	parts := strings.Split(tag, ".")
	if len(parts) >= 2 {
		return convertToVersion(parts[0], parts[1])
	}
	return Version{}, fmt.Errorf("invalid tag format: %s", tag)
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

func GetComponentNameWithVersion(component, tag string) (string, error) {
	version, err := GetVersionFromTag(tag)
	if err != nil {
		return "", fmt.Errorf("failed to get version from tag %s: %w", tag, err)
	}
	return fmt.Sprintf("%s-%d-%d", component, version.Major, version.Minor), nil
}
