package research

import (
	"github.com/spf13/cobra"
)

var ResearchCmd = &cobra.Command{
	Use:   "research",
	Short: "Research-related utilities",
	Long:  `Commands for researching for the release process`,
}

func init() {
	ResearchCmd.AddCommand(findSecurityTrackersCmd)
}
