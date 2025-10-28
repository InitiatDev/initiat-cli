package setup

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func ParseFile(filename string) (*SetupConfig, error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is controlled by user input for setup files
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return Parse(data)
}

func Parse(data []byte) (*SetupConfig, error) {
	var config SetupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &config, nil
}

func ParseFromReader(reader io.Reader) (*SetupConfig, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}
	return Parse(data)
}
