package jira

import (
	"testing"
)

func TestConvertJiraComponentName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		expectedError bool
	}{
		{
			name:     "standard format with rhel8 at end",
			input:    "advanced-cluster-security/rhacs-main-rhel8",
			expected: "main",
		},
		{
			name:     "rhel8 in the middle",
			input:    "advanced-cluster-security/rhacs-rhel8-operator",
			expected: "operator",
		},
		{
			name:     "scanner-db component",
			input:    "advanced-cluster-security/rhacs-scanner-db-rhel8",
			expected: "scanner-db",
		},
		{
			name:     "central-db component",
			input:    "advanced-cluster-security/rhacs-central-db-rhel8",
			expected: "central-db",
		},
		{
			name:     "rhel9 variant",
			input:    "advanced-cluster-security/rhacs-main-rhel9",
			expected: "main",
		},
		{
			name:     "rhel10 multi-digit variant",
			input:    "advanced-cluster-security/rhacs-scanner-rhel10",
			expected: "scanner",
		},
		{
			name:     "rhel11 multi-digit variant",
			input:    "advanced-cluster-security/rhacs-rhel11-operator",
			expected: "operator",
		},
		{
			name:     "rhacs-eng release component",
			input:    "rhacs-eng/release-scanner-v4",
			expected: "scanner-v4",
		},
		{
			name:     "rhacs-eng release prefix case-insensitive",
			input:    "RHACS-ENG/Release-main-rhel8",
			expected: "main",
		},
		{
			name:          "invalid format - no slash",
			input:         "rhacs-main-rhel8",
			expectedError: true,
		},
		{
			name:          "invalid format - empty component",
			input:         "advanced-cluster-security/rhacs-rhel8",
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := convertJiraComponentName(test.input)

			if test.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}
