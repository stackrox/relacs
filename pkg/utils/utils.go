package utils

import (
	"fmt"
	"strings"
)

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
