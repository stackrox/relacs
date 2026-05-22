package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type OperatorBundleImageOutput struct {
	ApplicationPrefix   string `json:"application-prefix" yaml:"application-prefix"`
	Organization        string `json:"organization" yaml:"organization"`
	Repository          string `json:"repository" yaml:"repository"`
	Tag                 string `json:"tag" yaml:"tag"`
	Snapshot            string `json:"snapshot" yaml:"snapshot"`
	OperatorBundleImage string `json:"operator-bundle-image" yaml:"operator-bundle-image"`
}

func (o OperatorBundleImageOutput) String() string {
	output := fmt.Sprintf("Input: %s/%s@%s\n", o.Organization, o.Repository, o.Tag)
	output += fmt.Sprintf("Application Prefix: %s\n", o.ApplicationPrefix)
	output += fmt.Sprintf("Snapshot: %s\n", o.Snapshot)
	output += fmt.Sprintf("Operator Bundle Image: %s\n", o.OperatorBundleImage)
	return output
}

var getOperatorBundleImageCmd = &cobra.Command{
	Use:   "get-operator-bundle-image",
	Short: "Get the operator bundle image for a given Git tag",
	Long:  `Retrieve the operator bundle image associated with a specific Git tag from the custom snapshot`,
	RunE:  common.RunCommand(runGetOperatorBundleImage),
}

func init() {
	getOperatorBundleImageCmd.Flags().String("application-prefix", "", "Application prefix (optional)")
	getOperatorBundleImageCmd.Flags().String("org", "", "GitHub organization (optional)")
	getOperatorBundleImageCmd.Flags().String("repo", "", "GitHub repository (optional)")
	getOperatorBundleImageCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getOperatorBundleImageCmd.MarkFlagRequired("tag")
}

func runGetOperatorBundleImage(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	cfg := c.GetConfig()

	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	if org == "" {
		org = cfg.GitHubOrganization
	}
	repo, _ := cmd.Flags().GetString("repo")
	if repo == "" {
		repo = cfg.MainGitHubRepository
	}
	tag, _ := cmd.Flags().GetString("tag")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")
	if applicationPrefix == "" {
		applicationPrefix = cfg.KonfluxApplicationPrefix
	}

	// Get the custom snapshot
	snapshot, err := konflux.GetCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	snapshotName := snapshot.Metadata.Name
	operatorBundleImage, err := snapshot.GetOperatorBundleImage(cfg, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator bundle image: %w", err)
	}

	return OperatorBundleImageOutput{
		ApplicationPrefix:   applicationPrefix,
		Organization:        org,
		Repository:          repo,
		Tag:                 tag,
		Snapshot:            snapshotName,
		OperatorBundleImage: operatorBundleImage,
	}, nil
}
