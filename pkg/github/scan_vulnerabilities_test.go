package github

import (
	"encoding/json"
	"testing"
)

func TestScanResultWrapperParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantCritical  int
		wantImportant int
		wantTotal     int
		wantVulnCount int
		wantErr       bool
	}{
		{
			name: "full scan result",
			input: `{
				"result": {
					"summary": {
						"CRITICAL": 2,
						"IMPORTANT": 9,
						"LOW": 196,
						"MODERATE": 94,
						"TOTAL-COMPONENTS": 86,
						"TOTAL-VULNERABILITIES": 299
					},
					"vulnerabilities": [
						{
							"cveId": "CVE-2026-1234",
							"cveSeverity": "CRITICAL",
							"cveCVSS": 9.8,
							"cveInfo": "https://example.com",
							"componentName": "glibc",
							"componentVersion": "2.28",
							"componentFixedVersion": "2.29"
						}
					]
				}
			}`,
			wantCritical:  2,
			wantImportant: 9,
			wantTotal:     299,
			wantVulnCount: 1,
		},
		{
			name: "empty vulnerabilities",
			input: `{
				"result": {
					"summary": {
						"CRITICAL": 0,
						"IMPORTANT": 0,
						"LOW": 0,
						"MODERATE": 0,
						"TOTAL-COMPONENTS": 10,
						"TOTAL-VULNERABILITIES": 0
					},
					"vulnerabilities": []
				}
			}`,
			wantCritical:  0,
			wantImportant: 0,
			wantTotal:     0,
			wantVulnCount: 0,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var wrapper scanResultWrapper
			err := json.Unmarshal([]byte(test.input), &wrapper)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			result := wrapper.Result
			if result.Summary.Critical != test.wantCritical {
				t.Errorf("Critical: got %d, want %d", result.Summary.Critical, test.wantCritical)
			}
			if result.Summary.Important != test.wantImportant {
				t.Errorf("Important: got %d, want %d", result.Summary.Important, test.wantImportant)
			}
			if result.Summary.TotalVulnerabilities != test.wantTotal {
				t.Errorf("Total: got %d, want %d", result.Summary.TotalVulnerabilities, test.wantTotal)
			}
			if len(result.Vulnerabilities) != test.wantVulnCount {
				t.Errorf("Vulnerabilities count: got %d, want %d", len(result.Vulnerabilities), test.wantVulnCount)
			}
		})
	}
}
