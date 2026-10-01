//go:build integration

package exporter_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/inf0-dev/alignment-matrix/internal/pkg/exporter"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func TestMarkdownGolden(t *testing.T) {
	recordPath := filepath.Join("testdata", "record.yaml")
	goldenPath := filepath.Join("testdata", "record.md.golden")

	rec, err := parser.ReadRecord(recordPath)
	require.NoError(t, err)

	got, err := exporter.Markdown(rec)
	require.NoError(t, err)

	if *update {
		err := os.WriteFile(goldenPath, []byte(got), 0644)
		require.NoError(t, err, "failed to update golden file")
		t.Log("golden file updated")
		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "failed to read golden file; run with -update to generate")

	assert.Equal(t, string(want), got)
}
