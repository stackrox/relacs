package konflux

import (
	"context"
	"fmt"
	"os"
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
	pipelineRuns, err := c.GetKonflux().GetPipelineRuns(ctx, labels)
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
		return filteredPLRs[i].Metadata.CreationTimestamp.Before(*filteredPLRs[j].Metadata.CreationTimestamp)
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

type OperatorIndexImage struct {
	ImageURL    string
	ImageDigest string
}

func GetOperatorIndexImage(ctx context.Context, c *client.Client, org, repo, tag string) (OperatorIndexImage, error) {
	commit, err := github.GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return OperatorIndexImage{}, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	labels := append(getLabelsForPipelineRunsForTag(org, repo, commit.SHA),
		"tekton.dev/pipeline=operator-index-on-push",
	)
	plrs, err := c.GetKonflux().GetPipelineRuns(ctx, labels)
	if err != nil {
		return OperatorIndexImage{}, fmt.Errorf("failed to get operator-index-on-push pipeline runs: %w", err)
	}

	// Filter to runs for this tag and pick the most recent one (or alphabetically highest if same timestamp).
	var matched []kflx.PipelineRun
	for _, plr := range plrs {
		if plr.Metadata.Annotations["pipelinesascode.tekton.dev/branch"] == "refs/tags/"+tag {
			matched = append(matched, plr)
		}
	}
	if len(matched) == 0 {
		return OperatorIndexImage{}, fmt.Errorf("no operator-index-on-push pipeline run found for tag %s", tag)
	}
	if len(matched) > 1 {
		fmt.Fprintf(os.Stderr, "WARNING: Multiple operator-index-on-push pipeline runs found for tag %s. Returning results from the latest run.\n", tag)
	}

	sort.Slice(matched, func(i, j int) bool {
		ti, tj := matched[i].Metadata.CreationTimestamp, matched[j].Metadata.CreationTimestamp
		if !ti.Equal(*tj) {
			return ti.After(*tj)
		}
		return matched[i].Metadata.Name > matched[j].Metadata.Name
	})
	found := &matched[0]

	var imageURL, imageDigest string
	for _, r := range found.Status.Results {
		switch r.Name {
		case "IMAGE_URL":
			v, err := r.StringValue()
			if err != nil {
				return OperatorIndexImage{}, fmt.Errorf("invalid IMAGE_URL in pipeline run %s: %w", found.Metadata.Name, err)
			}
			imageURL = v
		case "IMAGE_DIGEST":
			v, err := r.StringValue()
			if err != nil {
				return OperatorIndexImage{}, fmt.Errorf("invalid IMAGE_DIGEST in pipeline run %s: %w", found.Metadata.Name, err)
			}
			imageDigest = v
		}
	}
	if imageURL == "" {
		return OperatorIndexImage{}, fmt.Errorf("IMAGE_URL result not found in operator-index-on-push pipeline run %s", found.Metadata.Name)
	}
	if imageDigest == "" {
		return OperatorIndexImage{}, fmt.Errorf("IMAGE_DIGEST result not found in operator-index-on-push pipeline run %s", found.Metadata.Name)
	}

	return OperatorIndexImage{
		ImageURL:    imageURL,
		ImageDigest: imageDigest,
	}, nil
}
