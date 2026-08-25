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
