package konflux

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/github"
	"github.com/stackrox/relacs/pkg/utils"

	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

func GetCustomSnapshot(ctx context.Context, c *client.Client, applicationPrefix, org, repo, tag string) (kflx.Snapshot, error) {
	cfg := c.GetConfig()
	commitObj, err := github.GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return kflx.Snapshot{}, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	labels := getCustomSnapshotLabels(cfg, applicationPrefix, org, repo, commitObj.SHA, tag)
	snapshots, err := c.GetKonflux().GetSnapshots(labels)
	if err != nil {
		return kflx.Snapshot{}, fmt.Errorf("failed to retrieve custom snapshot: %w", err)
	}

	if len(snapshots) == 0 {
		return kflx.Snapshot{}, fmt.Errorf("no custom snapshots found for the tag %s. Has create-custom-snapshot pipeline succeeded?", tag)
	} else if len(snapshots) > 1 {
		// If there are multiple matching snapshots, return the most recent one with a warning.
		sort.Slice(snapshots, func(i, j int) bool {
			return snapshots[i].Metadata.CreationTimestamp.After(*snapshots[j].Metadata.CreationTimestamp)
		})
		fmt.Fprintf(os.Stderr, "WARNING: Multiple custom snapshots found for the tag %s. Check the PipelineRuns for the given tag, there may be re-runs. Returning the latest snapshot.\n", tag)
	}

	return snapshots[0], nil
}

func getCustomSnapshotLabels(cfg *config.Config, applicationPrefix, organization, repository, commit, tag string) []string {
	application := utils.Application{Name: applicationPrefix}.WithVersion(tag)
	productVersion := utils.GetProductVersionFromTag(cfg, tag)
	return []string{
		"pac.test.appstudio.openshift.io/url-org=" + organization,
		"pac.test.appstudio.openshift.io/url-repository=" + repository,
		"pac.test.appstudio.openshift.io/sha=" + commit,
		"acs-style-snapshot=true",
		"acs-product-version=" + productVersion,
		"appstudio.openshift.io/application=" + application,
	}
}

func PruneSnapshots(ctx context.Context, c *client.Client, version string, dryRun bool) ([]string, error) {
	// Get all snapshots from the real cluster (not KubeArchive)
	labels := []string{
		"acs-style-snapshot=true",
	}

	snapshots, err := c.GetKonflux().GetSnapshotsFromCluster(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve snapshots from cluster: %w", err)
	}

	var matchingSnapshots []string
	annotationKey := "test.appstudio.openshift.io/keep-snapshot"

	for _, snapshot := range snapshots {
		// Check if the snapshot has the keep-snapshot annotation
		if _, hasAnnotation := snapshot.Metadata.Annotations[annotationKey]; !hasAnnotation {
			continue
		}

		// Check if the version label matches (starts with the specified version)
		versionLabel, hasVersion := snapshot.Metadata.Labels["acs-product-version"]
		if !hasVersion {
			fmt.Fprintf(os.Stderr, "WARNING: Snapshot %s has keep-snapshot annotation but is missing acs-product-version label. Skipping.\n", snapshot.Metadata.Name)
			continue
		}

		if strings.HasPrefix(versionLabel, version) {
			matchingSnapshots = append(matchingSnapshots, snapshot.Metadata.Name)
		}
	}

	if !dryRun {
		for _, snapshotName := range matchingSnapshots {
			err := c.GetKonflux().RemoveSnapshotAnnotationFromCluster(snapshotName, annotationKey)
			if err != nil {
				return matchingSnapshots, fmt.Errorf("failed to remove annotation from snapshot %s: %w", snapshotName, err)
			}
		}
	}

	return matchingSnapshots, nil
}
