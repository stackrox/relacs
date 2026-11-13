package konflux

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ConformaViolationsOutput struct {
	Organization string             `json:"organization" yaml:"organization"`
	Repository   string             `json:"repository" yaml:"repository"`
	Tag          string             `json:"tag" yaml:"tag"`
	Violations   konflux.Violations `json:"violations" yaml:"violations"`
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
	return output
}

var listConformaViolationsCmd = &cobra.Command{
	Use:   "list-conforma-violations",
	Short: "List the Conforma violations for a given Git tag",
	Long:  `Retrieve the Conforma violations reported by the conforma pipeline runs for a given Git tag`,
	RunE:  common.RunCommand(runListConformaViolations),
}

func init() {
	listConformaViolationsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	listConformaViolationsCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	listConformaViolationsCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	listConformaViolationsCmd.MarkFlagRequired("tag")
}

func runListConformaViolations(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")

	violations, err := konflux.ListConformaViolations(ctx, c, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to list conforma violations: %w", err)
	}

	return ConformaViolationsOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		Violations:   violations,
	}, nil
}
