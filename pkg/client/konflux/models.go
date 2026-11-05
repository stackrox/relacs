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

func (s Snapshot) GetUID() string { return s.Metadata.UID }

type Metadata struct {
	Name              string            `json:"name" yaml:"name"`
	Namespace         string            `json:"namespace" yaml:"namespace"`
	UID               string            `json:"uid,omitempty" yaml:"uid,omitempty"`
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

func (r Release) GetUID() string { return r.Metadata.UID }

func (r Release) String() string {
	return fmt.Sprintf("Name: %s\nNamespace: %s\nRelease Plan: %s\nSnapshot: %s", r.Metadata.Name, r.Metadata.Namespace, r.Spec.ReleasePlan, r.Spec.Snapshot)
}

type CdnRelease struct {
	Kind       string         `json:"kind" yaml:"kind"`
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Metadata   Metadata       `json:"metadata" yaml:"metadata"`
	Spec       CdnReleaseSpec `json:"spec" yaml:"spec"`
}

func (r CdnRelease) String() string {
	return fmt.Sprintf("Name: %s\nNamespace: %s\nRelease Plan: %s\nSnapshot: %s", r.Metadata.Name, r.Metadata.Namespace, r.Spec.ReleasePlan, r.Spec.Snapshot)
}

type CdnReleaseSpec struct {
	ReleasePlan string         `json:"releasePlan" yaml:"releasePlan"`
	Snapshot    string         `json:"snapshot" yaml:"snapshot"`
	Data        CdnReleaseData `json:"data" yaml:"data"`
}

type CdnReleaseData struct {
	Mapping      CdnReleaseMapping   `json:"mapping" yaml:"mapping"`
	ReleaseNotes CdnReleaseNotes     `json:"releaseNotes" yaml:"releaseNotes"`
}

type CdnReleaseMapping struct {
	Components []CdnReleaseComponent `json:"components" yaml:"components"`
}

type CdnReleaseComponent struct {
	Name   string                  `json:"name" yaml:"name"`
	Staged CdnReleaseStagedVersion `json:"staged" yaml:"staged"`
}

type CdnReleaseStagedVersion struct {
	Version string `json:"version" yaml:"version"`
}

type CdnReleaseNotes struct {
	Type       string   `json:"type" yaml:"type"`
	Synopsis   string   `json:"synopsis" yaml:"synopsis"`
	References []string `json:"references" yaml:"references"`
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

func (p PipelineRun) GetUID() string { return p.Metadata.UID }

type PipelineRunStatus struct {
	Conditions []PipelineRunCondition `json:"conditions"`
	Results    []PipelineRunResult    `json:"results,omitempty"`
}

type PipelineRunResult struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

func (r PipelineRunResult) StringValue() (string, error) {
	var s string
	if err := json.Unmarshal(r.Value, &s); err != nil {
		return "", fmt.Errorf("result %q is not a JSON string: %s", r.Name, string(r.Value))
	}
	return s, nil
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
