package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type ReleasedOperatorBundleImageOutput struct {
	Release             string `json:"release" yaml:"release"`
	OperatorBundleImage string `json:"operator-bundle-image" yaml:"operator-bundle-image"`
}

func (o ReleasedOperatorBundleImageOutput) String() string {
	output := fmt.Sprintf("Release:               %s\n", o.Release)
	output += fmt.Sprintf("Operator Bundle Image: %s\n", o.OperatorBundleImage)
	return output
}

var getReleasedOperatorBundleImageCmd = &cobra.Command{
	Use:   "get-released-operator-bundle-image",
	Short: "Get the operator bundle image for a given release",
	Long:  `Retrieve the operator bundle image associated with a specific release from the published advisory`,
	RunE:  common.RunCommand(runGetReleasedOperatorBundleImage),
}

func init() {
	getReleasedOperatorBundleImageCmd.Flags().String("release", "", "Release (required)")

	//nolint:errcheck
	getReleasedOperatorBundleImageCmd.MarkFlagRequired("release")
}

func runGetReleasedOperatorBundleImage(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	// Get flag values
	release, _ := cmd.Flags().GetString("release")

	operatorBundleImage, err := konflux.GetOperatorBundleImageForRelease(ctx, c, release)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator bundle image for release %s: %w", release, err)
	}

	return ReleasedOperatorBundleImageOutput{
		Release:             release,
		OperatorBundleImage: operatorBundleImage,
	}, nil
}
