package github

import (
	"testing"

	"github.com/stackrox/relacs/pkg/github"
)

func TestAggregateCVEs(t *testing.T) {
	entries := []github.ImageScanEntry{
		{
			ImageName: "main", Source: "Upstream", Conclusion: "failure",
			Vulnerabilities: []github.ScanVulnerability{
				{CVEId: "CVE-1", CVESeverity: "CRITICAL", CVECVSS: 9.8, ComponentFixedVersion: "2.0"},
				{CVEId: "CVE-2", CVESeverity: "IMPORTANT", CVECVSS: 7.5, ComponentFixedVersion: ""},
				{CVEId: "CVE-3", CVESeverity: "LOW", CVECVSS: 2.0, ComponentFixedVersion: "1.1"},
			},
		},
		{
			ImageName: "scanner", Source: "Upstream", Conclusion: "success",
			Vulnerabilities: []github.ScanVulnerability{
				{CVEId: "CVE-1", CVESeverity: "CRITICAL", CVECVSS: 9.8, ComponentFixedVersion: "2.0"},
				{CVEId: "CVE-4", CVESeverity: "MODERATE", CVECVSS: 5.0, ComponentFixedVersion: ""},
			},
		},
	}

	allSeverities := map[string]bool{"CRITICAL": true, "IMPORTANT": true, "MODERATE": true, "LOW": true}
	defaultSeverities := map[string]bool{"CRITICAL": true, "IMPORTANT": true}

	tests := []struct {
		name       string
		severities map[string]bool
		fixable    bool
		cveFilter  map[string]bool
		wantIDs    []string
		check      func(t *testing.T, rows []cveRow)
	}{
		{
			name:       "default severity filter",
			severities: defaultSeverities,
			wantIDs:    []string{"CVE-1", "CVE-2"},
			check: func(t *testing.T, rows []cveRow) {
				if len(rows[0].Images) != 2 {
					t.Errorf("CVE-1 should affect 2 images, got %d", len(rows[0].Images))
				}
				if rows[1].Fixable {
					t.Error("CVE-2 should not be fixable")
				}
			},
		},
		{
			name:       "fixable only",
			severities: defaultSeverities,
			fixable:    true,
			wantIDs:    []string{"CVE-1"},
		},
		{
			name:       "all severities",
			severities: allSeverities,
			wantIDs:    []string{"CVE-1", "CVE-2", "CVE-4", "CVE-3"},
		},
		{
			name:       "cve id filter",
			severities: allSeverities,
			cveFilter:  map[string]bool{"CVE-3": true},
			wantIDs:    []string{"CVE-3"},
		},
		{
			name:       "failed scan marker",
			severities: map[string]bool{"CRITICAL": true},
			check: func(t *testing.T, rows []cveRow) {
				foundFailed := false
				for _, img := range rows[0].Images {
					if img == "Upstream: main (failed)" {
						foundFailed = true
					}
				}
				if !foundFailed {
					t.Errorf("expected failed marker on Upstream: main, got images: %v", rows[0].Images)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := aggregateCVEs(entries, test.severities, test.fixable, test.cveFilter)
			if test.wantIDs != nil {
				if len(rows) != len(test.wantIDs) {
					t.Fatalf("expected %d CVEs, got %d", len(test.wantIDs), len(rows))
				}
				for i, wantID := range test.wantIDs {
					if rows[i].ID != wantID {
						t.Errorf("row %d: expected %s, got %s", i, wantID, rows[i].ID)
					}
				}
			}
			if test.check != nil {
				test.check(t, rows)
			}
		})
	}
}
