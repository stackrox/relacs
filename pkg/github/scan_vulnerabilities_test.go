package github

import "testing"

func TestParseScanResultFromLogs(t *testing.T) {
	tests := []struct {
		name          string
		logText       string
		wantCritical  int
		wantImportant int
		wantTotal     int
		wantErr       bool
	}{
		{
			name: "valid scan result in logs",
			logText: `2026-05-20T16:14:03.2243032Z   "scan-image-vulnerabilities.sh" \
2026-05-20T16:14:07.5985632Z {
2026-05-20T16:14:07.5986000Z   "result": {
2026-05-20T16:14:07.5986767Z     "summary": {
2026-05-20T16:14:07.5987064Z       "CRITICAL": 2,
2026-05-20T16:14:07.5987899Z       "IMPORTANT": 9,
2026-05-20T16:14:07.5988206Z       "LOW": 196,
2026-05-20T16:14:07.5988490Z       "MODERATE": 94,
2026-05-20T16:14:07.5988808Z       "TOTAL-COMPONENTS": 86,
2026-05-20T16:14:07.5989213Z       "TOTAL-VULNERABILITIES": 299
2026-05-20T16:14:07.5989585Z     },
2026-05-20T16:14:07.5989852Z     "vulnerabilities": []
2026-05-20T16:14:07.5990000Z   }
2026-05-20T16:14:07.5990100Z }
2026-05-20T16:14:07.6758349Z ##[error]Found fixable critical or important vulnerabilities.`,
			wantCritical:  2,
			wantImportant: 9,
			wantTotal:     299,
			wantErr:       false,
		},
		{
			name: "result on single line",
			logText: `2026-05-20T16:14:07.5985632Z {"result": {"summary": {"CRITICAL": 0, "IMPORTANT": 3, "LOW": 10, "MODERATE": 5, "TOTAL-COMPONENTS": 20, "TOTAL-VULNERABILITIES": 18}, "vulnerabilities": []}}
2026-05-20T16:14:07.6758349Z ##[error]Done.`,
			wantCritical:  0,
			wantImportant: 3,
			wantTotal:     18,
			wantErr:       false,
		},
		{
			name:    "no scan result in logs",
			logText: "2026-05-20T16:14:03.2243032Z Some random log line\n",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseScanResultFromLogs(test.logText)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Summary.Critical != test.wantCritical {
				t.Errorf("Critical: got %d, want %d", result.Summary.Critical, test.wantCritical)
			}
			if result.Summary.Important != test.wantImportant {
				t.Errorf("Important: got %d, want %d", result.Summary.Important, test.wantImportant)
			}
			if result.Summary.TotalVulnerabilities != test.wantTotal {
				t.Errorf("Total: got %d, want %d", result.Summary.TotalVulnerabilities, test.wantTotal)
			}
		})
	}
}
