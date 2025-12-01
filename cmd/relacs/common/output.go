package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// getOutputFormat retrieves the output format flag from the command
func GetOutputFormat(cmd *cobra.Command) (string, error) {
	format, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", fmt.Errorf("failed to get output format: %w", err)
	}
	return format, nil
}

// formatOutput formats the output based on the specified format
func FormatOutput(format string, output interface{}) error {
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
		fmt.Println(output)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
