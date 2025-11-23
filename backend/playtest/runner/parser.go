package runner

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PlaytestDefinition represents a complete playtest scenario
type PlaytestDefinition struct {
	Turns []Turn `yaml:"turns"`
}

// Turn represents a single client action in a playtest
type Turn struct {
	Client  string                 `yaml:"client"`
	Message map[string]interface{} `yaml:"message"`
}

// ParsePlaytest reads and parses a YAML playtest file
func ParsePlaytest(filepath string) (*PlaytestDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	var playtest PlaytestDefinition
	if err := yaml.Unmarshal(data, &playtest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(playtest.Turns) == 0 {
		return nil, fmt.Errorf("playtest must contain at least one turn")
	}

	return &playtest, nil
}
