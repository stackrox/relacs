package konflux

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
)

type ConformaReport struct {
	Components []ConformaComponent `json:"components"`
}

type ConformaComponent struct {
	Name       string              `json:"name"`
	Violations []ConformaViolation `json:"violations"`
}

type ConformaViolation struct {
	Message string `json:"msg"`
}

type Violations struct {
	Violations []Violation `json:"violations"`
}

type Violation struct {
	Image     string `json:"image"`
	Violation string `json:"violation"`
}

func ListConformaViolations(ctx context.Context, c *client.Client, org, repo, tag string) (Violations, error) {
	pipelineRuns, err := getConformaPipelineRuns(ctx, c, org, repo, tag)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to get conforma pipeline runs: %w", err)
	}

	conformaRuns := latestPipelineRunPerComponentAndScenario(pipelineRuns)
	if len(conformaRuns) == 0 {
		return Violations{}, fmt.Errorf("no conforma pipeline runs found for tag %s", tag)
	}

	violations := Violations{}
	reportsRead := 0
	for _, conformaRun := range conformaRuns {
		podName := fmt.Sprintf("%s-verify-pod", conformaRun.Metadata.Name)
		report, err := c.GetKonflux().GetPodLogs(ctx, podName, "step-report-json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to get report for conforma run %s: %v\n", conformaRun.Metadata.Name, err)
			continue
		}

		runViolations, err := parseConformaViolations(report)
		if err != nil {
			return Violations{}, fmt.Errorf("failed to parse conforma violations for run %s: %w", conformaRun.Metadata.Name, err)
		}
		violations.Violations = append(violations.Violations, runViolations.Violations...)
		reportsRead++
	}
	if reportsRead == 0 {
		return Violations{}, fmt.Errorf("failed to retrieve any conforma report for tag %s", tag)
	}
	return violations, nil
}

// latestPipelineRunPerComponentAndScenario deduplicates conforma pipeline runs
// by their component and scenario labels, keeping only the most recent run for
// each combination. This ignores violations from older re-runs of the same
// component and scenario.
func latestPipelineRunPerComponentAndScenario(pipelineRuns []kflx.PipelineRun) []kflx.PipelineRun {
	latest := make(map[string]kflx.PipelineRun)
	for _, plr := range pipelineRuns {
		key := plr.Metadata.Labels["appstudio.openshift.io/component"] + "," + plr.Metadata.Labels["test.appstudio.openshift.io/scenario"]
		current, ok := latest[key]
		if !ok || plr.Metadata.CreationTimestamp.After(*current.Metadata.CreationTimestamp) {
			latest[key] = plr
		}
	}

	runs := make([]kflx.PipelineRun, 0, len(latest))
	for _, plr := range latest {
		runs = append(runs, plr)
	}
	return runs
}

func parseConformaViolations(report string) (Violations, error) {
	conformaReport := ConformaReport{}
	violations := Violations{}

	err := json.Unmarshal([]byte(report), &conformaReport)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to unmarshal conforma report: %w", err)
	}

	// Components in this context refer to both component and multi-arch images
	for _, component := range conformaReport.Components {
		for _, violation := range component.Violations {
			violations.Violations = append(violations.Violations, Violation{
				Image:     component.Name,
				Violation: violation.Message,
			})
		}
	}
	return violations, nil
}
