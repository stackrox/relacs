package konflux

import (
	"context"
	"fmt"
	"os"
	"sort"

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
		// If there are multiple matching snapshots, return the most recent one with a warning.
		sort.Slice(snapshots, func(i, j int) bool {
			return snapshots[i].Metadata.CreationTimestamp.After(snapshots[j].Metadata.CreationTimestamp)
		})
		fmt.Fprintf(os.Stderr, "WARNING: Multiple custom snapshots found for the tag %s. Check the PipelineRuns for the given tag, there may be re-runs. Returning the latest snapshot.\n", tag)
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
