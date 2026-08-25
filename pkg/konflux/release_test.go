package konflux

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stackrox/relacs/pkg/client/jira"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/utils"
)

func TestDetermineHighestSeverity(t *testing.T) {
	tests := []struct {
		name     string
		cves     []kflx.CVEInformation
		expected kflx.Severity
	}{
		{
			name: "no CVEs",
			cves: []kflx.CVEInformation{},
		},
		{
			name: "one CVE",
			cves: []kflx.CVEInformation{
				{Severity: kflx.Severity_Low},
			},
			expected: kflx.Severity_Low,
		},
		{
			name: "multiple CVEs",
			cves: []kflx.CVEInformation{
				{Severity: kflx.Severity_Low},
				{Severity: kflx.Severity_Critical},
				{Severity: kflx.Severity_Important},
				{Severity: kflx.Severity_Moderate},
			},
			expected: kflx.Severity_Critical,
		},
		{
			name: "multiple CVEs with same severity",
			cves: []kflx.CVEInformation{
				{Severity: kflx.Severity_Moderate},
				{Severity: kflx.Severity_Moderate},
				{Severity: kflx.Severity_Moderate},
			},
			expected: kflx.Severity_Moderate,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := determineHighestSeverity(test.cves)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestDetermineReleaseType(t *testing.T) {
	tests := []struct {
		name     string
		cves     []kflx.CVEInformation
		version  utils.Version
		expected kflx.ReleaseType
	}{
		{
			name:     "no CVEs and patch release",
			cves:     []kflx.CVEInformation{},
			version:  utils.Version{Major: 4, Minor: 8, Patch: 7},
			expected: kflx.ReleaseType_RHBA,
		},
		{
			name:     "no CVEs and minor release",
			cves:     []kflx.CVEInformation{},
			version:  utils.Version{Major: 4, Minor: 8, Patch: 0},
			expected: kflx.ReleaseType_RHEA,
		},
		{
			name: "any CVEs",
			cves: []kflx.CVEInformation{
				{Severity: kflx.Severity_Low},
			},
			version:  utils.Version{Major: 4, Minor: 8, Patch: 0},
			expected: kflx.ReleaseType_RHSA,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := determineReleaseType(test.cves, test.version)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestRenderSynopsis(t *testing.T) {
	tests := []struct {
		name            string
		releaseType     kflx.ReleaseType
		version         utils.Version
		issues          []string
		highestSeverity kflx.Severity
		expected        string
	}{
		{
			name:            "RHSA and patch release with issues",
			releaseType:     kflx.ReleaseType_RHSA,
			version:         utils.Version{Major: 4, Minor: 8, Patch: 7},
			issues:          []string{"CVE-2024-1234"},
			highestSeverity: kflx.Severity_Important,
			expected:        "Important: RHACS 4.8.7 security and bug fix update",
		},
		{
			name:            "RHSA and patch release without issues",
			releaseType:     kflx.ReleaseType_RHSA,
			version:         utils.Version{Major: 4, Minor: 8, Patch: 7},
			highestSeverity: kflx.Severity_Important,
			expected:        "Important: RHACS 4.8.7 security update",
		},
		{
			name:            "RHSA and minor release",
			releaseType:     kflx.ReleaseType_RHSA,
			version:         utils.Version{Major: 4, Minor: 8, Patch: 0},
			highestSeverity: kflx.Severity_Important,
			expected:        "Important: RHACS 4.8.0 security and enhancement update",
		},
		{
			name:        "RHBA release",
			releaseType: kflx.ReleaseType_RHBA,
			version:     utils.Version{Major: 4, Minor: 8, Patch: 7},
			expected:    "RHACS 4.8.7 bug fix update",
		},
		{
			name:        "RHEA release",
			releaseType: kflx.ReleaseType_RHEA,
			version:     utils.Version{Major: 4, Minor: 8, Patch: 0},
			expected:    "RHACS 4.8.0 enhancement update",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := renderSynopsis(test.releaseType, test.version, test.issues, test.highestSeverity)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestRenderReferences(t *testing.T) {
	tests := []struct {
		name     string
		version  utils.Version
		expected []string
	}{
		{name: "patch release", version: utils.Version{Major: 4, Minor: 8, Patch: 7}, expected: []string{"https://docs.redhat.com/en/documentation/red_hat_advanced_cluster_security_for_kubernetes/4.8/html-single/release_notes/index#about-this-release-487_release-notes-48"}},
		{name: "minor release", version: utils.Version{Major: 4, Minor: 8, Patch: 0}, expected: []string{"https://docs.redhat.com/en/documentation/red_hat_advanced_cluster_security_for_kubernetes/4.8/html-single/release_notes/index"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := renderReferences(test.version)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestRenderReleaseName(t *testing.T) {
	application := utils.Application{Name: utils.Application{Name: "acs"}.WithVersion("4.8.7")}
	tests := []struct {
		name     string
		version  utils.Version
		target   string
		attempt  int
		expected string
	}{
		{name: "patch release", version: utils.Version{Major: 4, Minor: 8, Patch: 7}, target: "stage", attempt: 2, expected: "acs-4-8--4-8-7--stage-2"},
		{name: "minor release", version: utils.Version{Major: 4, Minor: 8, Patch: 0}, target: "prod", attempt: 1, expected: "acs-4-8--4-8-0--prod-1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := renderReleaseName(application, test.version, test.target, test.attempt)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestRenderIssues(t *testing.T) {
	tests := []struct {
		name     string
		issues   []string
		expected []map[string]string
	}{
		{name: "no issues", issues: []string{}, expected: []map[string]string{}},
		{name: "multiple issues", issues: []string{"ROX-31406", "ROX-31951"}, expected: []map[string]string{{"id": "ROX-31406", "source": "redhat.atlassian.net"}, {"id": "ROX-31951", "source": "redhat.atlassian.net"}}},
		{name: "duplicate issues are skipped", issues: []string{"ROX-31406", "ROX-31406"}, expected: []map[string]string{{"id": "ROX-31406", "source": "redhat.atlassian.net"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := renderIssues(test.issues)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestCVELess(t *testing.T) {
	cves := []kflx.CVEInformation{
		{Key: "CVE-2026-53488", Component: "scanner", Year: "2026"},
		{Key: "CVE-2025-100", Component: "scanner", Year: "2025"},
		{Key: "CVE-2026-9165", Component: "scanner", Year: "2026"},
		{Key: "CVE-2026-100", Component: "central", Year: "2026"},
		{Key: "CVE-2026-50", Component: "scanner", Year: "2026"},
	}
	sort.Slice(cves, cveLess(cves))

	want := []string{
		"central/CVE-2026-100",
		"scanner/CVE-2025-100",
		"scanner/CVE-2026-50",
		"scanner/CVE-2026-9165",
		"scanner/CVE-2026-53488",
	}
	got := make([]string, len(cves))
	for i, cve := range cves {
		got[i] = cve.Component + "/" + cve.Key
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sort order:\n got %v\nwant %v", got, want)
	}
}

func TestRenderCVEs(t *testing.T) {
	application := utils.Application{Name: utils.Application{Name: "acs"}.WithVersion("4.8.7")}
	snapshot := kflx.Snapshot{
		Spec: kflx.SnapshotSpec{
			Components: []kflx.SnapshotComponent{{Name: "scanner-db-4-8"}},
		},
	}
	tests := []struct {
		name          string
		cves          []jira.CVE
		expected      []kflx.CVEInformation
		expectedError string
	}{
		{
			name:     "no CVEs",
			cves:     []jira.CVE{},
			expected: []kflx.CVEInformation{},
		},
		{
			name:     "single CVE",
			cves:     []jira.CVE{{Code: "CVE-2025-59375", Key: "CVE-2025-59375", Year: "2025", ComponentName: "scanner-db", Severity: "Important"}},
			expected: []kflx.CVEInformation{{Key: "CVE-2025-59375", Component: "scanner-db-4-8", Severity: kflx.Severity_Important, Year: "2025"}},
		},
		{
			name:          "CVE with invalid component",
			cves:          []jira.CVE{{Code: "CVE-2025-59375", Key: "CVE-2025-59375", Year: "2025", ComponentName: "invalid-component", Severity: "Important"}},
			expected:      nil,
			expectedError: "failed to validate component invalid-component for CVE-2025-59375: component not found in snapshot: invalid-component-4-8",
		},
		{
			name:          "CVE with empty severity",
			cves:          []jira.CVE{{Code: "CVE-2025-59375", Key: "CVE-2025-59375", Year: "2025", ComponentName: "scanner-db", Severity: ""}},
			expected:      nil,
			expectedError: "failed to parse CVE: failed to parse severity \"\" for CVE CVE-2025-59375: unexpected severity:  (expected one of: Low, Moderate, Important, Critical)",
		},
		{
			name:          "CVE with invalid severity",
			cves:          []jira.CVE{{Code: "CVE-2025-59375", Key: "CVE-2025-59375", Year: "2025", ComponentName: "scanner-db", Severity: "High"}},
			expected:      nil,
			expectedError: "failed to parse CVE: failed to parse severity \"High\" for CVE CVE-2025-59375: unexpected severity: High (expected one of: Low, Moderate, Important, Critical)",
		},
		{
			name:     "same CVE for multiple components",
			cves:     []jira.CVE{{Code: "CVE-2025-59375", Key: "CVE-2025-59375", Year: "2025", ComponentName: "scanner-db", Severity: "Important"}, {Code: "CVE-2025-59375", Key: "CVE-2025-59375", Year: "2025", ComponentName: "scanner-db", Severity: "Important"}},
			expected: []kflx.CVEInformation{{Key: "CVE-2025-59375", Component: "scanner-db-4-8", Severity: kflx.Severity_Important, Year: "2025"}, {Key: "CVE-2025-59375", Component: "scanner-db-4-8", Severity: kflx.Severity_Important, Year: "2025"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := renderCVEs(snapshot, application, test.cves)
			if err != nil && test.expectedError == "" {
				t.Errorf("expected no error, got %v", err)
			}
			if err != nil && err.Error() != test.expectedError {
				t.Errorf("expected error %v, got %v", test.expectedError, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

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
			err := validateProdTagFormat(tt.tag)
			if tt.expectError && err == nil {
				t.Errorf("validateTagFormat(%q) expected error, got nil", tt.tag)
			}
			if !tt.expectError && err != nil {
				t.Errorf("validateTagFormat(%q) unexpected error: %v", tt.tag, err)
			}
		})
	}
}
