package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/inf0-dev/alignment-matrix/app/cli/export"
	"github.com/inf0-dev/alignment-matrix/app/cli/render"
	"github.com/inf0-dev/alignment-matrix/app/cli/serve"
	"github.com/inf0-dev/alignment-matrix/app/cli/validate"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "align",
		Short: "Alignment Matrix CLI",
	}

	cmd.AddCommand(
		validate.NewValidateCommand(),
		render.NewRenderCommand(),
		export.NewExportCommand(),
		serve.NewServeCommand(),
	)

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}
