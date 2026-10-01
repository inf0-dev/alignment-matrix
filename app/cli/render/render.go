package render

import (
	"fmt"
	"os"

	"github.com/inf0-dev/alignment-matrix/internal/pkg/engine"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/parser"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/renderer"
	"github.com/spf13/cobra"
)

func NewRenderCommand() *cobra.Command {
	var path string
	var output string
	var docType string

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render the alignment matrix to a self-contained HTML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			var html string
			var err error

			switch docType {
			case "document":
				doc, err := parser.ReadDocument(path)
				if err != nil {
					return fmt.Errorf("failed to read document: %w", err)
				}
				if err := doc.Validate(); err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}
				rec, err := engine.Evaluate(doc, nil, nil)
				if err != nil {
					return fmt.Errorf("evaluation failed: %w", err)
				}
				html, err = renderer.HTML(rec)
				if err != nil {
					return fmt.Errorf("render failed: %w", err)
				}
			case "record":
				rec, err := parser.ReadRecord(path)
				if err != nil {
					return fmt.Errorf("failed to read record: %w", err)
				}
				if err := rec.Validate(); err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}
				html, err = renderer.HTML(rec)
				if err != nil {
					return fmt.Errorf("render failed: %w", err)
				}
			default:
				return fmt.Errorf("invalid type: %s. Must be one of: [document, record]", docType)
			}

			if output == "" {
				fmt.Fprint(cmd.OutOrStdout(), html)
				return nil
			}

			if err = os.WriteFile(output, []byte(html), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			cmd.Printf("Rendered to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the document or record file")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP(&docType, "type", "t", "document", "Input type: document or record")

	return cmd
}
