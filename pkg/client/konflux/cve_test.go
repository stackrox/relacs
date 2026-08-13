package konflux

import (
	"testing"
)

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      Severity
		expectedError bool
	}{
		{name: "Low", input: "Low", expected: Severity_Low},
		{name: "Moderate", input: "Moderate", expected: Severity_Moderate},
		{name: "Important", input: "Important", expected: Severity_Important},
		{name: "Critical", input: "Critical", expected: Severity_Critical},
		{name: "Invalid - High", input: "High", expectedError: true},
		{name: "Invalid - empty", input: "", expectedError: true},
		{name: "Invalid - lowercase", input: "low", expectedError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseSeverity(test.input)
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
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestExtractKeyAndYearFromCVE(t *testing.T) {
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
			key, year, err := extractKeyAndYearFromCVE(test.input)
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
