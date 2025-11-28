package konflux

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/rel-acs/pkg/client"
	"github.com/stackrox/rel-acs/pkg/konflux"
	"gopkg.in/yaml.v3"
)

type SnapshotOutput struct {
	Organization string `json:"organization" yaml:"organization"`
	Repository   string `json:"repository" yaml:"repository"`
	Tag          string `json:"tag" yaml:"tag"`
	Snapshot     string `json:"snapshot" yaml:"snapshot"`
	URL          string `json:"url" yaml:"url"`
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

	findCustomSnapshotForTagCmd.MarkFlagRequired("org")
	findCustomSnapshotForTagCmd.MarkFlagRequired("repo")
	findCustomSnapshotForTagCmd.MarkFlagRequired("tag")
}

func runFindCustomSnapshotForTag(cmd *cobra.Command, args []string) error {
	// Get flag values
	org, _ := cmd.Flags().GetString("org")
	repo, _ := cmd.Flags().GetString("repo")
	tag, _ := cmd.Flags().GetString("tag")

	outputFormat, err := getOutputFormat(cmd)
	if err != nil {
		return err
	}

	c, err := client.NewClient()
	if err != nil {
		return err
	}

	// Find custom snapshot
	snapshot, url, err := konflux.FindCustomSnapshotForTag(context.Background(), c, org, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to find custom snapshot for tag: %w", err)
	}

	// Format output
	return formatOutput(outputFormat, SnapshotOutput{
		Organization: org,
		Repository:   repo,
		Tag:          tag,
		Snapshot:     snapshot,
		URL:          url,
	})
}

// formatOutput formats the output based on the specified format
func formatOutput(format string, output interface{}) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		return encoder.Encode(output)
	case "text":
		if snapshotOutput, ok := output.(SnapshotOutput); ok {
			fmt.Println("Name: ", snapshotOutput.Snapshot)
			fmt.Println("URL:  ", snapshotOutput.URL)
			return nil
		}
		return fmt.Errorf("text format not supported for this output type")
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
