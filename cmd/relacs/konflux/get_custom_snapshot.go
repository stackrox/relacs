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

type SnapshotOutput struct {
	ApplicationPrefix string `json:"application-prefix" yaml:"application-prefix"`
	Organization      string `json:"organization" yaml:"organization"`
	Repository        string `json:"repository" yaml:"repository"`
	Tag               string `json:"tag" yaml:"tag"`
	Snapshot          string `json:"snapshot" yaml:"snapshot"`
	URL               string `json:"url" yaml:"url"`
}

func (s SnapshotOutput) String() string {
	output := fmt.Sprintf("Input: %s/%s@%s\n", s.Organization, s.Repository, s.Tag)
	output += fmt.Sprintf("Application Prefix: %s\n", s.ApplicationPrefix)
	output += fmt.Sprintf("Snapshot: %s\n", s.Snapshot)
	output += fmt.Sprintf("URL: %s\n", s.URL)
	return output
}

var getCustomSnapshotCmd = &cobra.Command{
	Use:   "get-custom-snapshot",
	Short: "Get a custom snapshot for a given Git tag",
	Long:  `Retrieve the custom snapshot associated with a specific Git tag`,
	RunE:  common.RunCommand(runGetCustomSnapshot),
}

func init() {
	getCustomSnapshotCmd.Flags().String("application-prefix", config.KonfluxApplicationPrefix, "Application prefix (optional)")
	getCustomSnapshotCmd.Flags().String("org", config.GitHubOrganization, "GitHub organization (optional)")
	getCustomSnapshotCmd.Flags().String("repo", config.MainGitHubRepository, "GitHub repository (optional)")
	getCustomSnapshotCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getCustomSnapshotCmd.MarkFlagRequired("tag")
}

func runGetCustomSnapshot(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")
	applicationPrefix, _ := cmd.Flags().GetString("application-prefix")

	snapshot, err := konflux.FindCustomSnapshot(ctx, c, applicationPrefix, org, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom snapshot: %w", err)
	}

	snapshotName := snapshot.Metadata.Name
	url := snapshot.GetURL(applicationPrefix, tag)
	return SnapshotOutput{
		ApplicationPrefix: applicationPrefix,
		Organization:      org,
		Repository:        repo,
		Tag:               tag,
		Snapshot:          snapshotName,
		URL:               url,
	}, nil
}
