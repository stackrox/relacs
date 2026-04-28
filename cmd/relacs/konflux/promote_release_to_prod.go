package konflux

import (
	"context"
	"fmt"
	"regexp"

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
	promoteReleaseToProdCmd.Flags().String("tag", "", "Final release tag (required)")
	promoteReleaseToProdCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	promoteReleaseToProdCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	promoteReleaseToProdCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")

	//nolint:errcheck
	promoteReleaseToProdCmd.MarkFlagRequired("name")
	//nolint:errcheck
	promoteReleaseToProdCmd.MarkFlagRequired("tag")
}

// validateTagFormat validates that the tag is in strict x.y.z format.
func validateTagFormat(tag string) error {
	tagPattern := regexp.MustCompile(`^[1-9]\d*\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$`)
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("tag must be in x.y.z format (e.g., 1.2.3), got: %s", tag)
	}
	return nil
}

func runPromoteReleaseToProd(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	name, _ := cmd.Flags().GetString("name")
	tag, _ := cmd.Flags().GetString("tag")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	// Validate tag format is strictly x.y.z
	if err := validateTagFormat(tag); err != nil {
		return nil, err
	}

	stageRelease, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage release: %w", err)
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	// Ensure the release is succeeded.
	if !stageRelease.IsSucceeded() {
		return nil, fmt.Errorf("stage release %s is not succeeded, cannot promote to prod", name)
	}

	prodRelease, err := konflux.PromoteReleaseToProd(stageRelease, snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to promote stagerelease to prod: %w", err)
	}

	return prodRelease, nil
}
