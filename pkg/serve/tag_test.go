package serve

import (
	"reflect"
	"testing"
	"time"

	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

func TestViolationImageGroup(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		// Image names look like
		// "<component>-<major>-<minor>-sha256:<digest>-<arch>".
		{image: "operator-5-0-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-ppc64le", expected: "operator-5-0"},
		{image: "main-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-x86-64", expected: "main-4-10"},
		// Components may themselves contain hyphens.
		{image: "scanner-db-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-s390x", expected: "scanner-db-4-10"},
		// Images without a numeric major-minor version form their own group.
		{image: "main", expected: "main"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if got := violationImageGroup(test.image); got != test.expected {
				t.Errorf("violationImageGroup(%q) = %q, expected %q", test.image, got, test.expected)
			}
		})
	}
}

func TestViolationImageArch(t *testing.T) {
	tests := []struct {
		image    string
		group    string
		expected string
	}{
		{image: "operator-5-0-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-ppc64le", group: "operator-5-0", expected: "ppc64le"},
		// Arch names may contain hyphens themselves.
		{image: "operator-5-0-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-x86-64", group: "operator-5-0", expected: "x86-64"},
		// Images without an arch suffix are the multi-arch image.
		{image: "operator-5-0-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", group: "operator-5-0", expected: "multi-arch"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if got := violationImageArch(test.image, test.group); got != test.expected {
				t.Errorf("violationImageArch(%q, %q) = %q, expected %q", test.image, test.group, got, test.expected)
			}
		})
	}
}

func TestGroupViolations(t *testing.T) {
	violations := []Violation{
		{Image: "scanner-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-s390x", Violation: "violation one"},
		{Image: "main-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", Violation: "violation two"},
		{Image: "scanner-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", Violation: "violation three"},
		{Image: "main-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", Violation: "violation four"},
	}

	groups := groupViolations(violations)

	expected := []ViolationGroup{
		{
			Name:  "main-4-10",
			Archs: []string{"multi-arch"},
			Violations: []Violation{
				{Image: "main-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", Arch: "multi-arch", Violation: "violation two"},
				{Image: "main-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", Arch: "multi-arch", Violation: "violation four"},
			},
		},
		{
			Name:  "scanner-4-10",
			Archs: []string{"multi-arch", "s390x"},
			Violations: []Violation{
				{Image: "scanner-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5-s390x", Arch: "s390x", Violation: "violation one"},
				{Image: "scanner-4-10-sha256:f0c5ee9a4de4d26e02d84154e2cd2856f133661e44c5e9dfb1867c75cd8069c5", Arch: "multi-arch", Violation: "violation three"},
			},
		},
	}

	if !reflect.DeepEqual(groups, expected) {
		t.Errorf("groupViolations() = %+v, expected %+v", groups, expected)
	}
}

func TestGroupViolations_Empty(t *testing.T) {
	if groups := groupViolations(nil); len(groups) != 0 {
		t.Errorf("groupViolations(nil) = %+v, expected no groups", groups)
	}
}

func TestGroupPipelineRuns_ComponentNames(t *testing.T) {
	newPipelineRun := func(name, pipeline, component string) kflx.PipelineRun {
		labels := map[string]string{
			"pipelinesascode.tekton.dev/original-prname": pipeline,
		}
		if component != "" {
			labels["appstudio.openshift.io/component"] = component
		}
		created := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		return kflx.PipelineRun{
			Metadata: kflx.Metadata{
				Name:              name,
				CreationTimestamp: &created,
				Labels:            labels,
			},
			Status: kflx.PipelineRunStatus{
				Conditions: []kflx.PipelineRunCondition{
					{Type: "Succeeded", Status: "True", Reason: "Succeeded"},
				},
			},
		}
	}

	plrs := []kflx.PipelineRun{
		newPipelineRun("plr-1", "operator-bundle-on-push", "operator-bundle-5-0"),
		newPipelineRun("plr-2", "operator-bundle-on-push", "operator-bundle-5-0"),
		newPipelineRun("plr-3", "create-custom-snapshot", ""),
	}

	groups, err := groupPipelineRuns(plrs)

	if err != nil {
		t.Fatalf("groupPipelineRuns() unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("groupPipelineRuns() returned %d groups, expected 2", len(groups))
	}
	byName := map[string]PipelineGroup{}
	for _, g := range groups {
		byName[g.Name] = g
	}
	bundle := byName["operator-bundle-on-push"]
	if bundle.Component != "operator-bundle-5-0" {
		t.Errorf("operator-bundle-on-push component = %q, expected operator-bundle-5-0", bundle.Component)
	}
	if len(bundle.PipelineRuns) != 2 {
		t.Errorf("operator-bundle-on-push group has %d runs, expected 2", len(bundle.PipelineRuns))
	}
	snapshot := byName["create-custom-snapshot"]
	if snapshot.Component != "create-custom-snapshot" {
		t.Errorf("create-custom-snapshot component = %q, expected fallback to the pipeline name", snapshot.Component)
	}
}

func TestGroupPipelineMilestones(t *testing.T) {
	group := func(name string, succeeded, running bool) PipelineGroup {
		return PipelineGroup{Name: name, PipelineRuns: []PipelineRunInfo{{Succeeded: succeeded, Running: running}}}
	}

	groups := []PipelineGroup{
		group("operator-index-on-push", true, false),
		group("main-on-push", false, true),
		group("create-custom-snapshot", true, false),
		group("operator-bundle-on-push", false, false),
		group("retag-main", true, false),
	}

	milestones := groupPipelineMilestones(groups)

	expected := []PipelineMilestone{
		{
			Name:            "Operands & Operator",
			Status:          "running", // main-on-push is running, retag-main passed
			Groups:          []PipelineGroup{group("main-on-push", false, true), group("retag-main", true, false)},
			FailedPipelines: nil, // the running group is not failed
			Retestable:      true,
		},
		{
			Name:            "Operator Bundle",
			Status:          "failed",
			Groups:          []PipelineGroup{group("operator-bundle-on-push", false, false)},
			FailedPipelines: []string{"operator-bundle-on-push"},
			Retestable:      false, // the previous milestone is still running
		},
		{
			Name:       "Custom Snapshot & Index Image",
			Status:     "passed",
			Groups:     []PipelineGroup{group("operator-index-on-push", true, false), group("create-custom-snapshot", true, false)},
			Retestable: false, // the previous milestone failed
		},
	}

	if !reflect.DeepEqual(milestones, expected) {
		t.Errorf("groupPipelineMilestones() = %+v, expected %+v", milestones, expected)
	}
}

func TestGroupPipelineMilestones_Retestable(t *testing.T) {
	group := func(name string, succeeded, running bool) PipelineGroup {
		return PipelineGroup{Name: name, PipelineRuns: []PipelineRunInfo{{Succeeded: succeeded, Running: running}}}
	}

	groups := []PipelineGroup{
		group("main-on-push", true, false),
		group("operator-bundle-on-push", false, false),
		group("create-custom-snapshot", true, false),
	}

	milestones := groupPipelineMilestones(groups)

	// The first milestone passed, so the failed Operator Bundle milestone is
	// retestable; it failed, so the last milestone is not.
	if !milestones[1].Retestable {
		t.Error("Operator Bundle milestone expected to be retestable")
	}
	if !reflect.DeepEqual(milestones[1].FailedPipelines, []string{"operator-bundle-on-push"}) {
		t.Errorf("Operator Bundle FailedPipelines = %v, expected [operator-bundle-on-push]", milestones[1].FailedPipelines)
	}
	if milestones[2].Retestable {
		t.Error("Custom Snapshot & Index Image milestone expected not to be retestable")
	}
}

func TestGroupPipelineMilestones_DropsEmptyMilestones(t *testing.T) {
	// A non-main repository has no operator bundle/index pipelines.
	groups := []PipelineGroup{{Name: "main-on-push", PipelineRuns: []PipelineRunInfo{{Succeeded: true}}}}

	milestones := groupPipelineMilestones(groups)

	if len(milestones) != 1 || milestones[0].Name != "Operands & Operator" {
		t.Errorf("groupPipelineMilestones() = %+v, expected only the Operands & Operator milestone", milestones)
	}
}

func TestMilestoneStatus(t *testing.T) {
	tests := []struct {
		name     string
		groups   []PipelineGroup
		expected string
	}{
		{
			name:     "passed when all latest runs succeeded",
			groups:   []PipelineGroup{{PipelineRuns: []PipelineRunInfo{{Succeeded: true}}}},
			expected: "passed",
		},
		{
			name:     "running while a latest run is running",
			groups:   []PipelineGroup{{PipelineRuns: []PipelineRunInfo{{Running: true}}}},
			expected: "running",
		},
		{
			name:     "failed when a latest run failed",
			groups:   []PipelineGroup{{PipelineRuns: []PipelineRunInfo{{}}}},
			expected: "failed",
		},
		{
			name: "failed takes precedence over running",
			groups: []PipelineGroup{
				{PipelineRuns: []PipelineRunInfo{{Running: true}}},
				{PipelineRuns: []PipelineRunInfo{{}}},
			},
			expected: "failed",
		},
		{
			name:     "passed for groups without runs",
			groups:   []PipelineGroup{{}},
			expected: "passed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := milestoneStatus(test.groups); got != test.expected {
				t.Errorf("milestoneStatus() = %q, expected %q", got, test.expected)
			}
		})
	}
}

func TestComputeTagOverallStatus(t *testing.T) {
	buildGroup := func(name string, succeeded, running bool) PipelineGroup {
		return PipelineGroup{Name: name, PipelineRuns: []PipelineRunInfo{{Succeeded: succeeded, Running: running}}}
	}
	mainRepoGreen := TagData{
		PipelineGroups:         []PipelineGroup{buildGroup("build", true, false)},
		SnapshotName:           "snap",
		OperatorBundleImage:    "bundle",
		OperatorIndexImage:     "index",
		ConformaReleasePreview: ConformaResults{},
	}

	tests := []struct {
		name      string
		isMainRep bool
		data      TagData
		expected  string
	}{
		{
			name:      "green when everything is satisfied",
			isMainRep: true,
			data:      mainRepoGreen,
			expected:  "green",
		},
		{
			name:      "error when pipeline runs could not be retrieved",
			isMainRep: true,
			data:      TagData{PipelinesError: true},
			expected:  "error",
		},
		{
			name:      "error when dev conforma results could not be retrieved",
			isMainRep: true,
			data:      TagData{ConformaDev: ConformaResults{Error: true}},
			expected:  "error",
		},
		{
			name:      "error when release preview results could not be retrieved",
			isMainRep: true,
			data:      TagData{ConformaReleasePreview: ConformaResults{Error: true}},
			expected:  "error",
		},
		{
			name:      "error when the custom snapshot is missing",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.SnapshotName = ""
				return d
			}(),
			expected: "error",
		},
		{
			name:      "error when the operator bundle image is missing",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.OperatorBundleImage = ""
				return d
			}(),
			expected: "error",
		},
		{
			name:      "error when the operator index image is missing",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.OperatorIndexImage = ""
				return d
			}(),
			expected: "error",
		},
		{
			name:      "failed when there are release preview violations",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.ConformaReleasePreview.Violations = []Violation{{Image: "main-4-10", Violation: "v"}}
				return d
			}(),
			expected: "failed",
		},
		{
			name:      "failed when release preview runs were skipped",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.ConformaReleasePreview.SkippedRuns = []ConformaRun{{Name: "run"}}
				return d
			}(),
			expected: "failed",
		},
		{
			name:      "failed when the latest build pipeline run failed",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.PipelineGroups = []PipelineGroup{buildGroup("build", false, false)}
				return d
			}(),
			expected: "failed",
		},
		{
			name:      "waiting when release preview runs are in progress",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.ConformaReleasePreview.InProgressRuns = []ConformaRun{{Name: "run"}}
				return d
			}(),
			expected: "waiting",
		},
		{
			name:      "waiting when the latest build pipeline run is running",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.PipelineGroups = []PipelineGroup{buildGroup("build", false, true)}
				return d
			}(),
			expected: "waiting",
		},
		{
			name:      "failed takes precedence over waiting",
			isMainRep: true,
			data: func() TagData {
				d := mainRepoGreen
				d.PipelineGroups = []PipelineGroup{buildGroup("a", false, true), buildGroup("b", false, false)}
				return d
			}(),
			expected: "failed",
		},
		{
			name:      "green for other repositories with only green builds",
			isMainRep: false,
			data:      TagData{PipelineGroups: []PipelineGroup{buildGroup("build", true, false)}},
			expected:  "green",
		},
		{
			name:      "waiting for other repositories with a running build",
			isMainRep: false,
			data:      TagData{PipelineGroups: []PipelineGroup{buildGroup("build", false, true)}},
			expected:  "waiting",
		},
		{
			name:      "failed for other repositories with a failed build",
			isMainRep: false,
			data:      TagData{PipelineGroups: []PipelineGroup{buildGroup("build", false, false)}},
			expected:  "failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			computeTagOverallStatus(&test.data, test.isMainRep)
			if test.data.OverallStatus != test.expected {
				t.Errorf("computeTagOverallStatus() = %q, expected %q", test.data.OverallStatus, test.expected)
			}
		})
	}
}

func TestComputeTagOverallStatus_PipelinesRunning(t *testing.T) {
	buildGroup := func(name string, succeeded, running bool) PipelineGroup {
		return PipelineGroup{Name: name, PipelineRuns: []PipelineRunInfo{{Succeeded: succeeded, Running: running}}}
	}

	tests := []struct {
		name          string
		groups        []PipelineGroup
		expectRunning bool
	}{
		{
			name:          "running when all latest runs are running",
			groups:        []PipelineGroup{buildGroup("a", false, true), buildGroup("b", false, true)},
			expectRunning: true,
		},
		{
			name:          "not running when a latest run failed",
			groups:        []PipelineGroup{buildGroup("a", false, true), buildGroup("b", false, false)},
			expectRunning: false,
		},
		{
			name:          "not running when all latest runs succeeded",
			groups:        []PipelineGroup{buildGroup("a", true, false)},
			expectRunning: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := TagData{PipelineGroups: test.groups}
			computeTagOverallStatus(&data, false)
			if data.PipelinesRunning != test.expectRunning {
				t.Errorf("PipelinesRunning = %v, expected %v", data.PipelinesRunning, test.expectRunning)
			}
		})
	}
}

func TestComputeGroupStatus_Running(t *testing.T) {
	newer := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	older := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	groups := []PipelineGroup{{
		Name: "scanner-on-push",
		PipelineRuns: []PipelineRunInfo{
			{Name: "old", CreationTimestamp: older, Succeeded: true},
			{Name: "new", CreationTimestamp: newer, Running: true},
		},
	}}

	ComputeGroupStatus(groups)

	if groups[0].AllPassed {
		t.Error("AllPassed = true, expected false for a running latest run")
	}
	if !groups[0].Running {
		t.Error("Running = false, expected true for a running latest run")
	}
	if groups[0].PipelineRuns[0].Name != "new" {
		t.Errorf("latest run = %q, expected the newest run to be sorted first", groups[0].PipelineRuns[0].Name)
	}
}
