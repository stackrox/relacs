package utils

import (
	"fmt"
	"testing"
)

func TestGetApplicationVersionName(t *testing.T) {
	// 4.8.6 --> $application-4-8
	// 4.8.6-rc.1 --> $application-4-8
	// 4.9.x-nightly-20251002 --> $application
	tests := []struct {
		tag             string
		baseApplication string
		expected        string
	}{
		{"4.8.6", "acs", "acs-4-8"},
		{"4.8.6-rc.1", "acs", "acs-4-8"},
		{"4.9.x-nightly-20251002", "acs", "acs"},
	}
	for _, test := range tests {
		result := GetApplicationWithVersionName(test.baseApplication, test.tag)
		if result != test.expected {
			t.Errorf("GetApplicationVersionName(%s, %s) = %s, expected %s", test.baseApplication, test.tag, result, test.expected)
		}
	}
}

func TestGetVersionFromApplicationName(t *testing.T) {
	// acs-4-9 --> Version{Major: 4, Minor: 9}
	// acs --> Version{Major: 0, Minor: 0}, invalid application name
	tests := []struct {
		application string
		expected    Version
		expectedErr error
	}{
		{"acs-4-9", Version{Major: 4, Minor: 9}, nil},
		{"acs", Version{Major: 0, Minor: 0}, fmt.Errorf("invalid application name: acs. Expected format: <application>-<major>-<minor>")},
		{"acs-4-9-nightly-20251002", Version{Major: 0, Minor: 0}, fmt.Errorf("invalid application name: acs-4-9-nightly-20251002. Expected format: <application>-<major>-<minor>")},
		{"acs-a-9", Version{Major: 0, Minor: 0}, fmt.Errorf("invalid major version in application name 'acs-a-9': strconv.Atoi: parsing \"a\": invalid syntax")},
		{"acs-4-a", Version{Major: 0, Minor: 0}, fmt.Errorf("invalid minor version in application name 'acs-4-a': strconv.Atoi: parsing \"a\": invalid syntax")},
	}
	for _, test := range tests {
		result, err := GetVersionFromApplicationName(test.application)
		if result != test.expected {
			t.Errorf("GetVersionFromApplicationName(%s) = %v, expected %v", test.application, result, test.expected)
		}
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("GetVersionFromApplicationName(%s) = %v, unexpected error %v", test.application, result, err)
			} else if err.Error() != test.expectedErr.Error() {
				t.Errorf("GetVersionFromApplicationName(%s) = %v, expected error %v", test.application, result, test.expectedErr)
			}
		} else if test.expectedErr != nil {
			t.Errorf("GetVersionFromApplicationName(%s) = %v, expected error %v", test.application, result, test.expectedErr)
		}
	}
}

func TestGetVersionFromTag(t *testing.T) {
	tests := []struct {
		tag         string
		expected    Version
		expectedErr error
	}{
		{"4.8.6", Version{Major: 4, Minor: 8}, nil},
		{"4.8.6-rc.1", Version{Major: 4, Minor: 8}, nil},
		{"4.9.x-nightly-20251002", Version{Major: 0, Minor: 0}, fmt.Errorf("nightly tags are not supported by this method")},
		{"4", Version{Major: 0, Minor: 0}, fmt.Errorf("invalid tag format: 4")},
	}
	for _, test := range tests {
		result, err := GetVersionFromTag(test.tag)
		if result != test.expected {
			t.Errorf("GetVersionFromTag(%s) = %v, expected %v", test.tag, result, test.expected)
		}
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("GetVersionFromTag(%s) = %v, unexpected error %v", test.tag, result, err)
			} else if err.Error() != test.expectedErr.Error() {
				t.Errorf("GetVersionFromTag(%s) = %v, expected error %v", test.tag, result, test.expectedErr)
			}
		} else if test.expectedErr != nil {
			t.Errorf("GetVersionFromTag(%s) = %v, expected error %v", test.tag, result, test.expectedErr)
		}
	}
}
