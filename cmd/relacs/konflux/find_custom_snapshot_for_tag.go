package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type SnapshotOutput struct {
	Organization string `json:"organization" yaml:"organization"`
	Repository   string `json:"repository" yaml:"repository"`
	Tag          string `json:"tag" yaml:"tag"`
	Snapshot     string `json:"snapshot" yaml:"snapshot"`
	URL          string `json:"url" yaml:"url"`
}

func (s SnapshotOutput) String() string {
	output := ""
	output += fmt.Sprintf("Name: %s\n", s.Snapshot)
	output += fmt.Sprintf("URL:  %s\n", s.URL)
	return output
}

var findCustomSnapshotForTagCmd = &cobra.Command{
	Use:   "find-custom-snapshot-for-tag",
	Short: "Find Konflux custom snapshot for a Git tag",
	Long:  `Find the Konflux custom snapshot associated with a specific Git tag`,
	RunE:  runFindCustomSnapshotForTag,
}

func init() {
	findCustomSnapshotForTagCmd.Flags().String("org", "", "GitHub organization (required)")
	findCustomSnapshotForTagCmd.Flags().String("repo", "", "GitHub repository (required)")
	findCustomSnapshotForTagCmd.Flags().String("tag", "", "Git tag (required)")
	findCustomSnapshotForTagCmd.Flags().String("base-application-name", "", "Base application name (required)")

	//nolint:errcheck
	findCustomSnapshotForTagCmd.MarkFlagRequired("base-application-name")
	//nolint:errcheck
	findCustomSnapshotForTagCmd.MarkFlagRequired("org")
	//nolint:errcheck
	findCustomSnapshotForTagCmd.MarkFlagRequired("repo")
	//nolint:errcheck
	findCustomSnapshotForTagCmd.MarkFlagRequired("tag")
}

func runFindCustomSnapshotForTag(cmd *cobra.Command, args []string) error {
	// Get flag values
	baseApplicationName, _ := cmd.Flags().GetString("base-application-name")
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")

	outputFormat, err := common.GetOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	// Find custom snapshot
	snapshot, url, err := konflux.FindCustomSnapshotForTag(context.Background(), c, baseApplicationName, org, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to find custom snapshot for tag: %w", err)
	}

	// Format output
	return common.FormatOutput(outputFormat, SnapshotOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		Snapshot:     snapshot,
		URL:          url,
	})
}
