package validate

import (
	"fmt"
	"strings"

	"github.com/inf0-dev/alignment-matrix/internal/pkg/parser"
	"github.com/spf13/cobra"
)

func NewValidateCommand() *cobra.Command {
	var path string
	var doc_type = "document"

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the alignment matrix",
		Long:  "Validate the alignment matrix for issues and report them to the user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch strings.ToLower(doc_type) {
			case "document":
				doc, err := parser.ReadDocument(path)
				if err != nil {
					return fmt.Errorf("failed to read document: %v", err)
				}
				if err := doc.Validate(); err != nil {
					return fmt.Errorf("document validation failed: %v", err)
				}
			case "record":
				record, err := parser.ReadRecord(path)
				if err != nil {
					return fmt.Errorf("failed to read record: %v", err)
				}
				if err := record.Validate(); err != nil {
					return fmt.Errorf("record validation failed: %v", err)
				}
			default:
				return fmt.Errorf("invalid type: %s. Must be one of: [document, record] (case-insensitive)", doc_type)
			}

			cmd.Println("Valid!")
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the alignment matrix file")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().StringVarP(&doc_type, "type", "t", "", "Type of the alignment matrix file, one of: [document, record] (case-insensitive). Default is document, if unset.")

	return cmd
}
