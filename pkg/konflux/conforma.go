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

// ConformaRunInfo describes a conforma pipeline run that did not contribute
// to the violations: it is either still in progress or its report could not
// be retrieved or parsed (Reason).
type ConformaRunInfo struct {
	Name   string `json:"name" yaml:"name"`
	URL    string `json:"url" yaml:"url"`
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// ConformaViolations is the result of listing the conforma violations for a
// tag. The result is partial when SkippedRuns or InProgressRuns is non-empty.
type ConformaViolations struct {
	Violations     Violations        `json:"violations" yaml:"violations"`
	SkippedRuns    []ConformaRunInfo `json:"skippedRuns" yaml:"skippedRuns"`
	InProgressRuns []ConformaRunInfo `json:"inProgressRuns" yaml:"inProgressRuns"`
}

// ListConformaViolations returns the violations reported by the conforma
// pipeline runs for the given tag and scenario. Runs that are still in
// progress are listed in InProgressRuns, and runs whose report could not be
// retrieved (e.g. they failed before the report step ran) in SkippedRuns. It
// errors only when no report could be read at all because runs were skipped;
// while every run is still in progress, the partial result is returned.
func ListConformaViolations(ctx context.Context, c *client.Client, org, repo, tag, scenario string) (ConformaViolations, error) {
	pipelineRuns, err := conformaPipelineRunsForScenario(ctx, c, org, repo, tag, scenario)
	if err != nil {
		return ConformaViolations{}, fmt.Errorf("failed to get conforma pipeline runs: %w", err)
	}

	return violationsFromPipelineRuns(ctx, c, pipelineRuns, tag)
}

// violationsFromPipelineRuns lists the violations reported by the given
// conforma pipeline runs, reading the report of the most recent run per
// component and scenario.
func violationsFromPipelineRuns(ctx context.Context, c *client.Client, pipelineRuns []kflx.PipelineRun, tag string) (ConformaViolations, error) {
	conformaRuns := latestPipelineRunPerComponentAndScenario(pipelineRuns)
	if len(conformaRuns) == 0 {
		return ConformaViolations{}, fmt.Errorf("no conforma pipeline runs found for tag %s", tag)
	}

	result := ConformaViolations{}
	reportsRead := 0
	for _, conformaRun := range conformaRuns {
		if isPipelineRunInProgress(conformaRun.Status.Conditions) {
			result.InProgressRuns = append(result.InProgressRuns, ConformaRunInfo{
				Name: conformaRun.Metadata.Name,
				URL:  pipelineRunURL(conformaRun),
			})
			continue
		}

		podName := fmt.Sprintf("%s-verify-pod", conformaRun.Metadata.Name)
		report, err := c.GetKonflux().GetPodLogs(ctx, podName, "step-report-json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to get report for conforma run %s: %v\n", conformaRun.Metadata.Name, err)
			result.SkippedRuns = append(result.SkippedRuns, ConformaRunInfo{
				Name:   conformaRun.Metadata.Name,
				URL:    pipelineRunURL(conformaRun),
				Reason: err.Error(),
			})
			continue
		}

		runViolations, err := parseConformaViolations(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to parse conforma report for run %s: %v\n", conformaRun.Metadata.Name, err)
			result.SkippedRuns = append(result.SkippedRuns, ConformaRunInfo{
				Name:   conformaRun.Metadata.Name,
				URL:    pipelineRunURL(conformaRun),
				Reason: err.Error(),
			})
			continue
		}
		result.Violations.Violations = append(result.Violations.Violations, runViolations.Violations...)
		reportsRead++
	}
	if reportsRead == 0 && len(result.SkippedRuns) > 0 {
		return ConformaViolations{}, fmt.Errorf("failed to retrieve any conforma report for tag %s (%d run(s) still in progress)", tag, len(result.InProgressRuns))
	}
	// When every run is still in progress, the result is partial with the
	// runs listed in InProgressRuns rather than an error.
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
