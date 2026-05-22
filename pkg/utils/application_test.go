package utils

import (
	"fmt"
	"testing"

	"github.com/stackrox/relacs/pkg/config"
)

func TestApplicationWithVersion(t *testing.T) {
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
		result := Application{Name: test.baseApplication}.WithVersion(test.tag)
		if result != test.expected {
			t.Errorf("Application{Name: %s}.WithVersion(%s) = %s, expected %s", test.baseApplication, test.tag, result, test.expected)
		}
	}
}

func TestApplicationGetVersion(t *testing.T) {
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
		{"acs-a-9", Version{Major: 0, Minor: 0}, fmt.Errorf("failed to convert version: invalid major version: strconv.Atoi: parsing \"a\": invalid syntax")},
		{"acs-4-a", Version{Major: 0, Minor: 0}, fmt.Errorf("failed to convert version: invalid minor version: strconv.Atoi: parsing \"a\": invalid syntax")},
	}
	for _, test := range tests {
		result, err := Application{Name: test.application}.GetVersion()
		if result != test.expected {
			t.Errorf("Application{Name: %s}.GetVersion() = %v, expected %v", test.application, result, test.expected)
		}
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("Application{Name: %s}.GetVersion() = %v, unexpected error %v", test.application, result, err)
			} else if err.Error() != test.expectedErr.Error() {
				fmt.Println(err.Error())
				t.Errorf("Application{Name: %s}.GetVersion() = %v, expected error %v", test.application, result, test.expectedErr)
			}
		} else if test.expectedErr != nil {
			t.Errorf("Application{Name: %s}.GetVersion() = %v, expected error %v", test.application, result, test.expectedErr)
		}
	}
}

func TestGetResourceWithVersion(t *testing.T) {
	cfg := &config.Config{
		KonfluxApplicationPrefix: "acs",
	}
	tests := []struct {
		resource    string
		application string
		expected    string
		expectedErr error
	}{
		{"operator-bundle", "acs-4-9", "operator-bundle-4-9", nil},
		{"operator-bundle", "acs", "operator-bundle", nil},
		{"operator-bundle", "something-4-9-0", "", fmt.Errorf("failed to get version from application name something-4-9-0: invalid application name: something-4-9-0. Expected format: <application>-<major>-<minor>")},
	}
	for _, test := range tests {
		result, err := Application{Name: test.application}.GetResourceWithVersion(cfg, test.resource)
		if result != test.expected {
			t.Errorf("Application{Name: %s}.GetResourceWithVersion(%s) = %s, expected %s", test.application, test.resource, result, test.expected)
		}
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("Application{Name: %s}.GetResourceWithVersion(%s) = %v, unexpected error %v", test.application, test.resource, result, err)
			} else if err.Error() != test.expectedErr.Error() {
				t.Errorf("Application{Name: %s}.GetResourceWithVersion(%s) = %v, expected error %v", test.application, test.resource, result, test.expectedErr)
			}
		} else if test.expectedErr != nil {
			t.Errorf("Application{Name: %s}.GetResourceWithVersion(%s) = %v, expected error %v", test.application, test.resource, result, test.expectedErr)
		}
	}
}
