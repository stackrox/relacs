package serve

import (
	"reflect"
	"testing"
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
