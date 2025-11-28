package utils

import "testing"

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
		result := GetApplicationVersionName(test.baseApplication, test.tag)
		if result != test.expected {
			t.Errorf("GetApplicationVersionName(%s, %s) = %s, expected %s", test.baseApplication, test.tag, result, test.expected)
		}
	}
}
