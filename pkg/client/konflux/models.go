package konflux

import (
	"encoding/json"
	"fmt"
	"time"
)

type SnapshotList struct {
	Items []Snapshot `json:"items"`
}

type Snapshot struct {
	Metadata Metadata     `json:"metadata"`
	Spec     SnapshotSpec `json:"spec"`
}

type Metadata struct {
	Name              string            `json:"name" yaml:"name"`
	Namespace         string            `json:"namespace" yaml:"namespace"`
	CreationTimestamp *time.Time        `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type SnapshotSpec struct {
	Components []SnapshotComponent `json:"components"`
}

type SnapshotComponent struct {
	Name           string `json:"name"`
	ContainerImage string `json:"containerImage"`
}

type ReleaseList struct {
	Items []Release `json:"items"`
}

type Release struct {
	Kind       string         `json:"kind" yaml:"kind"`
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Metadata   Metadata       `json:"metadata" yaml:"metadata"`
	Spec       ReleaseSpec    `json:"spec" yaml:"spec"`
	Status     *ReleaseStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (r Release) String() string {
	return fmt.Sprintf("Name: %s\nNamespace: %s\nRelease Plan: %s\nSnapshot: %s", r.Metadata.Name, r.Metadata.Namespace, r.Spec.ReleasePlan, r.Spec.Snapshot)
}

func (r Release) DeepCopy() (Release, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return Release{}, fmt.Errorf("failed to marshal release for deep copy: %w", err)
	}
	var copy Release
	err = json.Unmarshal(jsonBytes, &copy)
	if err != nil {
		return Release{}, fmt.Errorf("failed to unmarshal release for deep copy: %w", err)
	}
	return copy, nil
}

type ReleaseSpec struct {
	ReleasePlan string      `json:"releasePlan" yaml:"releasePlan"`
	Snapshot    string      `json:"snapshot"`
	Data        ReleaseData `json:"data"`
}

type ReleaseData struct {
	Mapping      ReleaseMapping `json:"mapping"`
	ReleaseNotes ReleaseNotes   `json:"releaseNotes" yaml:"releaseNotes"`
}

type ReleaseMapping struct {
	Defaults ReleaseMappingDefaults `json:"defaults"`
}

type ReleaseMappingDefaults struct {
	Tags []string `json:"tags"`
}

type ReleaseType int

const (
	ReleaseType_RHSA ReleaseType = iota
	ReleaseType_RHBA
	ReleaseType_RHEA
)

var releaseTypeStrings = []string{
	"RHSA",
	"RHBA",
	"RHEA",
}

func (r ReleaseType) String() string {
	return releaseTypeStrings[r]
}

type ReleaseNotes struct {
	Synopsis   string           `json:"synopsis"`
	Type       string           `json:"type"`
	References []string         `json:"references"`
	Issues     ReleaseIssues    `json:"issues"`
	CVEs       []CVEInformation `json:"cves"`
}

type ReleaseIssues struct {
	Fixed []map[string]string `json:"fixed"`
}

type ReleaseStatus struct {
	Artifacts      ReleaseArtifacts   `json:"artifacts"`
	CompletionTime *time.Time         `json:"completionTime"`
	Conditions     []ReleaseCondition `json:"conditions"`
}

type ReleaseCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type ReleaseArtifacts struct {
	Advisory ReleaseAdvisory `json:"advisory"`
}

type ReleaseAdvisory struct {
	InternalURL string `json:"internal_url"`
	URL         string `json:"url"`
}

type Advisory struct {
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     AdvisorySpec `json:"spec" yaml:"spec"`
}

type AdvisorySpec struct {
	Content AdvisoryContent `json:"content" yaml:"content"`
}

type AdvisoryContent struct {
	Images []AdvisoryImage `json:"images" yaml:"images"`
}

type AdvisoryImage struct {
	Component      string `json:"component" yaml:"component"`
	ContainerImage string `json:"containerImage" yaml:"containerImage"`
}

type PipelineRunList struct {
	Items []PipelineRun `json:"items"`
}

type PipelineRun struct {
	Metadata Metadata          `json:"metadata"`
	Status   PipelineRunStatus `json:"status"`
}

type PipelineRunStatus struct {
	Conditions []PipelineRunCondition `json:"conditions"`
}

type PipelineRunCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// Severity is the severity of a CVE.
type Severity int

const (
	Severity_Low Severity = iota
	Severity_Moderate
	Severity_Important
	Severity_Critical
)

var severityStrings = []string{
	"Low",
	"Moderate",
	"Important",
	"Critical",
}

func (s Severity) String() string {
	return severityStrings[s]
}

// Application represents a Konflux application
type Application struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

// ApplicationList represents a list of Konflux applications
type ApplicationList struct {
	Items []Application `json:"items"`
}

// ProductStream represents a product version stream
type ProductStream struct {
	Major int
	Minor int
}

// String returns the string representation of a product stream (e.g., "4.6")
func (ps ProductStream) String() string {
	return fmt.Sprintf("%d.%d", ps.Major, ps.Minor)
}
