package utils

import (
	"fmt"
	"testing"
)

func TestVersionNewVersion(t *testing.T) {
	tests := []struct {
		version     string
		expected    Version
		expectedErr error
	}{
		{"4.10.1", Version{Major: 4, Minor: 10, Patch: 1}, nil},
		{"4.10.1-rc.1", Version{Major: 4, Minor: 10, Patch: 1}, nil},
		{"4.10.x-nightly-20251002", Version{Major: 0, Minor: 0, Patch: 0}, fmt.Errorf("invalid version: 4.10.x-nightly-20251002. Expected format: <major>.<minor>.<patch>-(rc.N)")},
	}

	for _, test := range tests {
		result, err := NewVersion(test.version)
		if result != test.expected {
			t.Errorf("NewVersion(%s) = %v, expected %v", test.version, result, test.expected)
		}
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("NewVersion(%s) = %v, unexpected error %v", test.version, result, err)
			} else if err.Error() != test.expectedErr.Error() {
				t.Errorf("NewVersion(%s) = %v, expected error %v", test.version, result, test.expectedErr)
			}
		} else if test.expectedErr != nil {
			t.Errorf("NewVersion(%s) = %v, expected error %v", test.version, result, test.expectedErr)
		}
	}
}

func TestVersionLessThan(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{
			name:     "different major versions",
			v1:       "3.0.0",
			v2:       "4.0.0",
			expected: true,
		},
		{
			name:     "different minor versions",
			v1:       "4.7.0",
			v2:       "4.8.0",
			expected: true,
		},
		{
			name:     "different patch versions",
			v1:       "4.8.0",
			v2:       "4.8.1",
			expected: true,
		},
		{
			name:     "equal versions",
			v1:       "4.8.1",
			v2:       "4.8.1",
			expected: false,
		},
		{
			name:     "first version greater",
			v1:       "4.8.2",
			v2:       "4.8.1",
			expected: false,
		},
		{
			name:     "RC versions treated as equal to base version",
			v1:       "4.8.1-rc.1",
			v2:       "4.8.1-rc.2",
			expected: false,
		},
		{
			name:     "RC version equal to base version",
			v1:       "4.8.1-rc.1",
			v2:       "4.8.1",
			expected: false,
		},
		{
			name:     "RC version compared with different patch",
			v1:       "4.8.1-rc.1",
			v2:       "4.8.2",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v1, err := NewVersion(test.v1)
			if err != nil {
				t.Fatalf("failed to parse version %s: %v", test.v1, err)
			}
			v2, err := NewVersion(test.v2)
			if err != nil {
				t.Fatalf("failed to parse version %s: %v", test.v2, err)
			}

			result := v1.LessThan(v2)
			if result != test.expected {
				t.Errorf("%s.LessThan(%s) = %v, expected %v", test.v1, test.v2, result, test.expected)
			}
		})
	}
}
