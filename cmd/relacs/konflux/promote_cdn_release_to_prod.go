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

var promoteCdnReleaseToProdCmd = &cobra.Command{
	Use:   "promote-cdn-release-to-prod",
	Short: "Promote a given Konflux stage CDN release to production",
	Long:  `Prepare a custom resource YAML for a given Konflux stage CDN release promoted to production`,
	RunE:  common.RunCommand(runPromoteCdnReleaseToProd),
}

func init() {
	promoteCdnReleaseToProdCmd.Flags().String("name", "", "Konflux stage CDN release name (required)")
	promoteCdnReleaseToProdCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	promoteCdnReleaseToProdCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	promoteCdnReleaseToProdCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	promoteCdnReleaseToProdCmd.Flags().StringP("output", "o", "yaml", "Output format (text|json|yaml)")

	//nolint:errcheck
	promoteCdnReleaseToProdCmd.MarkFlagRequired("name")
}

func runPromoteCdnReleaseToProd(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	name, _ := cmd.Flags().GetString("name")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	stageRelease, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage release: %w", err)
	}

	if !stageRelease.IsSucceeded() {
		return nil, fmt.Errorf("stage release %s is not succeeded, cannot promote to prod", name)
	}

	author, err := c.GetKonflux().GetCurrentUsername()
	if err != nil {
		return nil, fmt.Errorf("failed to get current username: %w", err)
	}

	prodRelease, err := konflux.PromoteCdnReleaseToProd(ctx, c, stageRelease, applicationPrefix, org, repo, author)
	if err != nil {
		return nil, fmt.Errorf("failed to promote CDN release to prod: %w", err)
	}

	return prodRelease, nil
}
