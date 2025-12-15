package konflux

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/konflux"
)

type OperatorCatalogImagesOutput struct {
	Commit string                 `json:"commit" yaml:"commit"`
	Branch string                 `json:"branch" yaml:"branch"`
	Images []OperatorCatalogImage `json:"images" yaml:"images"`
}

type OperatorCatalogImage struct {
	Application  string `json:"application" yaml:"application"`
	ImageRef     string `json:"image-ref" yaml:"image-ref"`
	SnapshotName string `json:"snapshot-name" yaml:"snapshot-name"`
}

func (s OperatorCatalogImagesOutput) String() string {
	output := fmt.Sprintf("Commit: %s\n", s.Commit)
	output += fmt.Sprintf("Branch: %s\n", s.Branch)
	output += fmt.Sprintln("Operator catalog images:")
	for _, image := range s.Images {
		output += fmt.Sprintf("  %s (application: %s, snapshot: %s)\n", image.ImageRef, image.Application, image.SnapshotName)
	}
	return output
}

var listFbcBuiltImagesCmd = &cobra.Command{
	Use:   "list-fbc-built-images",
	Short: "List the FBC built images for a given Git commit and branch",
	Long:  `Retrieve the FBC built images (operator catalog images) associated with a specific Git commit and branch`,
	RunE:  common.RunCommand(runListFbcBuiltImages),
}

func init() {
	listFbcBuiltImagesCmd.Flags().String("commit", "", "Git commit (required)")
	listFbcBuiltImagesCmd.Flags().String("branch", "", "Git branch (required)")

	//nolint:errcheck
	listFbcBuiltImagesCmd.MarkFlagRequired("commit")
	//nolint:errcheck
	listFbcBuiltImagesCmd.MarkFlagRequired("branch")
}

func runListFbcBuiltImages(cmd *cobra.Command, args []string, c *client.Client, ctx context.Context) (any, error) {
	// Get flag values
	commit, _ := cmd.Flags().GetString("commit")
	branch, _ := cmd.Flags().GetString("branch")

	snapshots, err := konflux.ListFbcSnapshots(ctx, c, commit, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get FBC snapshots: %w", err)
	}

	images := []OperatorCatalogImage{}
	for _, snapshot := range snapshots {
		for _, component := range snapshot.Spec.Components {
			images = append(images, OperatorCatalogImage{
				Application:  snapshot.Spec.Application,
				ImageRef:     component.ContainerImage,
				SnapshotName: snapshot.Metadata.Name,
			})
		}
	}

	return OperatorCatalogImagesOutput{
		Commit: commit,
		Branch: branch,
		Images: images,
	}, nil
}
