package konflux

import (
	"testing"
)

func TestValidateTagFormat(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		// Valid formats
		{
			name:        "valid format 1.2.3",
			tag:         "1.2.3",
			expectError: false,
		},
		{
			name:        "valid format 10.20.30",
			tag:         "10.20.30",
			expectError: false,
		},
		{
			name:        "valid format with larger numbers",
			tag:         "100.200.300",
			expectError: false,
		},
		// Invalid formats
		{
			name:        "invalid format 1.2 (missing patch)",
			tag:         "1.2",
			expectError: true,
		},
		{
			name:        "invalid format 1.2.3.4 (too many components)",
			tag:         "1.2.3.4",
			expectError: true,
		},
		{
			name:        "invalid format v1.2.3 (with v prefix)",
			tag:         "v1.2.3",
			expectError: true,
		},
		{
			name:        "invalid format 1.2.3-rc1 (with suffix)",
			tag:         "1.2.3-rc1",
			expectError: true,
		},
		{
			name:        "invalid format empty string",
			tag:         "",
			expectError: true,
		},
		{
			name:        "invalid format 1.2.3-beta",
			tag:         "1.2.3-beta",
			expectError: true,
		},
		{
			name:        "invalid format 1.2.3.rc1",
			tag:         "1.2.3.rc1",
			expectError: true,
		},
		{
			name:        "invalid format 1 (single number)",
			tag:         "1",
			expectError: true,
		},
		{
			name:        "invalid format 1.x.3 (non-numeric)",
			tag:         "1.x.3",
			expectError: true,
		},
		{
			name:        "invalid format with leading zero",
			tag:         "01.02.03",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTagFormat(tt.tag)
			if tt.expectError && err == nil {
				t.Errorf("validateTagFormat(%q) expected error, got nil", tt.tag)
			}
			if !tt.expectError && err != nil {
				t.Errorf("validateTagFormat(%q) unexpected error: %v", tt.tag, err)
			}
		})
	}
}
