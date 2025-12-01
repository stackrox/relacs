package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/konflux"
)

type OperatorBundleImageRefOutput struct {
	ImageRef string `json:"image_ref" yaml:"image_ref"`
}

func (o OperatorBundleImageRefOutput) String() string {
	return fmt.Sprintf("ImageRef: %s\n", o.ImageRef)
}

var getOperatorBundleImageFromReleaseCmd = &cobra.Command{
	Use:   "get-operator-bundle-image-from-release",
	Short: "Get the operator bundle image ref for a Konflux release",
	Long:  `Get the operator bundle image ref for a Konflux release by name and namespace`,
	RunE:  runGetOperatorBundleImageFromRelease,
}

func init() {
	getOperatorBundleImageFromReleaseCmd.Flags().String("name", "", "Release name (required)")
	getOperatorBundleImageFromReleaseCmd.Flags().String("namespace", "", "Release namespace (required)")

	//nolint:errcheck
	getOperatorBundleImageFromReleaseCmd.MarkFlagRequired("name")
	//nolint:errcheck
	getOperatorBundleImageFromReleaseCmd.MarkFlagRequired("namespace")
}

func runGetOperatorBundleImageFromRelease(cmd *cobra.Command, args []string) error {
	// Get flag values
	name, _ := cmd.Flags().GetString("name")
	namespace, _ := cmd.Flags().GetString("namespace")

	outputFormat, err := common.GetOutputFormat(cmd)
	if err != nil {
		return err
	}

	// Get operator bundle image ref for release
	operatorBundleImageRef, err := konflux.GetOperatorBundleImageRefForRelease(context.Background(), name, namespace)
	if err != nil {
		return fmt.Errorf("failed to get operator bundle image ref for release: %w", err)
	}

	// Format output
	return common.FormatOutput(outputFormat, OperatorBundleImageRefOutput{
		ImageRef: operatorBundleImageRef,
	})
}
