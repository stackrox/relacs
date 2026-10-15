package konflux

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	kflx "github.com/stackrox/relacs/pkg/client/konflux"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/konflux"
)

var promoteCdnReleaseToProdCmd = &cobra.Command{
	Use:   "promote-cdn-release-to-prod",
	Short: "Promote Konflux stage CDN releases to production",
	Long:  `Reads stage CDN release YAML file(s) and generates production release CRs. Verifies each stage release succeeded before promoting.`,
	RunE:  common.RunCommand(runPromoteCdnReleaseToProd),
}

func init() {
	promoteCdnReleaseToProdCmd.Flags().StringP("file", "f", "", "Stage CDN release YAML file (supports multi-doc YAML) (required)")
	promoteCdnReleaseToProdCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	promoteCdnReleaseToProdCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	promoteCdnReleaseToProdCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	promoteCdnReleaseToProdCmd.Flags().StringP("output", "o", "yaml", "Output format (text|json|yaml)")

	//nolint:errcheck
	promoteCdnReleaseToProdCmd.MarkFlagRequired("file")
}

func runPromoteCdnReleaseToProd(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	filePath, _ := cmd.Flags().GetString("file")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(fileContent))
	var stageReleases []kflx.CdnRelease

	for {
		var release kflx.CdnRelease
		err := decoder.Decode(&release)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		stageReleases = append(stageReleases, release)
	}

	if len(stageReleases) == 0 {
		return nil, fmt.Errorf("no releases found in file %s", filePath)
	}

	author, err := c.GetKonflux().GetCurrentUsername()
	if err != nil {
		return nil, fmt.Errorf("failed to get current username: %w", err)
	}

	prodReleases := make(common.MultiDocYAML, 0, len(stageReleases))
	for _, stageRelease := range stageReleases {
		// Fetch the release from Konflux to verify it exists and succeeded
		liveRelease, err := konflux.GetRelease(ctx, c, stageRelease.Metadata.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get stage release %s: %w", stageRelease.Metadata.Name, err)
		}

		if !liveRelease.IsSucceeded() {
			return nil, fmt.Errorf("stage release %s is not succeeded, cannot promote to prod", stageRelease.Metadata.Name)
		}

		prodRelease, err := konflux.PromoteCdnReleaseToProd(ctx, c, liveRelease, applicationPrefix, org, repo, author)
		if err != nil {
			return nil, fmt.Errorf("failed to promote CDN release %s to prod: %w", stageRelease.Metadata.Name, err)
		}

		prodReleases = append(prodReleases, prodRelease)
	}

	return prodReleases, nil
}
