package konflux

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

var getReleaseCmd = &cobra.Command{
	Use:   "get-release",
	Short: "Get information about a given Konflux release",
	Long:  `Retrieve the information about a specific Konflux release`,
	RunE:  common.RunCommand(runGetRelease),
}

func init() {
	getReleaseCmd.Flags().String("name", "", "Konflux release name (required)")

	//nolint:errcheck
	getReleaseCmd.MarkFlagRequired("name")
}

func runGetRelease(cmd *cobra.Command, args []string, c *client.Client) (any, error) {
	ctx := cmd.Context()
	name, _ := cmd.Flags().GetString("name")
	release, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}
	url, err := release.GetURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get release URL: %w", err)
	}
	return ReleaseOutput{
		Name:        name,
		ReleasePlan: release.Spec.ReleasePlan,
		Snapshot:    release.Spec.Snapshot,
		InternalURL: release.GetAdvisoryInternalURL(),
		PublicURL:   release.GetAdvisoryPublicURL(),
		IsCompleted: release.IsCompleted(),
		IsSucceeded: release.IsSucceeded(),
		URL:         url,
	}, nil
}

type ReleaseOutput struct {
	Name        string `json:"name" yaml:"name"`
	ReleasePlan string `json:"release-plan" yaml:"release-plan"`
	Snapshot    string `json:"snapshot" yaml:"snapshot"`
	InternalURL string `json:"internal-url" yaml:"internal-url"`
	PublicURL   string `json:"public-url" yaml:"public-url"`
	IsCompleted bool   `json:"is-completed" yaml:"is-completed"`
	IsSucceeded bool   `json:"is-succeeded" yaml:"is-succeeded"`
	URL         string `json:"url" yaml:"url"`
}

func (o ReleaseOutput) String() string {
	output := fmt.Sprintf("Name:         %s\n", o.Name)
	output += fmt.Sprintf("Release Plan: %s\n", o.ReleasePlan)
	output += fmt.Sprintf("Snapshot:     %s\n", o.Snapshot)
	output += fmt.Sprintf("Internal URL: %s\n", o.InternalURL)
	output += fmt.Sprintf("Public URL:   %s\n", o.PublicURL)
	output += fmt.Sprintf("Completed:    %t\n", o.IsCompleted)
	output += fmt.Sprintf("Succeeded:    %t\n", o.IsSucceeded)
	output += fmt.Sprintf("URL:          %s\n", o.URL)
	return output
}
