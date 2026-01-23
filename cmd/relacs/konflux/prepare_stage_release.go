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

var prepareStageReleaseCmd = &cobra.Command{
	Use:   "prepare-stage-release",
	Short: "Prepare a given Konflux stage release",
	Long:  `Prepare a custom resource YAML for a given Konflux stage release from a snapshot`,
	RunE:  common.RunCommand(runPrepareStageRelease),
}

func init() {
	prepareStageReleaseCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	prepareStageReleaseCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	prepareStageReleaseCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	prepareStageReleaseCmd.Flags().String("tag", "", "Git tag (required)")
	prepareStageReleaseCmd.Flags().StringSlice("issues", []string{}, "Comma-separated list of issues (optional)")
	// TODO: the CVEs are a list of CVE mapping (CVE -> component) that are resolved with this advisory as identified above
	prepareStageReleaseCmd.Flags().StringSlice("cves", []string{}, "Comma-separated list of CVE:component pairs (optional)")

	//nolint:errcheck
	prepareStageReleaseCmd.MarkFlagRequired("tag")
}

func runPrepareStageRelease(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	issues, _ := cmd.Flags().GetStringSlice("issues")
	cves, err := cmd.Flags().GetStringSlice("cves")
	if err != nil {
		return nil, fmt.Errorf("failed to parse CVEs: %w", err)
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}
	release, err := konflux.RenderStageRelease(snapshot, issues, cves)
	if err != nil {
		return nil, fmt.Errorf("failed to render release: %w", err)
	}

	return release, nil
}
