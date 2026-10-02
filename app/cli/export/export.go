package export

import (
	"encoding/json"
	"fmt"

	v1 "github.com/inf0-dev/alma/api/v1"
	"github.com/inf0-dev/alma/internal/pkg/exporter"
	"github.com/inf0-dev/alma/internal/pkg/parser"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func NewExportCommand() *cobra.Command {
	var path string
	var format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a record to stdout in the given format",
		Long:  "Read an alma record and export it to stdout as markdown, JSON, or YAML.",
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := parser.ReadRecord(path)
			if err != nil {
				return fmt.Errorf("failed to read record: %w", err)
			}

			if err := record.Validate(); err != nil {
				return fmt.Errorf("record validation failed: %w", err)
			}

			out, err := render(record, format)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the record file")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().StringVarP(&format, "output", "o", "md", "Output format: md, json, yaml")

	return cmd
}

func render(record *v1.Record, format string) (string, error) {
	switch format {
	case "md", "markdown":
		return exporter.Markdown(record)
	case "json":
		data, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(data) + "\n", nil
	case "yaml", "yml":
		data, err := yaml.Marshal(record)
		if err != nil {
			return "", fmt.Errorf("failed to marshal YAML: %w", err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported format: %s. Must be one of: [md, json, yaml]", format)
	}
}
