package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/client/github"
)

const (
	devVersion = "dev"
)

// CheckMinimumVersion checks if the current version meets the minimum required version.
// It returns an error if the current version is below the minimum or if any error occurs.
func CheckMinimumVersion(ctx context.Context, c *client.Client, currentVersion string) error {
	// Special case: "dev" version skips the check with a warning
	if currentVersion == devVersion {
		fmt.Println("Warning: Running development version, skipping minimum version check")
		return nil
	}

	minVersion, err := github.FetchMinimumRequiredVersion(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to fetch minimum required version: %w", err)
	}

	// Parse versions
	current, err := parseVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current version %q: %w", currentVersion, err)
	}

	minimum, err := parseVersion(minVersion)
	if err != nil {
		return fmt.Errorf("failed to parse minimum required version %q: %w", minVersion, err)
	}

	// Compare versions using custom logic for git-describe format
	if isLessThan(current, currentVersion, minimum, minVersion) {
		return fmt.Errorf(
			"current version %s is below the minimum required version %s\n\n"+
				"Please update relacs to at least version %s.\n"+
				"Installation instructions: https://github.com/stackrox/relacs#installation\n"+
				"Or run: bash <(gh api -H \"Accept: application/vnd.github.v3.raw\" \"/repos/stackrox/relacs/contents/scripts/install-relacs.sh\") [version]",
			currentVersion, minVersion, minVersion,
		)
	}

	return nil
}

// isLessThan compares two versions with special handling for git-describe format.
// Git describe format: v0.0.9-12-g9a829b8 means "12 commits after v0.0.9"
// This should be considered greater than v0.0.9 but less than v0.0.10
func isLessThan(v1 *semver.Version, v1Str string, v2 *semver.Version, v2Str string) bool {
	// Pattern for git describe format: version-commitCount-gCommitHash
	gitDescribePattern := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)-(\d+)-g[a-f0-9]+$`)

	v1IsGitDescribe := gitDescribePattern.MatchString(v1Str)
	v2IsGitDescribe := gitDescribePattern.MatchString(v2Str)

	// If both are git-describe format or neither is, use standard semver comparison
	if v1IsGitDescribe == v2IsGitDescribe {
		return v1.LessThan(v2)
	}

	// If v1 is git-describe format (e.g., v0.0.9-12-g9a829b8)
	if v1IsGitDescribe {
		matches := gitDescribePattern.FindStringSubmatch(v1Str)
		baseVersion := matches[1]

		// Parse the base version without the commit info
		v1Base, err := parseVersion(baseVersion)
		if err != nil {
			// Fall back to standard comparison if parsing fails
			return v1.LessThan(v2)
		}

		// If base versions are equal, git-describe version is greater than the tag
		// e.g., v0.0.9-12-g9a829b8 > v0.0.9
		if v1Base.Equal(v2) {
			return false
		}

		// Otherwise compare base versions
		return v1Base.LessThan(v2)
	}

	// If v2 is git-describe format
	matches := gitDescribePattern.FindStringSubmatch(v2Str)
	baseVersion := matches[1]

	v2Base, err := parseVersion(baseVersion)
	if err != nil {
		// Fall back to standard comparison if parsing fails
		return v1.LessThan(v2)
	}

	// If base versions are equal, normal version is less than git-describe version
	// e.g., v0.0.9 < v0.0.9-12-g9a829b8
	if v1.Equal(v2Base) {
		return true
	}

	// Otherwise compare base versions
	return v1.LessThan(v2Base)
}

// parseVersion parses a version string into a semver.Version
// Handles versions with or without 'v' prefix and git describe format (e.g., v0.0.9-12-g9a829b8)
func parseVersion(version string) (*semver.Version, error) {
	// Ensure version has 'v' prefix for semver parsing
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid semantic version: %w", err)
	}

	return v, nil
}
