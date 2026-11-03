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

var prepareStageCdnReleaseCmd = &cobra.Command{
	Use:   "prepare-stage-cdn-release",
	Short: "Prepare a Release CR for CDN binary publishing to stage",
	Long:  `Prepare a custom resource YAML for a CDN push release from a snapshot. Generates one Release CR per architecture by default; pass --arch to target a single architecture (also useful for re-run attempts with --attempt).`,
	RunE:  common.RunCommand(runPrepareStageCdnRelease),
}

func init() {
	prepareStageCdnReleaseCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	prepareStageCdnReleaseCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	prepareStageCdnReleaseCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	prepareStageCdnReleaseCmd.Flags().String("tag", "", "Git tag (required)")
	prepareStageCdnReleaseCmd.Flags().String("arch", "", fmt.Sprintf("Target architecture (optional, one of: %s); omit to generate for all architectures", strings.Join(config.GetCDNArchitectures(), ", ")))
	prepareStageCdnReleaseCmd.Flags().Int("attempt", 1, "Attempt number (optional, use with --arch to re-run a specific architecture)")
	prepareStageCdnReleaseCmd.Flags().StringP("output", "o", "yaml", "Output format (text|json|yaml)")

	//nolint:errcheck
	prepareStageCdnReleaseCmd.MarkFlagRequired("tag")
}

func runPrepareStageCdnRelease(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	arch, _ := cmd.Flags().GetString("arch")
	attempt, _ := cmd.Flags().GetInt("attempt")

	arches := config.GetCDNArchitectures()
	if arch != "" {
		if !konflux.IsValidCDNArch(arch) {
			return nil, fmt.Errorf("invalid arch %q, must be one of: %s", arch, strings.Join(config.GetCDNArchitectures(), ", "))
		}
		arches = []string{arch}
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	author, err := c.GetKonflux().GetCurrentUsername(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current username: %w", err)
	}

	docs := make(common.MultiDocYAML, 0, len(arches))
	for _, a := range arches {
		release, err := konflux.RenderCdnRelease(snapshot, a, "stage", attempt, author)
		if err != nil {
			return nil, fmt.Errorf("failed to render CDN release for arch %s: %w", a, err)
		}
		docs = append(docs, release)
	}
	return docs, nil
}
