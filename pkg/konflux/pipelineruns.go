package konflux

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func parsePipelineRunFromPipelineRunList(output []byte) (PipelineRun, error) {
	var pipelineRuns PipelineRunList
	err := yaml.Unmarshal(output, &pipelineRuns)
	if err != nil {
		return PipelineRun{}, fmt.Errorf("failed to unmarshal pipeline runs: %w", err)
	}

	if len(pipelineRuns.Items) == 0 {
		return PipelineRun{}, fmt.Errorf("no pipeline runs found")
	}
	if len(pipelineRuns.Items) > 1 {
		return PipelineRun{}, fmt.Errorf("multiple pipeline runs found")
	}

	return pipelineRuns.Items[0], nil
}
