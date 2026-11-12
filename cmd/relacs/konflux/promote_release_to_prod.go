package konflux

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

var promoteReleaseToProdCmd = &cobra.Command{
	Use:   "promote-release-to-prod",
	Short: "Promote a given Konflux stage release to production",
	Long:  `Promote a custom resource YAML for a given Konflux stage release to production`,
	RunE:  common.RunCommand(runPromoteReleaseToProd),
}

func init() {
	promoteReleaseToProdCmd.Flags().String("name", "", "Konflux stage release name (required)")
	promoteReleaseToProdCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	promoteReleaseToProdCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	promoteReleaseToProdCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")

	// Default to YAML output for this command since it generates a Kubernetes CR
	promoteReleaseToProdCmd.Flags().StringP("output", "o", "yaml", "Output format (text|json|yaml)")

	//nolint:errcheck
	promoteReleaseToProdCmd.MarkFlagRequired("name")
}

func runPromoteReleaseToProd(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	name, _ := cmd.Flags().GetString("name")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	stageRelease, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage release: %w", err)
	}

	tag, err := konflux.GetValidTagFromStageRelease(stageRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid tag from stage release: %w", err)
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	// Ensure the release or one of its reruns succeeded, and get the successful release
	stageRelease, err = konflux.EnsureSuccessfulRelease(ctx, c, stageRelease, name)
	if err != nil {
		return nil, err
	}

	author, err := c.GetKonflux().GetCurrentUsername(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current username: %w", err)
	}

	prodRelease, err := konflux.PromoteReleaseToProd(stageRelease, snapshot, author)
	if err != nil {
		return nil, fmt.Errorf("failed to promote stagerelease to prod: %w", err)
	}

	return prodRelease, nil
}
