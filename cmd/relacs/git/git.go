package git

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GitCmd represents the git command group
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git-related utilities",
	Long:  `Commands for interacting with Git repositories and tags`,
}

func init() {
	// Subcommands will be added here
	GitCmd.AddCommand(getCommitForTagCmd)
}

func formatOutput(format string, output CommitOutput) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		return encoder.Encode(output)
	case "text":
		fmt.Println(output.Commit)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
