package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/inf0-dev/alma/app/cli/export"
	"github.com/inf0-dev/alma/app/cli/render"
	"github.com/inf0-dev/alma/app/cli/serve"
	"github.com/inf0-dev/alma/app/cli/validate"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "alma",
		Short: "alma — alignment matrix CLI",
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
