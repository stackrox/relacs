package jira

import "testing"

func TestFindLowestMatchingVersion(t *testing.T) {
	tests := []struct {
		name          string
		versions      []FixVersion
		major         int
		minor         int
		expected      string
		expectedError bool
	}{
		{
			name: "picks lowest patch among matching versions",
			versions: []FixVersion{
				{Name: "4.8.2"},
				{Name: "4.8.0"},
				{Name: "4.8.1"},
				{Name: "4.9.0"},
			},
			major:    4,
			minor:    8,
			expected: "4.8.0",
		},
		{
			name: "ignores non-semver and other major.minor",
			versions: []FixVersion{
				{Name: "next"},
				{Name: "4.9.0"},
				{Name: "4.8.1-rc.1"},
				{Name: "4.8.2"},
			},
			major:    4,
			minor:    8,
			expected: "4.8.1-rc.1",
		},
		{
			name: "skips released and archived versions",
			versions: []FixVersion{
				{Name: "4.8.0", Released: true},
				{Name: "4.8.1", Archived: true},
				{Name: "4.8.2"},
				{Name: "4.8.3"},
			},
			major:    4,
			minor:    8,
			expected: "4.8.2",
		},
		{
			name: "no matching version",
			versions: []FixVersion{
				{Name: "4.9.0"},
				{Name: "5.0.0"},
			},
			major:         4,
			minor:         8,
			expectedError: true,
		},
		{
			name: "only released matching versions",
			versions: []FixVersion{
				{Name: "4.8.0", Released: true},
				{Name: "4.8.1", Released: true},
			},
			major:         4,
			minor:         8,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := findLowestMatchingVersion(test.versions, test.major, test.minor)
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

func TestExtractCVEKeyAndYear(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedKey   string
		expectedYear  string
		expectedError bool
	}{
		{name: "uppercase", input: "CVE-2025-12345", expectedKey: "CVE-2025-12345", expectedYear: "2025"},
		{name: "lowercase", input: "cve-2024-678", expectedKey: "CVE-2024-678", expectedYear: "2024"},
		{name: "mixed case should fail", input: "CvE-2023-999", expectedError: true},
		{name: "invalid format", input: "INVALID-2025-12345", expectedError: true},
		{name: "missing year", input: "CVE-12345", expectedError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, year, err := ExtractCVEKeyAndYear(test.input)
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
			if key != test.expectedKey {
				t.Errorf("expected key %q, got %q", test.expectedKey, key)
			}
			if year != test.expectedYear {
				t.Errorf("expected year %q, got %q", test.expectedYear, year)
			}
		})
	}
}
