package konflux

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

var prepareRoxctlReleaseCmd = &cobra.Command{
	Use:   "prepare-roxctl-release",
	Short: "Prepare a Release CR for roxctl CDN binary publishing",
	Long:  `Prepare a custom resource YAML for a roxctl CDN push release from a snapshot`,
	RunE:  common.RunCommand(runPrepareRoxctlRelease),
}

var validRoxctlArches = []string{"x86-64", "s390x", "ppc64le"}

func init() {
	prepareRoxctlReleaseCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	prepareRoxctlReleaseCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	prepareRoxctlReleaseCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	prepareRoxctlReleaseCmd.Flags().String("tag", "", "Git tag (required)")
	prepareRoxctlReleaseCmd.Flags().String("arch", "", fmt.Sprintf("Target architecture (optional, one of: %s); omit to generate for all architectures", strings.Join(validRoxctlArches, ", ")))
	prepareRoxctlReleaseCmd.Flags().Int("attempt", 1, "Attempt number (optional)")
	prepareRoxctlReleaseCmd.Flags().StringP("output", "o", "yaml", "Output format (text|json|yaml)")

	//nolint:errcheck
	prepareRoxctlReleaseCmd.MarkFlagRequired("tag")
}

func runPrepareRoxctlRelease(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	arch, _ := cmd.Flags().GetString("arch")
	attempt, _ := cmd.Flags().GetInt("attempt")

	arches := validRoxctlArches
	if arch != "" {
		if !isValidRoxctlArch(arch) {
			return nil, fmt.Errorf("invalid arch %q, must be one of: %s", arch, strings.Join(validRoxctlArches, ", "))
		}
		arches = []string{arch}
	}

	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	author, err := c.GetKonflux().GetCurrentUsername()
	if err != nil {
		return nil, fmt.Errorf("failed to get current username: %w", err)
	}

	if len(arches) == 1 {
		release, err := konflux.RenderRoxctlRelease(snapshot, arches[0], "stage", attempt, author)
		if err != nil {
			return nil, fmt.Errorf("failed to render roxctl release: %w", err)
		}
		return release, nil
	}

	docs := make(common.MultiDocYAML, 0, len(arches))
	for _, a := range arches {
		release, err := konflux.RenderRoxctlRelease(snapshot, a, "stage", attempt, author)
		if err != nil {
			return nil, fmt.Errorf("failed to render roxctl release for arch %s: %w", a, err)
		}
		docs = append(docs, release)
	}
	return docs, nil
}

func isValidRoxctlArch(arch string) bool {
	return slices.Contains(validRoxctlArches, arch)
}
