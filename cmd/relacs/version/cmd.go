package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
)

var version = "0.0.0"

type VersionOutput struct {
	Version string `json:"version" yaml:"version"`
}

func (v VersionOutput) String() string {
	return fmt.Sprintf("relacs version %s", v.Version)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of relacs",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	outputFormat, err := common.GetOutputFormat(cmd)
	if err != nil {
		return err
	}

	return common.FormatOutput(outputFormat, VersionOutput{
		Version: version,
	})
}
