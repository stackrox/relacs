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
	Tag                 string `json:"tag" yaml:"tag"`
	OperatorBundleImage string `json:"operator-bundle-image" yaml:"operator-bundle-image"`
}

func (o ReleasedOperatorBundleImageOutput) String() string {
	output := fmt.Sprintf("Tag: %s\n", o.Tag)
	output += fmt.Sprintf("Operator Bundle Image: %s\n", o.OperatorBundleImage)
	return output
}

var getReleasedOperatorBundleImageCmd = &cobra.Command{
	Use:   "get-released-operator-bundle-image",
	Short: "Get the operator bundle image for a given Git tag",
	Long:  `Retrieve the operator bundle image associated with a specific Git tag`,
	RunE:  common.RunCommand(runGetReleasedOperatorBundleImage),
}

func init() {
	getReleasedOperatorBundleImageCmd.Flags().String("release", "", "Release (required)")
	getReleasedOperatorBundleImageCmd.Flags().String("tag", "", "Git tag (required)")

	//nolint:errcheck
	getReleasedOperatorBundleImageCmd.MarkFlagRequired("release")
	//nolint:errcheck
	getReleasedOperatorBundleImageCmd.MarkFlagRequired("tag")
}

func runGetReleasedOperatorBundleImage(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	// Get flag values
	release, _ := cmd.Flags().GetString("release")
	tag, _ := cmd.Flags().GetString("tag")

	operatorBundleImage, err := konflux.GetOperatorBundleImageForRelease(ctx, c, release, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator bundle image for release %s and tag %s: %w", release, tag, err)
	}

	return ReleasedOperatorBundleImageOutput{
		Release:             release,
		Tag:                 tag,
		OperatorBundleImage: operatorBundleImage,
	}, nil
}
