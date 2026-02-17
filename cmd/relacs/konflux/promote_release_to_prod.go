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

var promoteReleaseToProdCmd = &cobra.Command{
	Use:   "promote-release-to-prod",
	Short: "Promote a given Konflux stage release to production",
	Long:  `Promote a custom resource YAML for a given Konflux stage release to production`,
	RunE:  common.RunCommand(runPromoteReleaseToProd),
}

func init() {
	promoteReleaseToProdCmd.Flags().String("name", "", "Konflux release name (required)")
	promoteReleaseToProdCmd.Flags().String("tag", "", "Git tag (required)")
	promoteReleaseToProdCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	promoteReleaseToProdCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	promoteReleaseToProdCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")

	//nolint:errcheck
	promoteReleaseToProdCmd.MarkFlagRequired("name")
	//nolint:errcheck
	promoteReleaseToProdCmd.MarkFlagRequired("tag")
}

func runPromoteReleaseToProd(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	name, _ := cmd.Flags().GetString("name")
	tag, _ := cmd.Flags().GetString("tag")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	release, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	// Ensure the release is succeeded.
	if !release.IsSucceeded() {
		return nil, fmt.Errorf("release %s is not succeeded, cannot promote to prod", name)
	}

	prodRelease, err := konflux.PromoteReleaseToProd(release, snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to promote release to prod: %w", err)
	}

	return prodRelease, nil
}
