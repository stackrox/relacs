package konflux

import (
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/client/konflux/kubearchive"
)

func (c *Client) GetPipelineRun(labels []string) (PipelineRun, error) {
	pipelineRuns, err := c._getPipelineRuns(labels, 1)
	if err != nil {
		return PipelineRun{}, fmt.Errorf("failed to get pipeline run: %w", err)
	}
	return pipelineRuns[0], nil
}

func (c *Client) GetPipelineRuns(labels []string) ([]PipelineRun, error) {
	return c._getPipelineRuns(labels, -1)
}

func (c *Client) _getPipelineRuns(labels []string, limit int) ([]PipelineRun, error) {
	opts := kubearchive.QueryOptions{
		LabelSelector: strings.Join(labels, ","),
	}
	pipelineRuns, err := kubearchive.QueryResources[PipelineRun](c.kubearchive, "tekton.dev", "v1", "pipelineruns", opts)
	if err != nil {
		return []PipelineRun{}, fmt.Errorf("failed to retrieve pipeline run: %w", err)
	}
	if len(pipelineRuns) == 0 {
		return []PipelineRun{}, fmt.Errorf("no pipeline runs found for the given labels")
	}
	if limit >= 0 && len(pipelineRuns) > limit {
		return []PipelineRun{}, fmt.Errorf("more than %d pipeline runs found for the given labels", limit)
	}
	return pipelineRuns, nil
}
