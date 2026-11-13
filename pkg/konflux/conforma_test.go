package konflux

import (
	"testing"
	"time"

	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

func TestLatestPipelineRunPerComponentAndScenario(t *testing.T) {
	older := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	newPipelineRun := func(name, component, scenario string, creationTimestamp time.Time) kflx.PipelineRun {
		return kflx.PipelineRun{
			Metadata: kflx.Metadata{
				Name:              name,
				CreationTimestamp: &creationTimestamp,
				Labels: map[string]string{
					"appstudio.openshift.io/component":     component,
					"test.appstudio.openshift.io/scenario": scenario,
				},
			},
		}
	}

	tests := []struct {
		name     string
		plrs     []kflx.PipelineRun
		expected []string // names of the remaining pipeline runs
	}{
		{
			name: "keeps only the newest run per component and scenario",
			plrs: []kflx.PipelineRun{
				newPipelineRun("main-old", "main-4-10", "acs-conforma-dev-4-10", older),
				newPipelineRun("main-new", "main-4-10", "acs-conforma-dev-4-10", newer),
			},
			expected: []string{"main-new"},
		},
		{
			name: "keeps runs for different components of the same scenario",
			plrs: []kflx.PipelineRun{
				newPipelineRun("main", "main-4-10", "acs-conforma-dev-4-10", older),
				newPipelineRun("scanner", "scanner-4-10", "acs-conforma-dev-4-10", older),
			},
			expected: []string{"main", "scanner"},
		},
		{
			name: "keeps runs for different scenarios of the same component",
			plrs: []kflx.PipelineRun{
				newPipelineRun("dev", "main-4-10", "acs-conforma-dev-4-10", older),
				newPipelineRun("other", "main-4-10", "acs-conforma-other-4-10", older),
			},
			expected: []string{"dev", "other"},
		},
		{
			name:     "returns no runs for empty input",
			plrs:     []kflx.PipelineRun{},
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runs := latestPipelineRunPerComponentAndScenario(test.plrs)
			if len(runs) != len(test.expected) {
				t.Fatalf("latestPipelineRunPerComponentAndScenario() returned %d runs, expected %d", len(runs), len(test.expected))
			}
			found := make(map[string]bool)
			for _, run := range runs {
				found[run.Metadata.Name] = true
			}
			for _, name := range test.expected {
				if !found[name] {
					t.Errorf("latestPipelineRunPerComponentAndScenario() result missing run %s", name)
				}
			}
		})
	}
}

func TestIsPipelineRunInProgress(t *testing.T) {
	tests := []struct {
		name       string
		conditions []kflx.PipelineRunCondition
		expected   bool
	}{
		{
			name:       "no conditions yet",
			conditions: []kflx.PipelineRunCondition{},
			expected:   true,
		},
		{
			name: "succeeded condition with unknown status and started reason",
			conditions: []kflx.PipelineRunCondition{
				{Type: "Succeeded", Status: "Unknown", Reason: "Started"},
			},
			expected: true,
		},
		{
			name: "succeeded condition with unknown status and running reason",
			conditions: []kflx.PipelineRunCondition{
				{Type: "Succeeded", Status: "Unknown", Reason: "Running"},
			},
			expected: true,
		},
		{
			name: "succeeded condition with true status",
			conditions: []kflx.PipelineRunCondition{
				{Type: "Succeeded", Status: "True", Reason: "Succeeded"},
			},
			expected: false,
		},
		{
			name: "succeeded condition with false status",
			conditions: []kflx.PipelineRunCondition{
				{Type: "Succeeded", Status: "False", Reason: "Failed"},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isPipelineRunInProgress(test.conditions); got != test.expected {
				t.Errorf("isPipelineRunInProgress() = %v, expected %v", got, test.expected)
			}
		})
	}
}

func TestParseConformaViolations(t *testing.T) {
	tests := []struct {
		name        string
		report      string
		expected    Violations
		expectedErr bool
	}{
		{
			name:   "multiple components with violations",
			report: `{"components": [{"name": "main", "violations": [{"msg": "violation one"}, {"msg": "violation two"}]}, {"name": "scanner", "violations": [{"msg": "violation three"}]}]}`,
			expected: Violations{
				Violations: []Violation{
					{Image: "main", Violation: "violation one"},
					{Image: "main", Violation: "violation two"},
					{Image: "scanner", Violation: "violation three"},
				},
			},
		},
		{
			name:     "no components",
			report:   `{"components": []}`,
			expected: Violations{},
		},
		{
			name:     "component without violations",
			report:   `{"components": [{"name": "main", "violations": []}]}`,
			expected: Violations{},
		},
		{
			name:        "invalid JSON",
			report:      `not json`,
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			violations, err := parseConformaViolations(test.report)
			if test.expectedErr {
				if err == nil {
					t.Fatalf("parseConformaViolations() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseConformaViolations() unexpected error: %v", err)
			}
			if len(violations.Violations) != len(test.expected.Violations) {
				t.Fatalf("parseConformaViolations() returned %d violations, expected %d", len(violations.Violations), len(test.expected.Violations))
			}
			for i, violation := range violations.Violations {
				if violation != test.expected.Violations[i] {
					t.Errorf("parseConformaViolations() violation %d = %+v, expected %+v", i, violation, test.expected.Violations[i])
				}
			}
		})
	}
}
