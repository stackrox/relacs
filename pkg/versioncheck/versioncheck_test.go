package versioncheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "simple version with v prefix",
			version:     "v0.0.9",
			expectError: false,
		},
		{
			name:        "simple version without v prefix",
			version:     "0.0.9",
			expectError: false,
		},
		{
			name:        "version with pre-release info (git describe)",
			version:     "v0.0.9-12-g9a829b8",
			expectError: false,
		},
		{
			name:        "version with rc tag",
			version:     "v1.2.3-rc.1",
			expectError: false,
		},
		{
			name:        "invalid version",
			version:     "invalid",
			expectError: true,
		},
		{
			name:        "empty version",
			version:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := parseVersion(tt.version)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, v)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, v)
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name       string
		current    string
		minimum    string
		shouldFail bool
	}{
		{
			name:       "current version equals minimum",
			current:    "v0.0.9",
			minimum:    "v0.0.9",
			shouldFail: false,
		},
		{
			name:       "current version greater than minimum",
			current:    "v0.0.10",
			minimum:    "v0.0.9",
			shouldFail: false,
		},
		{
			name:       "current version less than minimum",
			current:    "v0.0.8",
			minimum:    "v0.0.9",
			shouldFail: true,
		},
		{
			name:       "major version difference",
			current:    "v1.0.0",
			minimum:    "v0.0.9",
			shouldFail: false,
		},
		{
			name:       "minor version difference",
			current:    "v0.1.0",
			minimum:    "v0.0.9",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := parseVersion(tt.current)
			require.NoError(t, err)

			minimum, err := parseVersion(tt.minimum)
			require.NoError(t, err)

			isLess := isLessThan(current, tt.current, minimum, tt.minimum)
			if tt.shouldFail {
				assert.True(t, isLess, "expected %s < %s", tt.current, tt.minimum)
			} else {
				assert.False(t, isLess, "expected %s >= %s", tt.current, tt.minimum)
			}
		})
	}
}

func TestCheckMinimumVersion_DevVersion(t *testing.T) {
	err := CheckMinimumVersion("dev")
	assert.NoError(t, err)
}

func TestCheckMinimumVersion_InvalidCurrentVersion(t *testing.T) {
	err := CheckMinimumVersion("invalid-version")
	assert.Error(t, err)
}

func TestVersionComparisonEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		version1   string
		version2   string
		expectLess bool
	}{
		{
			name:       "v0.0.9 vs v0.0.10",
			version1:   "v0.0.9",
			version2:   "v0.0.10",
			expectLess: true,
		},
		{
			name:       "v0.0.9-12-g9a829b8 vs v0.0.9",
			version1:   "v0.0.9-12-g9a829b8",
			version2:   "v0.0.9",
			expectLess: false, // git-describe version is greater than the base tag
		},
		{
			name:       "v0.0.9-12-g9a829b8 vs v0.0.10",
			version1:   "v0.0.9-12-g9a829b8",
			version2:   "v0.0.10",
			expectLess: true,
		},
		{
			name:       "v1.0.0-rc.1 vs v1.0.0",
			version1:   "v1.0.0-rc.1",
			version2:   "v1.0.0",
			expectLess: true, // rc is less than final release (standard semver)
		},
		{
			name:       "v0.0.9 vs v0.0.9-12-g9a829b8",
			version1:   "v0.0.9",
			version2:   "v0.0.9-12-g9a829b8",
			expectLess: true, // tag is less than commits after tag
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := parseVersion(tt.version1)
			require.NoError(t, err)

			v2, err := parseVersion(tt.version2)
			require.NoError(t, err)

			result := isLessThan(v1, tt.version1, v2, tt.version2)
			assert.Equal(t, tt.expectLess, result,
				"expected isLessThan = %v for %s vs %s",
				tt.expectLess, tt.version1, tt.version2)
		})
	}
}
