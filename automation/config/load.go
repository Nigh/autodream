package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadJSON reads path and unmarshals into dest.
func LoadJSON(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("config: json %s: %w", path, err)
	}
	return nil
}

// LoadYAML reads path and unmarshals into dest.
func LoadYAML(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("config: yaml %s: %w", path, err)
	}
	return nil
}

// DecodeJSON unmarshals JSON bytes into dest.
func DecodeJSON(b []byte, dest any) error {
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("config: json: %w", err)
	}
	return nil
}

// DecodeYAML unmarshals YAML bytes into dest.
func DecodeYAML(b []byte, dest any) error {
	if err := yaml.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("config: yaml: %w", err)
	}
	return nil
}
