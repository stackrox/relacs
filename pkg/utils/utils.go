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

func GetApplicationVersionName(application, tag string) string {
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
	major, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version in application name '%s': %w", application, err)
	}
	minor, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version in application name '%s': %w", application, err)
	}
	return Version{
		Major: major,
		Minor: minor,
	}, nil
}
