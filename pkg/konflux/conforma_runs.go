package konflux

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/github"
	"github.com/stackrox/relacs/pkg/utils"
)

type ConformaRunEntry struct {
	Name      string `json:"name" yaml:"name"`
	Component string `json:"component" yaml:"component"`
	Scenario  string `json:"scenario" yaml:"scenario"`
	Status    string `json:"status" yaml:"status"`
	URL       string `json:"url" yaml:"url"`
}

func ListConformaRuns(ctx context.Context, c *client.Client, org, repo, tag string) ([]ConformaRunEntry, error) {
	pipelineRuns, err := getConformaPipelineRuns(ctx, c, org, repo, tag)
	if err != nil {
		return nil, err
	}

	var entries []ConformaRunEntry
	for _, plr := range pipelineRuns {
		entry := convertToConformaRunEntry(plr)
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

func getConformaPipelineRuns(ctx context.Context, c *client.Client, org, repo, tag string) ([]kflx.PipelineRun, error) {
	commitObj, err := github.GetCommitForTag(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit for tag: %w", err)
	}

	labels := []string{
		"pac.test.appstudio.openshift.io/sha=" + commitObj.SHA,
		"pac.test.appstudio.openshift.io/url-org=" + org,
		"pac.test.appstudio.openshift.io/url-repository=" + repo,
		"pipelines.appstudio.openshift.io/type=test",
		"build.appstudio.redhat.com/pipeline=enterprise-contract",
	}
	pipelineRuns, err := c.GetKonflux().GetPipelineRuns(ctx, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline runs: %w", err)
	}

	var conformaRuns []kflx.PipelineRun
	for _, plr := range pipelineRuns {
		if plr.Metadata.Annotations["pac.test.appstudio.openshift.io/branch"] != "refs/tags/"+tag {
			continue
		}
		scenario := plr.Metadata.Labels["test.appstudio.openshift.io/scenario"]
		if !strings.Contains(scenario, "conforma") {
			continue
		}

		conformaRuns = append(conformaRuns, plr)
	}

	return conformaRuns, nil
}

func convertToConformaRunEntry(plr kflx.PipelineRun) ConformaRunEntry {
	appName := plr.Metadata.Labels["appstudio.openshift.io/application"]
	component := plr.Metadata.Labels["appstudio.openshift.io/component"]
	scenario := plr.Metadata.Labels["test.appstudio.openshift.io/scenario"]

	plrURL := utils.Application{Name: appName}.GetPipelineRunURL(plr.Metadata.Name)

	return ConformaRunEntry{
		Name:      plr.Metadata.Name,
		Component: component,
		Scenario:  scenario,
		Status:    resolveStatus(plr.Status.Conditions),
		URL:       plrURL,
	}
}

func resolveStatus(conditions []kflx.PipelineRunCondition) string {
	for _, cond := range conditions {
		if cond.Type == "Succeeded" {
			if cond.Status == "True" {
				return "Succeeded"
			}
			if cond.Reason != "" {
				return cond.Reason
			}
			return "Failed"
		}
	}
	return "Running"
}
