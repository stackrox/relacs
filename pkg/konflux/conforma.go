package konflux

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/utils"
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

func ListConformaViolations(ctx context.Context, c *client.Client, applicationPrefix, org, repo, tag string) (Violations, error) {
	snapshot, err := GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	conformaRun, err := getConformaRunForSnapshot(c, snapshot)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to get conforma run for snapshot: %w", err)
	}

	podName := fmt.Sprintf("%s-verify-pod", conformaRun.Metadata.Name)
	report, err := c.GetKonflux().GetPodLogs(podName, "step-report-json")
	if err != nil {
		return Violations{}, fmt.Errorf("failed to get pod logs: %w", err)
	}

	violations, err := parseConformaViolations(report)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to parse conforma violations: %w", err)
	}
	return violations, nil
}

func getConformaRunForSnapshot(c *client.Client, snapshot kflx.Snapshot) (kflx.PipelineRun, error) {
	cfg := c.GetConfig()
	application, ok := snapshot.Metadata.Labels["appstudio.openshift.io/application"]
	if !ok {
		return kflx.PipelineRun{}, fmt.Errorf("application not found in snapshot labels")
	}
	conformaScenario, err := utils.Application{Name: application}.GetResourceWithVersion(cfg, cfg.ConformaDevScenarioPrefix)
	if err != nil {
		return kflx.PipelineRun{}, fmt.Errorf("failed to get conforma scenario name: %w", err)
	}
	labels := []string{
		"appstudio.openshift.io/snapshot=" + snapshot.Metadata.Name,
		"test.appstudio.openshift.io/scenario=" + conformaScenario,
	}
	pipelineRun, err := c.GetKonflux().GetPipelineRun(labels)
	if err != nil {
		return kflx.PipelineRun{}, fmt.Errorf("failed to get pipeline run: %w", err)
	}
	return pipelineRun, nil
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
