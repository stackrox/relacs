package konflux

import (
	"context"
	"fmt"
	"sort"

	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/github"
)

func ListPipelineRunsForTag(ctx context.Context, c *client.Client, org, repo, tag string) ([]kflx.PipelineRun, error) {
	commit, err := github.GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	labels := getLabelsForPipelineRunsForTag(org, repo, commit.SHA)
	pipelineRuns, err := c.GetKonflux().GetPipelineRuns(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline runs: %w", err)
	}

	// Filter the pipeline runs to only ones for the given tag, including retest/test tag:<tag> comments
	filteredPLRs := []kflx.PipelineRun{}
	for _, plr := range pipelineRuns {
		if plr.Metadata.Annotations["pipelinesascode.tekton.dev/branch"] == "refs/tags/"+tag {
			filteredPLRs = append(filteredPLRs, plr)
		}
	}

	sort.Slice(filteredPLRs, func(i, j int) bool {
		return filteredPLRs[i].Metadata.CreationTimestamp.Before(filteredPLRs[j].Metadata.CreationTimestamp)
	})
	return filteredPLRs, nil
}

func getLabelsForPipelineRunsForTag(org, repo, commit string) []string {
	return []string{
		"pipelinesascode.tekton.dev/sha=" + commit,
		"pipelinesascode.tekton.dev/url-org=" + org,
		"pipelinesascode.tekton.dev/url-repository=" + repo,
	}
}
