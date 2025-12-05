package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ConformaViolationsOutput struct {
	ApplicationPrefix string             `json:"application-prefix" yaml:"application-prefix"`
	Organization      string             `json:"organization" yaml:"organization"`
	Repository        string             `json:"repository" yaml:"repository"`
	Tag               string             `json:"tag" yaml:"tag"`
	Violations        konflux.Violations `json:"violations" yaml:"violations"`
}

func (c ConformaViolationsOutput) String() string {
	output := fmt.Sprintf("Input: %s/%s@%s\n", c.Organization, c.Repository, c.Tag)
	output += fmt.Sprintf("Application Prefix: %s\n", c.ApplicationPrefix)
	output += fmt.Sprintln("Violations:")
	for _, violation := range c.Violations.Violations {
		output += fmt.Sprintf("  - Image: %s\n", violation.Image)
		output += fmt.Sprintf("    Violation: %s\n", violation.Violation)
	}
	return output
}

var getConformaViolationsCmd = &cobra.Command{
	Use:   "get-conforma-violations",
	Short: "Get the Conforma violations for a given Git tag",
	Long:  `Retrieve the Conforma violations associated with a specific Git tag`,
	RunE:  common.RunCommand(runGetConformaViolations),
}

func init() {
	getConformaViolationsCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	getConformaViolationsCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	getConformaViolationsCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	getConformaViolationsCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getConformaViolationsCmd.MarkFlagRequired("tag")
}

func runGetConformaViolations(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")

	violations, err := konflux.GetConformaViolations(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get conforma violations: %w", err)
	}

	return ConformaViolationsOutput{
		ApplicationPrefix: applicationPrefix,
		Organization:      org,
		Repository:        repo,
		Tag:               tag,
		Violations:        violations,
	}, nil
}
