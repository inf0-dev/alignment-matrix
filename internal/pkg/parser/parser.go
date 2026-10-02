package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/inf0-dev/alma/api/v1"
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

// ParseDocument parses a Document from raw bytes. Format is determined by filename extension.
func ParseDocument(data []byte, filename string) (*v1.Document, error) {
	return parse[v1.Document](data, filename)
}

// ParseRecord parses a Record from raw bytes. Format is determined by filename extension.
func ParseRecord(data []byte, filename string) (*v1.Record, error) {
	return parse[v1.Record](data, filename)
}

func parse[T Input](data []byte, filename string) (*T, error) {
	var out T
	ext := filepath.Ext(filename)
	switch ext {
	case extJSON:
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	case extYAML, extYML:
		if err := yaml.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
	return &out, nil
}

// versionProbe is used to detect the type of input file by reading version fields.
type versionProbe struct {
	Version  string `json:"version" yaml:"version"`
	Metadata struct {
		Version string `json:"version" yaml:"version"`
	} `json:"metadata" yaml:"metadata"`
}

// DetectType returns "record" or "document" based on version fields in the data.
func DetectType(data []byte, filename string) (string, error) {
	var probe versionProbe

	ext := filepath.Ext(filename)
	switch ext {
	case extJSON:
		if err := json.Unmarshal(data, &probe); err != nil {
			return "", fmt.Errorf("failed to detect type: %w", err)
		}
	case extYAML, extYML:
		if err := yaml.Unmarshal(data, &probe); err != nil {
			return "", fmt.Errorf("failed to detect type: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	if probe.Version == v1.RecordVersionV1 {
		return "record", nil
	}
	if probe.Metadata.Version == v1.VersionV1 {
		return "document", nil
	}

	return "", fmt.Errorf("unable to detect type: no recognized version field found")
}
