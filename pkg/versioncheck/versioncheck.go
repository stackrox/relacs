// Package versioncheck provides utilities for validating that the current
// relacs version meets the minimum required version specified in the repository.
package versioncheck

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/stackrox/relacs/pkg/client"
)

const devVersion = "dev"

// CheckMinimumVersion checks if the current version meets the minimum required version.
// It returns an error if the current version is below the minimum or if any error occurs.
func CheckMinimumVersion(ctx context.Context, currentVersion string) error {
	// Special case: "dev" version skips the check with a warning
	if currentVersion == devVersion {
		fmt.Fprintln(os.Stderr, "Warning: Running development version, skipping minimum version check")
		return nil
	}

	current, err := parseVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to parse current version %q: %w", currentVersion, err)
	}

	// Create throwaway client for fetching the minimum required version.
	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	minVersion, err := fetchMinimumRequiredVersion(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to fetch minimum required version: %w", err)
	}

	minimum, err := parseVersion(minVersion)
	if err != nil {
		return fmt.Errorf("failed to parse minimum required version %q: %w", minVersion, err)
	}

	// Compare versions using semver
	if current.LessThan(minimum) {
		return fmt.Errorf(
			"current version %s is below the minimum required version %s\n\n"+
				"Please update relacs to at least version %s.\n"+
				"Installation instructions: https://github.com/stackrox/relacs#installation\n"+
				"Or run:\n"+
				"    bash <(gh api -H \"Accept: application/vnd.github.v3.raw\" \"/repos/stackrox/relacs/contents/scripts/install-relacs.sh\") [version]",
			currentVersion, minVersion, minVersion,
		)
	}

	return nil
}

// parseVersion parses a version string into a semver.Version
// Handles versions with or without 'v' prefix and build metadata (e.g., v0.9.9+g1234567)
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
