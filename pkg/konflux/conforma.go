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

// ConformaViolations is the result of listing the conforma violations for a
// tag. The result is partial when SkippedRuns or InProgressRuns is non-empty.
type ConformaViolations struct {
	Violations     Violations `json:"violations" yaml:"violations"`
	SkippedRuns    []string   `json:"skippedRuns" yaml:"skippedRuns"`
	InProgressRuns []string   `json:"inProgressRuns" yaml:"inProgressRuns"`
}

// ListConformaViolations returns the violations reported by the conforma
// pipeline runs for the given tag. Runs that are still in progress are listed
// in InProgressRuns, and runs whose report could not be retrieved (e.g. they
// failed before the report step ran) in SkippedRuns.
func ListConformaViolations(ctx context.Context, c *client.Client, org, repo, tag string) (ConformaViolations, error) {
	pipelineRuns, err := getConformaPipelineRuns(ctx, c, org, repo, tag)
	if err != nil {
		return ConformaViolations{}, fmt.Errorf("failed to get conforma pipeline runs: %w", err)
	}

	conformaRuns := latestPipelineRunPerComponentAndScenario(pipelineRuns)
	if len(conformaRuns) == 0 {
		return ConformaViolations{}, fmt.Errorf("no conforma pipeline runs found for tag %s", tag)
	}

	result := ConformaViolations{}
	reportsRead := 0
	for _, conformaRun := range conformaRuns {
		if isPipelineRunInProgress(conformaRun.Status.Conditions) {
			result.InProgressRuns = append(result.InProgressRuns, conformaRun.Metadata.Name)
			continue
		}

		podName := fmt.Sprintf("%s-verify-pod", conformaRun.Metadata.Name)
		report, err := c.GetKonflux().GetPodLogs(ctx, podName, "step-report-json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to get report for conforma run %s: %v\n", conformaRun.Metadata.Name, err)
			result.SkippedRuns = append(result.SkippedRuns, conformaRun.Metadata.Name)
			continue
		}

		runViolations, err := parseConformaViolations(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to parse conforma report for run %s: %v\n", conformaRun.Metadata.Name, err)
			result.SkippedRuns = append(result.SkippedRuns, conformaRun.Metadata.Name)
			continue
		}
		result.Violations.Violations = append(result.Violations.Violations, runViolations.Violations...)
		reportsRead++
	}
	if reportsRead == 0 {
		if len(result.SkippedRuns) == 0 {
			return ConformaViolations{}, fmt.Errorf("conforma checks for tag %s are still in progress (%d run(s) running)", tag, len(result.InProgressRuns))
		}
		return ConformaViolations{}, fmt.Errorf("failed to retrieve any conforma report for tag %s (%d run(s) still in progress)", tag, len(result.InProgressRuns))
	}
	return result, nil
}

// isPipelineRunInProgress reports whether the pipeline run has not reached a
// final state yet: either it has no Succeeded condition, or the condition
// status is still Unknown (e.g. reason "Started" or "Running"). It matches on
// the condition status rather than the reason, which varies while a run is
// in progress.
func isPipelineRunInProgress(conditions []kflx.PipelineRunCondition) bool {
	for _, cond := range conditions {
		if cond.Type == "Succeeded" {
			return cond.Status == "Unknown"
		}
	}
	return true
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
