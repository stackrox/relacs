package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GetOutputFormat retrieves the output format flag from the command
func GetOutputFormat(cmd *cobra.Command) (string, error) {
	format, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", fmt.Errorf("failed to get output format: %w", err)
	}
	return format, nil
}

// MultiDocYAML is a slice of values rendered as separate YAML documents separated by ---.
type MultiDocYAML []any

// FormatOutput formats the output based on the specified format
func FormatOutput(format string, output any) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		if docs, ok := output.(MultiDocYAML); ok {
			for _, doc := range docs {
				if err := encoder.Encode(doc); err != nil {
					return err
				}
			}
			return encoder.Close()
		}
		return encoder.Encode(output)
	case "text":
		fmt.Println(output)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
