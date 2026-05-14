package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/markus/claude-setup/internal/config"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-setup.yaml")

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/markus/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd", "strict-review"},
	}

	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.Profile != "backend" {
		t.Errorf("Profile: got %q, want %q", got.Profile, "backend")
	}
	if len(got.Flavors) != 2 {
		t.Errorf("Flavors: got %d, want 2", len(got.Flavors))
	}
}

func TestReadMissing(t *testing.T) {
	_, err := config.Read("/nonexistent/.claude-setup.yaml")
	if !os.IsNotExist(err) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}
