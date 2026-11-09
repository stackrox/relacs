package konflux

import (
	"github.com/spf13/cobra"
)

// KonfluxCmd represents the konflux command group
var KonfluxCmd = &cobra.Command{
	Use:   "konflux",
	Short: "Konflux-related utilities",
	Long:  `Commands for interacting with Konflux build system`,
}

func init() {
	// Subcommands will be added here
	KonfluxCmd.AddCommand(getCustomSnapshotCmd)
	KonfluxCmd.AddCommand(listConformaRunsCmd)
	KonfluxCmd.AddCommand(getOperatorBundleImageCmd)
	KonfluxCmd.AddCommand(getOperatorIndexImageCmd)
	KonfluxCmd.AddCommand(getReleaseCmd)
	KonfluxCmd.AddCommand(getReleasedOperatorBundleImageCmd)
	KonfluxCmd.AddCommand(listConformaViolationsCmd)
	KonfluxCmd.AddCommand(listReleasesCmd)
	KonfluxCmd.AddCommand(listPipelineRunsForTagCmd)
	// TODO ROX-37182: consider merging the prepare-stage-release and prepare-stage-cdn-release commands
	KonfluxCmd.AddCommand(prepareStageReleaseCmd)
	KonfluxCmd.AddCommand(prepareStageCdnReleaseCmd)
	KonfluxCmd.AddCommand(promoteReleaseToProdCmd)
	KonfluxCmd.AddCommand(promoteCdnReleaseToProdCmd)
}
