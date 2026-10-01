package export

import (
	"github.com/spf13/cobra"
)

func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the alignment matrix to a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			panic("unimplemented")
		},
	}

	return cmd
}
