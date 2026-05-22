package utils

import (
	"fmt"
	"testing"

	"github.com/stackrox/relacs/pkg/config"
)

func TestIsACSVersionTagged(t *testing.T) {
	tests := []struct {
		acsVersion string
		expected   bool
	}{
		{"4.10.x-423-g6e85db6baa", false},
		{"4.10.1", true},
		{"4.10.x-nightly-20251002", true},
	}
	for _, test := range tests {
		result := IsACSVersionTagged(test.acsVersion)
		if result != test.expected {
			t.Errorf("IsACSVersionTagged(%s) = %v, expected %v", test.acsVersion, result, test.expected)
		}
	}
}

func TestGetProductVersionFromTag(t *testing.T) {
	cfg := &config.Config{
		ImageTagSuffix: "fast",
	}
	tests := []struct {
		tag         string
		expected    string
		expectedErr error
	}{
		{"4.10.x-nightly-20260108", "4.10.0-nightly-20260108-fast", nil},
		{"4.10.1", "4.10.1", nil},
	}

	for _, test := range tests {
		result := GetProductVersionFromTag(cfg, test.tag)
		if result != test.expected {
			t.Errorf("GetProductVersionFromTag(%s) = %s, expected %s", test.tag, result, test.expected)
		}
	}
}

func TestGetCommitSHAFromGitDescribeVersion(t *testing.T) {
	tests := []struct {
		acsVersion  string
		expected    string
		expectedErr error
	}{
		{"4.10.x-423-g6e85db6baa", "6e85db6baa", nil},
		{"4.10.x-invalid-format", "", fmt.Errorf("invalid git-describe version: 4.10.x-invalid-format. Expected format: <tag>-<commits>-g<sha>")},
	}
	for _, test := range tests {
		result, err := GetCommitSHAFromGitDescribeVersion(test.acsVersion)
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("GetCommitSHAFromGitDescribeVersion(%s) = %v, unexpected error %v", test.acsVersion, result, err)
			} else if err.Error() != test.expectedErr.Error() {
				t.Errorf("GetCommitSHAFromGitDescribeVersion(%s) = %v, expected error %v", test.acsVersion, result, test.expectedErr)
			}
		} else if test.expectedErr != nil {
			t.Errorf("GetCommitSHAFromGitDescribeVersion(%s) = %v, expected error %v", test.acsVersion, result, test.expectedErr)
		}
	}
}
