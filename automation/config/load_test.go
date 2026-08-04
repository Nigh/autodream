package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nigh/autodream/automation/config"
)

func TestLoadYAMLJSON(t *testing.T) {
	dir := t.TempDir()
	y := filepath.Join(dir, "c.yaml")
	j := filepath.Join(dir, "c.json")
	if err := os.WriteFile(y, []byte("threshold: 0.8\nname: color\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(j, []byte(`{"threshold":0.5,"name":"json"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var got struct {
		Threshold float64 `yaml:"threshold" json:"threshold"`
		Name      string  `yaml:"name" json:"name"`
	}
	if err := config.LoadYAML(y, &got); err != nil {
		t.Fatal(err)
	}
	if got.Threshold != 0.8 || got.Name != "color" {
		t.Fatalf("yaml: %+v", got)
	}
	if err := config.LoadJSON(j, &got); err != nil {
		t.Fatal(err)
	}
	if got.Threshold != 0.5 || got.Name != "json" {
		t.Fatalf("json: %+v", got)
	}
}
