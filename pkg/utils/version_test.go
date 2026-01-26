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
