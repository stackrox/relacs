package konflux

import (
	"testing"
)

func TestGetState(t *testing.T) {
	tests := []struct {
		name     string
		release  Release
		expected string
	}{
		{
			name:     "nil status",
			release:  Release{},
			expected: "unknown",
		},
		{
			name: "no conditions",
			release: Release{
				Status: &ReleaseStatus{},
			},
			expected: "unknown",
		},
		{
			name: "Released condition with Succeeded reason",
			release: Release{
				Status: &ReleaseStatus{
					Conditions: []ReleaseCondition{
						{Type: "Released", Status: "True", Reason: "Succeeded"},
					},
				},
			},
			expected: "succeeded",
		},
		{
			name: "Released condition with Failed reason",
			release: Release{
				Status: &ReleaseStatus{
					Conditions: []ReleaseCondition{
						{Type: "Released", Status: "False", Reason: "Failed"},
					},
				},
			},
			expected: "failed",
		},
		{
			name: "Released condition with Progressing reason",
			release: Release{
				Status: &ReleaseStatus{
					Conditions: []ReleaseCondition{
						{Type: "Released", Status: "Unknown", Reason: "Progressing"},
					},
				},
			},
			expected: "progressing",
		},
		{
			name: "Released condition with unexpected reason",
			release: Release{
				Status: &ReleaseStatus{
					Conditions: []ReleaseCondition{
						{Type: "Released", Status: "Unknown", Reason: "SomethingElse"},
					},
				},
			},
			expected: "somethingelse",
		},
		{
			name: "only non-Released conditions",
			release: Release{
				Status: &ReleaseStatus{
					Conditions: []ReleaseCondition{
						{Type: "Validated", Status: "True", Reason: "Succeeded"},
					},
				},
			},
			expected: "unknown",
		},
		{
			name: "multiple conditions with Released present",
			release: Release{
				Status: &ReleaseStatus{
					Conditions: []ReleaseCondition{
						{Type: "Validated", Status: "True", Reason: "Succeeded"},
						{Type: "Released", Status: "True", Reason: "Succeeded"},
					},
				},
			},
			expected: "succeeded",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.release.GetState()
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestGetTarget(t *testing.T) {
	tests := []struct {
		name     string
		release  Release
		expected string
	}{
		{
			name: "stage release plan",
			release: Release{
				Spec: ReleaseSpec{ReleasePlan: "acs-stage-4-10"},
			},
			expected: "stage",
		},
		{
			name: "prod release plan",
			release: Release{
				Spec: ReleaseSpec{ReleasePlan: "acs-prod-4-10"},
			},
			expected: "prod",
		},
		{
			name: "unknown release plan",
			release: Release{
				Spec: ReleaseSpec{ReleasePlan: "something-else"},
			},
			expected: "unknown",
		},
		{
			name:     "empty release plan",
			release:  Release{},
			expected: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.release.GetTarget()
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}
