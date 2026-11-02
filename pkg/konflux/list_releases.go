package konflux

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/github"
)

type ReleaseResult struct {
	Name              string    `json:"name" yaml:"name"`
	Application       string    `json:"application" yaml:"application"`
	ReleasePlan       string    `json:"releasePlan" yaml:"releasePlan"`
	Target            string    `json:"target" yaml:"target"`
	Snapshot          string    `json:"snapshot" yaml:"snapshot"`
	State             string    `json:"state" yaml:"state"`
	CreationTimestamp time.Time `json:"creationTimestamp" yaml:"creationTimestamp"`
	URL               string    `json:"url" yaml:"url"`
}

type ListReleasesOptions struct {
	Commit       string
	Tag          string
	Org          string
	Repo         string
	Application  string
	StateFilters []string
	ProdOnly     bool
	StageOnly    bool
}

func ListReleases(ctx context.Context, c *client.Client, opts ListReleasesOptions) ([]ReleaseResult, error) {
	commit := opts.Commit
	if opts.Tag != "" {
		commitObj, err := github.GetCommitForTag(ctx, c, opts.Org, opts.Repo, opts.Tag)
		if err != nil {
			return nil, fmt.Errorf("failed to get commit for tag: %w", err)
		}
		commit = commitObj.SHA
	}

	labels := []string{
		fmt.Sprintf("pac.test.appstudio.openshift.io/sha=%s", commit),
	}
	if opts.Application != "" {
		labels = append(labels, fmt.Sprintf("appstudio.openshift.io/application=%s", opts.Application))
	}

	releases, err := c.GetKonflux().GetReleases(ctx, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
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
			Application:       release.Metadata.Labels["appstudio.openshift.io/application"],
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
