package serve

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	v1 "github.com/inf0-dev/alignment-matrix/api/v1"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/engine"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/parser"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/server"
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	var path string
	var addr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the alignment matrix as a web application",
		RunE: func(cmd *cobra.Command, args []string) error {
			var rec *v1.Record
			if path != "" {
				var err error
				rec, err = loadRecord(path)
				if err != nil {
					return err
				}
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			srv := server.New(server.Config{Addr: addr}, rec)
			return srv.Run(ctx)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the document or record file (optional; if omitted, starts with upload UI)")
	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "Address to listen on")

	return cmd
}

func loadRecord(path string) (*v1.Record, error) {
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
