package konflux

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ConformaViolationsOutput struct {
	Organization   string                    `json:"organization" yaml:"organization"`
	Repository     string                    `json:"repository" yaml:"repository"`
	Tag            string                    `json:"tag" yaml:"tag"`
	Violations     konflux.Violations        `json:"violations" yaml:"violations"`
	SkippedRuns    []konflux.ConformaRunInfo `json:"skippedRuns" yaml:"skippedRuns"`
	InProgressRuns []konflux.ConformaRunInfo `json:"inProgressRuns" yaml:"inProgressRuns"`
}

func (c ConformaViolationsOutput) String() string {
	output := fmt.Sprintf("Input: %s/%s@%s\n", c.Organization, c.Repository, c.Tag)
	output += fmt.Sprintln("Violations:")
	if len(c.Violations.Violations) == 0 {
		output += "  - No violations found\n"
	} else {
		for _, violation := range c.Violations.Violations {
			output += fmt.Sprintf("  - Image: %s\n", violation.Image)
			output += fmt.Sprintf("    Violation: %s\n", violation.Violation)
		}
	}
	if len(c.InProgressRuns) > 0 {
		output += fmt.Sprintln("In-progress runs:")
		for _, run := range c.InProgressRuns {
			output += fmt.Sprintf("  - %s (%s)\n", run.Name, run.URL)
		}
	}
	if len(c.SkippedRuns) > 0 {
		output += fmt.Sprintln("Skipped runs (no report available):")
		for _, run := range c.SkippedRuns {
			output += fmt.Sprintf("  - %s: %s\n", run.Name, run.Reason)
		}
	}
	return output
}

var listConformaViolationsCmd = &cobra.Command{
	Use:   "list-conforma-violations",
	Short: "List the Conforma violations for a given Git tag",
	Long: `Retrieve the Conforma violations reported by the conforma pipeline runs for a given Git tag.

By default (scenario acs-conforma-dev), only conforma runs triggered
directly on the tag's commit (via pipelines-as-code) are considered.
With --scenario acs-conforma-prod-like (only supported for the
stackrox repository), the violations of the conforma runs attached
to the tag's custom snapshot are listed instead.`,
	RunE: common.RunCommand(runListConformaViolations),
}

func init() {
	listConformaViolationsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listConformaViolationsCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	listConformaViolationsCmd.Flags().String("tag", "", "Git tag (required)")
	listConformaViolationsCmd.Flags().String("scenario", konflux.ConformaScenarioDev, "Conforma scenario: "+konflux.ConformaScenarioDev+" or "+konflux.ConformaScenarioProdLike+" (the latter is only supported for the stackrox repository)")

	//nolint:errcheck
	listConformaViolationsCmd.MarkFlagRequired("tag")
}

func runListConformaViolations(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	scenario, _ := cmd.Flags().GetString("scenario")

	result, err := konflux.ListConformaViolations(ctx, c, org, repo, tag, scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to list conforma violations: %w", err)
	}

	output := ConformaViolationsOutput{
		Organization:   org,
		Repository:     repo,
		Tag:            tag,
		Violations:     result.Violations,
		SkippedRuns:    result.SkippedRuns,
		InProgressRuns: result.InProgressRuns,
	}
	if len(result.SkippedRuns) > 0 || len(result.InProgressRuns) > 0 {
		// The result is partial: still render the output, but exit non-zero.
		var reasons []string
		if len(result.SkippedRuns) > 0 {
			reasons = append(reasons, fmt.Sprintf("no conforma report available for %d run(s): %s", len(result.SkippedRuns), runNames(result.SkippedRuns)))
		}
		if len(result.InProgressRuns) > 0 {
			reasons = append(reasons, fmt.Sprintf("%d conforma run(s) still in progress: %s", len(result.InProgressRuns), runNames(result.InProgressRuns)))
		}
		return output, &common.PartialResultError{
			Err: fmt.Errorf("%s", strings.Join(reasons, "; ")),
		}
	}
	return output, nil
}

func runNames(runs []konflux.ConformaRunInfo) string {
	names := make([]string, len(runs))
	for i, run := range runs {
		names[i] = run.Name
	}
	return strings.Join(names, ", ")
}
