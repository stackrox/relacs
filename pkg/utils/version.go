package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

var semVerPattern = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-rc\.\d+)?$`)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) IsPatchRelease() bool {
	return v.Patch > 0
}

func (v Version) LessThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}

func (v Version) RenderTags() []string {
	return []string{
		fmt.Sprintf("%d.%d", v.Major, v.Minor),
		fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch),
		fmt.Sprintf("%d.%d.%d-{{ incrementer }}", v.Major, v.Minor, v.Patch),
	}
}

func NewVersion(version string) (Version, error) {
	matches := semVerPattern.FindStringSubmatch(version)
	if len(matches) == 0 {
		return Version{}, fmt.Errorf("invalid version: %s. Expected format: <major>.<minor>.<patch>-(rc.N)", version)
	}

	versionObj, err := convertToVersion(matches[1], matches[2], matches[3])
	if err != nil {
		return Version{}, fmt.Errorf("failed to convert version: %w", err)
	}
	return versionObj, nil
}

func convertToVersion(major, minor, patch string) (Version, error) {
	majorInt, err := strconv.Atoi(major)
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	minorInt, err := strconv.Atoi(minor)
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	patchInt, err := strconv.Atoi(patch)
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}

	return Version{
		Major: majorInt,
		Minor: minorInt,
		Patch: patchInt,
	}, nil
}
