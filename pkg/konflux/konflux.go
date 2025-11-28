package konflux

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/stackrox/rel-acs/pkg/client"
	"github.com/stackrox/rel-acs/pkg/github"
	"github.com/stackrox/rel-acs/pkg/utils"
	"gopkg.in/yaml.v3"
)

// SnapshotList represents a Kubernetes list of snapshots
type SnapshotList struct {
	Items []Snapshot `yaml:"items"`
}

// Snapshot represents a Konflux snapshot resource
type Snapshot struct {
	Metadata SnapshotMetadata `yaml:"metadata"`
}

// SnapshotMetadata represents the metadata section of a snapshot
type SnapshotMetadata struct {
	Name string `yaml:"name"`
}

// FindCustomSnapshotForTag finds a Konflux custom snapshot for a given Git tag
// by first resolving the tag to a commit SHA, then querying for snapshots
func FindCustomSnapshotForTag(ctx context.Context, c *client.Client, organization, repository, tag string) (string, error) {
	commit, err := github.GetCommitForTag(ctx, c, organization, repository, tag)
	if err != nil {
		return "", err
	}

	// TODO: Get the application name from user input
	application := utils.GetApplicationVersionName("acs", tag)
	labels := []string{
		"pac.test.appstudio.openshift.io/url-org=" + organization,
		"pac.test.appstudio.openshift.io/url-repository=" + repository,
		"pac.test.appstudio.openshift.io/sha=" + commit,
		"acs-style-snapshot=true",
		"acs-product-version=" + tag,
		"appstudio.openshift.io/application=" + application,
	}

	query := strings.Join(labels, ",")
	snapshotsYAML, err := getSnapshots(ctx, c, query)
	if err != nil {
		return "", err
	}

	names, err := parseSnapshotNamesFromSnapshotList(snapshotsYAML)
	if err != nil {
		return "", err
	}

	if len(names) != 1 {
		return "", fmt.Errorf("expected 1 snapshot, got %d. Check the PipelineRuns for the given tag", len(names))
	}
	return names[0], nil
}

func getSnapshots(ctx context.Context, _ *client.Client, query string) (string, error) {
	command := fmt.Sprintf("kubectl ka get snapshot -l %s -o yaml", query)
	output, err := exec.CommandContext(ctx, "sh", "-c", command).Output()
	if err != nil {
		if len(output) > 0 {
			return "", fmt.Errorf("failed to get snapshots: %s", string(output))
		}
		return "", nil
	}
	return string(output), nil
}

func parseSnapshotNamesFromSnapshotList(snapshots string) ([]string, error) {
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found for the given labels")
	}

	// Parse the snapshots YAML and return the snapshot names
	var snapshotList SnapshotList
	if err := yaml.Unmarshal([]byte(snapshots), &snapshotList); err != nil {
		return nil, fmt.Errorf("failed to parse snapshots YAML: %w", err)
	}

	if len(snapshotList.Items) == 0 {
		return nil, fmt.Errorf("no snapshots found in the YAML output")
	}

	// Extract snapshot names and join them with newlines
	var names []string
	for _, snapshot := range snapshotList.Items {
		if snapshot.Metadata.Name != "" {
			names = append(names, snapshot.Metadata.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no valid snapshot names found")
	}

	return names, nil
}
