package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ViolationsOutput struct {
	Violations konflux.Violations `json:"violations" yaml:"violations"`
}

func (s ViolationsOutput) String() string {
	output := ""
	for _, violation := range s.Violations {
		output += fmt.Sprintf("Image:     %s\n", violation.Image)
		output += fmt.Sprintf("Violation: %s\n", violation.Violation)
		output += "\n"
	}
	return output
}

var getConformaViolationsCmd = &cobra.Command{
	Use:   "get-conforma-violations",
	Short: "Get Conforma violations",
	Long:  `Get the Conforma violations for a specific Git tag`,
	RunE:  runGetConformaViolations,
}

func init() {
	getConformaViolationsCmd.Flags().String("org", "", "GitHub organization (required)")
	getConformaViolationsCmd.Flags().String("repo", "", "GitHub repository (required)")
	getConformaViolationsCmd.Flags().String("tag", "", "Git tag (required)")
	getConformaViolationsCmd.Flags().String("base-application-name", "", "Base application name (required)")
	getConformaViolationsCmd.Flags().String("namespace", "", "Namespace (required)")
	//nolint:errcheck
	getConformaViolationsCmd.MarkFlagRequired("base-application-name")
	//nolint:errcheck
	getConformaViolationsCmd.MarkFlagRequired("org")
	//nolint:errcheck
	getConformaViolationsCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	getConformaViolationsCmd.MarkFlagRequired("tag")
	//nolint:errcheck
	getConformaViolationsCmd.MarkFlagRequired("namespace")
}

func runGetConformaViolations(cmd *cobra.Command, args []string) error {
	// Get flag values
	baseApplicationName, _ := cmd.Flags().GetString("base-application-name")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	namespace, _ := cmd.Flags().GetString("namespace")
	outputFormat, err := common.GetOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	// Get Conforma violations
	violations, err := konflux.GetConformaViolations(context.Background(), c, baseApplicationName, org, repo, tag, namespace)
	if err != nil {
		return fmt.Errorf("failed to get Conforma violations: %w", err)
	}
	// Format output
	return common.FormatOutput(outputFormat, ViolationsOutput{Violations: violations})
}
