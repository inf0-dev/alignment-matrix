//go:build integration

package parser_test

import (
	"path/filepath"
	"testing"

	v1 "github.com/inf0-dev/alma/api/v1"
	"github.com/inf0-dev/alma/internal/pkg/parser"
	"github.com/inf0-dev/alma/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	return filepath.Join("testdata", name)
}

var expectedDocument = v1.Document{
	Metadata: v1.Metadata{
		Version: v1.VersionV1,
	},
	Schema: v1.Schema{
		Title: "Test Schema",
		Requirements: []v1.Requirement{
			{ID: "req1", Description: "First requirement", IsHard: true},
		},
		Items: []v1.Item{
			{
				ID:          "q1",
				Description: "Question 1",
				Kind:        v1.KindChoice,
				ChoiceOptions: []v1.ChoiceOption{
					{ID: "yes", Description: "Yes"},
					{ID: "no", Description: "No"},
				},
			},
		},
		DesignOptions: []v1.DesignOption{
			{
				ID:    "opt1",
				Title: "Option 1",
				RequirementsMet: v1.RequirementsMet{
					"req1": v1.RequirementStatus{Met: true},
				},
			},
		},
	},
}

var expectedRecord = v1.Record{
	Version:     v1.RecordVersionV1,
	Model:       expectedDocument,
	ModelSHA256: "abc123",
	DecidedOn:   "2026-10-01",
	Requirements: []v1.RecordRequirement{
		{ID: "req1", IsHard: true, Checked: true},
	},
	Items: []v1.RecordItem{
		{ID: "q1", Kind: v1.KindChoice, Answer: testutil.StrPtr("yes")},
	},
	Options: []v1.RecordOption{},
	Final: &v1.FinalDecision{
		Option:    "opt1",
		Title:     "Option 1",
		Rationale: "Best fit",
	},
}

func TestReadDocument(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr string
		want    *v1.Document
	}{
		{
			name: "valid YAML document",
			path: testdataPath("valid_document.yaml"),
			want: &expectedDocument,
		},
		{
			name: "valid JSON document",
			path: testdataPath("valid_document.json"),
			want: &expectedDocument,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: "path is empty",
		},
		{
			name:    "nonexistent file",
			path:    testdataPath("nonexistent.yaml"),
			wantErr: "failed to read file",
		},
		{
			name:    "unsupported extension",
			path:    testdataPath("file.txt"),
			wantErr: "unsupported file extension",
		},
		{
			name:    "invalid YAML",
			path:    testdataPath("invalid.yaml"),
			wantErr: "failed to unmarshal YAML",
		},
		{
			name:    "invalid JSON",
			path:    testdataPath("invalid.json"),
			wantErr: "failed to unmarshal JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parser.ReadDocument(tt.path)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, doc)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, doc)
			assert.Equal(t, tt.want, doc)
		})
	}
}

func TestReadRecord(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr string
		want    *v1.Record
	}{
		{
			name: "valid YAML record",
			path: testdataPath("valid_record.yaml"),
			want: &expectedRecord,
		},
		{
			name: "valid JSON record",
			path: testdataPath("valid_record.json"),
			want: &expectedRecord,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: "path is empty",
		},
		{
			name:    "nonexistent file",
			path:    testdataPath("nonexistent.yaml"),
			wantErr: "failed to read file",
		},
		{
			name:    "invalid YAML",
			path:    testdataPath("invalid.yaml"),
			wantErr: "failed to unmarshal YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := parser.ReadRecord(tt.path)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, rec)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rec)
			assert.Equal(t, tt.want, rec)
		})
	}
}
