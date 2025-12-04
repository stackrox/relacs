package konflux

import "time"

type SnapshotList struct {
	Items []Snapshot `json:"items"`
}

type Snapshot struct {
	Metadata Metadata     `json:"metadata"`
	Spec     SnapshotSpec `json:"spec"`
}

type Metadata struct {
	Name string `json:"name"`
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
	Metadata Metadata      `json:"metadata"`
	Spec     ReleaseSpec   `json:"spec"`
	Status   ReleaseStatus `json:"status"`
}

type ReleaseSpec struct {
	ReleasePlan string `json:"releasePlan"`
	Snapshot    string `json:"snapshot"`
}

type ReleaseStatus struct {
	Artifacts      ReleaseArtifacts `json:"artifacts"`
	CompletionTime *time.Time       `json:"completionTime,omitempty"`
}

type ReleaseArtifacts struct {
	Advisory ReleaseAdvisory `json:"advisory"`
}

type ReleaseAdvisory struct {
	InternalURL string `json:"internal_url"`
	URL         string `json:"url"`
}

type Advisory struct {
	Metadata Metadata     `yaml:"metadata"`
	Spec     AdvisorySpec `yaml:"spec"`
}

type AdvisorySpec struct {
	Content AdvisoryContent `yaml:"content"`
}

type AdvisoryContent struct {
	Images []AdvisoryImage `yaml:"images"`
}

type AdvisoryImage struct {
	Component      string `yaml:"component"`
	ContainerImage string `yaml:"containerImage"`
}
