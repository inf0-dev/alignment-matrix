package render

import (
	"fmt"
	"os"

	v1 "github.com/inf0-dev/alma/api/v1"
	"github.com/inf0-dev/alma/internal/pkg/engine"
	"github.com/inf0-dev/alma/internal/pkg/parser"
	"github.com/inf0-dev/alma/internal/pkg/renderer"
	"github.com/spf13/cobra"
)

func NewRenderCommand() *cobra.Command {
	var path string
	var output string

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render an alma document to a self-contained HTML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			rec, err := loadAndParse(path)
			if err != nil {
				return err
			}

			html, err := renderer.HTML(rec)
			if err != nil {
				return fmt.Errorf("render failed: %w", err)
			}

			if output == "" {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), html)
				return nil
			}

			if err := os.WriteFile(output, []byte(html), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			cmd.Printf("Rendered to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the document or record file")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}

func loadAndParse(path string) (*v1.Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	docType, err := parser.DetectType(data, path)
	if err != nil {
		return nil, err
	}

	switch docType {
	case "document":
		doc, err := parser.ParseDocument(data, path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse document: %w", err)
		}
		if err := doc.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		return engine.Evaluate(doc, nil, nil)
	case "record":
		rec, err := parser.ParseRecord(data, path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record: %w", err)
		}
		if err := rec.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		return rec, nil
	default:
		return nil, fmt.Errorf("unrecognized type: %s", docType)
	}
}
