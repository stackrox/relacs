package konflux

type SnapshotList struct {
	Items []Snapshot `json:"items"`
}

type Snapshot struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Name string `json:"name"`
}
