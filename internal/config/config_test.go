package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/claude-setup/internal/config"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-setup.yaml")

	want := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd", "strict-review"},
	}

	if err := config.Write(path, want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.Version != want.Version {
		t.Errorf("Version: got %q, want %q", got.Version, want.Version)
	}
	if got.Source != want.Source {
		t.Errorf("Source: got %q, want %q", got.Source, want.Source)
	}
	if got.Ref != want.Ref {
		t.Errorf("Ref: got %q, want %q", got.Ref, want.Ref)
	}
	if got.Profile != want.Profile {
		t.Errorf("Profile: got %q, want %q", got.Profile, want.Profile)
	}
	if len(got.Flavors) != len(want.Flavors) {
		t.Errorf("Flavors len: got %d, want %d", len(got.Flavors), len(want.Flavors))
	}
}

func TestReadMissing(t *testing.T) {
	_, err := config.Read("/nonexistent/.claude-setup.yaml")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}
