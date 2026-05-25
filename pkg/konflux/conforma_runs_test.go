package konflux

import (
	"testing"

	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

func TestConvertToConformaRunEntry(t *testing.T) {
	tests := []struct {
		name     string
		plr      kflx.PipelineRun
		expected ConformaRunEntry
	}{
		{
			name: "succeeded run",
			plr: kflx.PipelineRun{
				Metadata: kflx.Metadata{
					Name: "acs-conforma-dev-4-10-abc12",
					Labels: map[string]string{
						"appstudio.openshift.io/application":    "acs-4-10",
						"appstudio.openshift.io/component":      "main-4-10",
						"test.appstudio.openshift.io/scenario":  "acs-conforma-dev-4-10",
					},
				},
				Status: kflx.PipelineRunStatus{
					Conditions: []kflx.PipelineRunCondition{
						{Type: "Succeeded", Status: "True", Reason: "Succeeded"},
					},
				},
			},
			expected: ConformaRunEntry{
				Name:     "acs-conforma-dev-4-10-abc12",
				Component: "main-4-10",
				Scenario: "acs-conforma-dev-4-10",
				Status:   "Succeeded",
				URL:      "https://konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com/ns/rh-acs-tenant/applications/acs-4-10/pipelineruns/acs-conforma-dev-4-10-abc12",
			},
		},
		{
			name: "failed run",
			plr: kflx.PipelineRun{
				Metadata: kflx.Metadata{
					Name: "acs-conforma-dev-4-10-xyz99",
					Labels: map[string]string{
						"appstudio.openshift.io/application":    "acs-4-10",
						"appstudio.openshift.io/component":      "scanner-4-10",
						"test.appstudio.openshift.io/scenario":  "acs-conforma-dev-4-10",
					},
				},
				Status: kflx.PipelineRunStatus{
					Conditions: []kflx.PipelineRunCondition{
						{Type: "Succeeded", Status: "False", Reason: "Failed"},
					},
				},
			},
			expected: ConformaRunEntry{
				Name:     "acs-conforma-dev-4-10-xyz99",
				Component: "scanner-4-10",
				Scenario: "acs-conforma-dev-4-10",
				Status:   "Failed",
				URL:      "https://konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com/ns/rh-acs-tenant/applications/acs-4-10/pipelineruns/acs-conforma-dev-4-10-xyz99",
			},
		},
		{
			name: "no succeeded condition shows running",
			plr: kflx.PipelineRun{
				Metadata: kflx.Metadata{
					Name: "acs-conforma-dev-4-10-noc0n",
					Labels: map[string]string{
						"appstudio.openshift.io/application":    "acs-4-10",
						"appstudio.openshift.io/component":      "main-4-10",
						"test.appstudio.openshift.io/scenario":  "acs-conforma-dev-4-10",
					},
				},
				Status: kflx.PipelineRunStatus{},
			},
			expected: ConformaRunEntry{
				Name:     "acs-conforma-dev-4-10-noc0n",
				Component: "main-4-10",
				Scenario: "acs-conforma-dev-4-10",
				Status:   "Running",
				URL:      "https://konflux-ui.apps.stone-prd-rh01.pg1f.p1.openshiftapps.com/ns/rh-acs-tenant/applications/acs-4-10/pipelineruns/acs-conforma-dev-4-10-noc0n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertToConformaRunEntry(test.plr)
			if result != test.expected {
				t.Errorf("expected %+v, got %+v", test.expected, result)
			}
		})
	}
}
