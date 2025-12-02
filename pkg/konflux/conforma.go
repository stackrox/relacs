package konflux

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/utils"
)

type Violations []Violation

type Violation struct {
	Image     string `json:"image"`
	Violation string `json:"violation"`
}

func GetConformaViolations(ctx context.Context, c *client.Client, applicationName, org, repo, tag, namespace string) (Violations, error) {
	baseApplicationName := utils.GetApplicationVersionName(applicationName, tag)
	version, err := utils.GetVersionFromApplicationName(baseApplicationName)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to get version from application name: %w", err)
	}

	snapshot, _, err := FindCustomSnapshotForTag(ctx, c, applicationName, org, repo, tag)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to find custom snapshot for tag: %w", err)
	}

	pipelineRunName, err := findConformaRunForSnapshot(snapshot, version)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to find conforma run for snapshot: %w", err)
	}

	podName := fmt.Sprintf("%s-verify-pod", pipelineRunName)
	logs, err := fetchPodLogs(namespace, podName, "step-report-json")
	if err != nil {
		return Violations{}, fmt.Errorf("failed to fetch pipeline run logs: %w", err)
	}

	violations, err := extractViolationsFromLogs(logs)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to extract violations from logs: %w", err)
	}

	return violations, nil
}

func extractViolationsFromLogs(logs string) (Violations, error) {
	var report ConformaReport
	err := json.Unmarshal([]byte(logs), &report)
	if err != nil {
		return Violations{}, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	var violationsOutput Violations
	components := report.Components
	for _, component := range components {
		for _, violation := range component.Violations {
			violationsOutput = append(violationsOutput, Violation{
				Image:     component.Name,
				Violation: violation.Message,
			})
		}
	}
	return violationsOutput, nil
}

func getConformaScenario(version utils.Version) string {
	return fmt.Sprintf("acs-conforma-dev-%d-%d", version.Major, version.Minor)
}

func findConformaRunForSnapshot(snapshot string, version utils.Version) (string, error) {
	labels := []string{
		"appstudio.openshift.io/snapshot=" + snapshot,
		"test.appstudio.openshift.io/scenario=" + getConformaScenario(version),
	}
	query := strings.Join(labels, ",")
	cmd, err := getK8sCommand("ka", "get", "pipelinerun", "-l", query, "-o", "yaml")
	if err != nil {
		return "", fmt.Errorf("failed to get pipelineruns for snapshot: %w", err)
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pipelineruns for snapshot: %w", err)
	}

	pipelineRun, err := parsePipelineRunFromPipelineRunList(output)
	if err != nil {
		return "", fmt.Errorf("failed to parse pipelinerun: %w", err)
	}

	return pipelineRun.Metadata.Name, nil
}
