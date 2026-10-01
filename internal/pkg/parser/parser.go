package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/inf0-dev/alignment-matrix/api/v1"
	"go.yaml.in/yaml/v4"
)

const (
	extJSON = ".json"
	extYAML = ".yaml"
	extYML  = ".yml"
)

type Input interface {
	v1.Document | v1.Record
}

func read[T Input](path string) (*T, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var out T

	ext := filepath.Ext(path)
	switch ext {
	case extJSON:
		if err := json.Unmarshal(content, &out); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case extYAML, extYML:
		if err := yaml.Unmarshal(content, &out); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return &out, nil
}

func ReadDocument(path string) (*v1.Document, error) {
	return read[v1.Document](path)
}

func ReadRecord(path string) (*v1.Record, error) {
	return read[v1.Record](path)
}
