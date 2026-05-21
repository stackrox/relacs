package konflux

import (
	"encoding/json"
	"fmt"
	"strings"
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
	cmd, err := c.k8sCommand("get", "pipelinerun", "-l", strings.Join(labels, ","))
	if err != nil {
		return []PipelineRun{}, fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := runCommand(cmd)
	if err != nil {
		return []PipelineRun{}, fmt.Errorf("failed to retrieve pipeline run: %w", err)
	}

	pipelineRunList := PipelineRunList{}
	err = json.Unmarshal(output, &pipelineRunList)
	if err != nil {
		return []PipelineRun{}, fmt.Errorf("failed to unmarshal pipeline run list: %w", err)
	}
	if len(pipelineRunList.Items) == 0 {
		return []PipelineRun{}, fmt.Errorf("no pipeline runs found for the given labels")
	}
	if limit >= 0 && len(pipelineRunList.Items) > limit {
		return []PipelineRun{}, fmt.Errorf("more than %d pipeline runs found for the given labels", limit)
	}
	return pipelineRunList.Items, nil
}
