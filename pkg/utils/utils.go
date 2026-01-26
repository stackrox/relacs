package utils

import (
	"fmt"
	"regexp"

	"github.com/stackrox/relacs/pkg/config"
)

// gitDescribePattern matches the git-describe suffix: -<commits>-g<sha>
// Example: 4.10.x-423-g6e85db6baa matches because it ends with -423-g6e85db6baa
var gitDescribePattern = regexp.MustCompile(`^(.+)-\d+-g([a-f0-9]+)$`)
var nightlyTagPattern = regexp.MustCompile(`^(\d+)\.(\d+)\.x-nightly-(\d{8})$`)

func IsACSVersionTagged(version string) bool {
	return !gitDescribePattern.MatchString(version)
}

func IsNightlyVersion(version string) bool {
	return nightlyTagPattern.MatchString(version)
}

func GetCommitSHAFromGitDescribeVersion(version string) (string, error) {
	// Example: 4.10.x-423-g6e85db6baa --> 6e85db6baa
	matches := gitDescribePattern.FindStringSubmatch(version)
	if len(matches) == 3 {
		return matches[2], nil
	}
	return "", fmt.Errorf("invalid git-describe version: %s. Expected format: <tag>-<commits>-g<sha>", version)
}

func GetProductVersionFromTag(version string) string {
	// Nightly: product version replaces the patch version with 0.
	matches := nightlyTagPattern.FindStringSubmatch(version)
	if len(matches) == 4 {
		return fmt.Sprintf("%s.%s.0-nightly-%s-%s", matches[1], matches[2], matches[3], config.ImageTagSuffix)
	}
	return version
}
