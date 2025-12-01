package konflux

// Metadata represents the metadata section of a K8s resource
type Metadata struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels"`
}

// SnapshotList represents a Kubernetes list of snapshots
type SnapshotList struct {
	Items []Snapshot `yaml:"items"`
}

// Snapshot represents a Konflux snapshot resource
type Snapshot struct {
	Metadata Metadata `yaml:"metadata"`
}

type ReleaseList struct {
	Items []Release `yaml:"items"`
}

// Release represents a Konflux release resource
type Release struct {
	Metadata Metadata `yaml:"metadata"`
	Status   Status   `yaml:"status"`
}

type Status struct {
	Artifacts Artifacts `yaml:"artifacts"`
}

type Artifacts struct {
	Advisory ReleaseAdvisory `yaml:"advisory"`
}

type ReleaseAdvisory struct {
	InternalURL string `yaml:"internal_url"`
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
