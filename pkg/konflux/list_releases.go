package konflux

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stackrox/relacs/pkg/client"
)

type ReleaseResult struct {
	Name              string    `json:"name" yaml:"name"`
	ReleasePlan       string    `json:"releasePlan" yaml:"releasePlan"`
	Target            string    `json:"target" yaml:"target"`
	Snapshot          string    `json:"snapshot" yaml:"snapshot"`
	State             string    `json:"state" yaml:"state"`
	CreationTimestamp time.Time `json:"creationTimestamp" yaml:"creationTimestamp"`
	URL               string    `json:"url" yaml:"url"`
}

type ListReleasesOptions struct {
	StateFilters []string
	ProdOnly     bool
	StageOnly    bool
}

func ListReleasesForCommit(ctx context.Context, c *client.Client, commit string, opts ListReleasesOptions) ([]ReleaseResult, error) {
	labels := []string{
		fmt.Sprintf("pac.test.appstudio.openshift.io/sha=%s", commit),
	}

	releases, err := c.GetKonflux().GetReleases(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases for commit: %w", err)
	}

	stateSet := make(map[string]bool, len(opts.StateFilters))
	for _, s := range opts.StateFilters {
		stateSet[strings.ToLower(s)] = true
	}

	var results []ReleaseResult
	for _, release := range releases {
		state := release.GetState()
		target := release.GetTarget()

		if len(stateSet) > 0 && !stateSet[state] {
			continue
		}
		if opts.ProdOnly && target != "prod" {
			continue
		}
		if opts.StageOnly && target != "stage" {
			continue
		}

		url, _ := release.GetURL()
		var creationTimestamp time.Time
		if release.Metadata.CreationTimestamp != nil {
			creationTimestamp = *release.Metadata.CreationTimestamp
		}
		results = append(results, ReleaseResult{
			Name:              release.Metadata.Name,
			ReleasePlan:       release.Spec.ReleasePlan,
			Target:            target,
			Snapshot:          release.Spec.Snapshot,
			State:             state,
			CreationTimestamp: creationTimestamp,
			URL:               url,
		})
	}

	return results, nil
}
