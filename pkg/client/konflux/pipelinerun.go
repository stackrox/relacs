package konflux

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c *Client) GetPipelineRun(labels []string) (PipelineRun, error) {
	cmd, err := c.k8sCommand("get", "pipelinerun", "-l", strings.Join(labels, ","))
	if err != nil {
		return PipelineRun{}, fmt.Errorf("failed to get k8s command: %w", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return PipelineRun{}, fmt.Errorf("failed to retrieve pipeline run: %w, %s", err, string(output))
	}

	pipelineRunList := PipelineRunList{}
	err = json.Unmarshal(output, &pipelineRunList)
	if err != nil {
		return PipelineRun{}, fmt.Errorf("failed to unmarshal pipeline run list: %w", err)
	}
	if len(pipelineRunList.Items) == 0 {
		return PipelineRun{}, fmt.Errorf("no pipeline runs found for the given labels")
	}
	if len(pipelineRunList.Items) > 1 {
		return PipelineRun{}, fmt.Errorf("multiple pipeline runs found for the given labels")
	}
	return pipelineRunList.Items[0], nil
}
