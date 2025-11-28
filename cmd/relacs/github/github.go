package github

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GitHubCmd represents the github command group
var GitHubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub-related utilities",
	Long:  `Commands for interacting with GitHub repositories, teams, and organizations`,
}

func init() {
	// Subcommands will be added here
	GitHubCmd.AddCommand(getCommitForTagCmd)
	GitHubCmd.AddCommand(isTeamMemberCmd)
}

// getOutputFormat retrieves the output format flag from the command
func getOutputFormat(cmd *cobra.Command) (string, error) {
	format, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", fmt.Errorf("failed to get output format: %w", err)
	}
	return format, nil
}

// formatOutput formats the output based on the specified format
func formatOutput(format string, output interface{}) error {
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
		// For text output, each command should handle its own formatting
		// This is a fallback that shouldn't be reached
		if commitOutput, ok := output.(CommitOutput); ok {
			fmt.Println(commitOutput.Commit)
			return nil
		}
		return fmt.Errorf("text format not supported for this output type")
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
