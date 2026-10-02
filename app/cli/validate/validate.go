package validate

import (
	"fmt"
	"os"

	"github.com/inf0-dev/alma/internal/pkg/parser"
	"github.com/spf13/cobra"
)

func NewValidateCommand() *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an alma document",
		Long:  "Validate an alma document or record for issues and report them to the user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			docType, err := parser.DetectType(data, path)
			if err != nil {
				return err
			}

			switch docType {
			case "document":
				doc, err := parser.ParseDocument(data, path)
				if err != nil {
					return fmt.Errorf("failed to parse document: %w", err)
				}
				if err := doc.Validate(); err != nil {
					return fmt.Errorf("document validation failed: %w", err)
				}
			case "record":
				rec, err := parser.ParseRecord(data, path)
				if err != nil {
					return fmt.Errorf("failed to parse record: %w", err)
				}
				if err := rec.Validate(); err != nil {
					return fmt.Errorf("record validation failed: %w", err)
				}
			default:
				return fmt.Errorf("unrecognized type: %s", docType)
			}

			cmd.Println("Valid!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the alma file")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}
