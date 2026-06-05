package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/relacs/cmd/relacs/common"
)

// version is set at build time using git describe
var version = "dev"

// GetVersion returns the current version of relacs
func GetVersion() string {
	return version
}

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
