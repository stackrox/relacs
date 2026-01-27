package konflux

import (
	"reflect"
	"testing"
)

func TestReleaseDeepCopy(t *testing.T) {
	release := Release{
		Metadata: Metadata{
			Name: "test-release",
		},
		Spec: ReleaseSpec{
			ReleasePlan: "test-release-plan",
			Snapshot:    "test-snapshot",
			Data: ReleaseData{
				Mapping: ReleaseMapping{
					Defaults: ReleaseMappingDefaults{
						Tags: []string{"test-tag"},
					},
				},
				ReleaseNotes: ReleaseNotes{
					Synopsis:   "test-synopsis",
					Type:       "test-type",
					References: []string{"test-reference"},
					Issues: ReleaseIssues{
						Fixed: []map[string]string{{"test-issue": "test-issue-value"}},
					},
					CVEs: []CVEInformation{{Key: "test-cve", Component: "test-component"}},
				},
			},
		},
		Status: &ReleaseStatus{
			Conditions: []ReleaseCondition{
				{Type: "Succeeded", Status: "True"},
			},
		},
	}
	prodRelease, err := release.DeepCopy()
	if err != nil {
		t.Errorf("unexpected error: failed to deep copy release: %v", err)
	}
	if !reflect.DeepEqual(prodRelease, release) {
		t.Errorf("deep copy release mismatch: expected %v, got %v", release, prodRelease)
	}
}
