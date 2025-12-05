package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

var getReleaseCmd = &cobra.Command{
	Use:   "get-release",
	Short: "Get a release for a given Konflux release",
	Long:  `Retrieve the release associated with a specific Konflux release`,
	RunE:  common.RunCommand(runGetRelease),
}

func init() {
	getReleaseCmd.Flags().String("name", "", "Konflux release name (required)")

	//nolint:errcheck
	getReleaseCmd.MarkFlagRequired("name")
}

func runGetRelease(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	name, _ := cmd.Flags().GetString("name")
	release, err := konflux.GetRelease(ctx, c, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}
	return ReleaseOutput{
		Name:        name,
		ReleasePlan: release.Spec.ReleasePlan,
		Snapshot:    release.Spec.Snapshot,
		InternalURL: release.GetAdvisoryInternalURL(),
		PublicURL:   release.GetAdvisoryPublicURL(),
		IsCompleted: release.IsCompleted(),
	}, nil
}

type ReleaseOutput struct {
	Name        string `json:"name" yaml:"name"`
	ReleasePlan string `json:"release-plan" yaml:"release-plan"`
	Snapshot    string `json:"snapshot" yaml:"snapshot"`
	InternalURL string `json:"internal-url" yaml:"internal-url"`
	PublicURL   string `json:"public-url" yaml:"public-url"`
	IsCompleted bool   `json:"is-completed" yaml:"is-completed"`
}

func (r ReleaseOutput) String() string {
	return fmt.Sprintf("Internal URL: %s", r.InternalURL)
}
