package konflux

import (
	"context"
	"fmt"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/github"
	"github.com/stackrox/relacs/pkg/utils"

	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

func GetCustomSnapshot(ctx context.Context, c *client.Client, applicationPrefix, org, repo, tag string) (kflx.Snapshot, error) {
	commitObj, err := github.GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return kflx.Snapshot{}, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	labels := getCustomSnapshotLabels(applicationPrefix, org, repo, commitObj.SHA, tag)
	snapshots, err := c.GetKonflux().GetSnapshots(labels)
	if err != nil {
		return kflx.Snapshot{}, fmt.Errorf("failed to retrieve custom snapshot: %w", err)
	}

	if len(snapshots) == 0 {
		return kflx.Snapshot{}, fmt.Errorf("no custom snapshots found for the tag %s. Has create-custom-snapshot pipeline succeeded?", tag)
	} else if len(snapshots) > 1 {
		// TODO: If there are multiple matching snapshots, we should return the one that is the most recent with a warning.
		return kflx.Snapshot{}, fmt.Errorf("multiple custom snapshots found for the tag %s. Check the PipelineRuns for the given tag, there may be re-runs", tag)
	}

	return snapshots[0], nil
}

func getCustomSnapshotLabels(applicationPrefix, organization, repository, commit, tag string) []string {
	application := utils.GetApplicationWithVersionName(applicationPrefix, tag)
	return []string{
		"pac.test.appstudio.openshift.io/url-org=" + organization,
		"pac.test.appstudio.openshift.io/url-repository=" + repository,
		"pac.test.appstudio.openshift.io/sha=" + commit,
		"acs-style-snapshot=true",
		"acs-product-version=" + tag,
		"appstudio.openshift.io/application=" + application,
	}
}

func ListFbcSnapshots(ctx context.Context, c *client.Client, commit, branch string) ([]kflx.Snapshot, error) {
	labels := getFbcSnapshotLabels(commit)
	snapshots, err := c.GetKonflux().GetSnapshots(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to get FBC Snapshot: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no FBC Snapshot found for the commit %s and branch %s", commit, branch)
	}

	filteredSnapshots := filterfbcSnapshots(snapshots, branch)
	latestSnapshots := filterLatestSnapshotPerApplication(filteredSnapshots)

	return latestSnapshots, nil
}

func getFbcSnapshotLabels(commit string) []string {
	return []string{
		"pac.test.appstudio.openshift.io/sha=" + commit,
	}
}

func filterfbcSnapshots(snapshots []kflx.Snapshot, branch string) []kflx.Snapshot {
	filteredSnapshots := []kflx.Snapshot{}
	for _, snapshot := range snapshots {
		// Must have correct source-branch
		if snapshot.Metadata.Annotations["pac.test.appstudio.openshift.io/source-branch"] != "refs/heads/"+branch {
			continue
		}
		// Must not have the original snapshot annotation, ie not be a stage or production release snapshot
		if snapshot.Metadata.Annotations["pac.test.appstudio.openshift.io/original-snapshot"] == "" {
			filteredSnapshots = append(filteredSnapshots, snapshot)
		}
	}
	return filteredSnapshots
}

func filterLatestSnapshotPerApplication(snapshots []kflx.Snapshot) []kflx.Snapshot {
	// .spec.application is the application name
	// ..metadata.creationTimestamp is the creation timestamp
	// We want to keep the snapshot with the latest creation timestamp for each application
	latestSnapshotsMaps := make(map[string]kflx.Snapshot)
	for _, snapshot := range snapshots {
		application := snapshot.Spec.Application
		creationTimestamp := snapshot.Metadata.CreationTimestamp
		_, ok := latestSnapshotsMaps[application]
		if !ok || creationTimestamp.After(latestSnapshotsMaps[application].Metadata.CreationTimestamp) {
			latestSnapshotsMaps[application] = snapshot
		}
	}

	latestSnapshots := []kflx.Snapshot{}
	for _, snapshot := range latestSnapshotsMaps {
		latestSnapshots = append(latestSnapshots, snapshot)
	}

	return latestSnapshots
}
