package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
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

	//nolint:errcheck
	promoteReleaseToProdCmd.MarkFlagRequired("name")
}

func runPromoteReleaseToProd(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	name, _ := cmd.Flags().GetString("name")

	release, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	// Ensure the release is succeeded.
	if !release.IsSucceeded() {
		return nil, fmt.Errorf("release %s is not succeeded, cannot promote to prod.", name)
	}

	err = konflux.PromoteReleaseToProd(&release)
	if err != nil {
		return nil, fmt.Errorf("failed to promote release to prod: %w", err)
	}

	return release, nil
}
