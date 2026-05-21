package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".forgecrate.yaml")

	want := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
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
	_, err := config.Read("/nonexistent/.forgecrate.yaml")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}

func TestDeployedFilesRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".forgecrate.yaml")

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd"},
		DeployedFiles: map[string]string{
			".claude/settings.json":     "sha256:abc123",
			".claude/hooks/pre-tool.sh": "sha256:def456",
		},
	}

	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := config.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.DeployedFiles[".claude/settings.json"] != "sha256:abc123" {
		t.Errorf("settings.json hash lost: %q", got.DeployedFiles[".claude/settings.json"])
	}
	if got.DeployedFiles[".claude/hooks/pre-tool.sh"] != "sha256:def456" {
		t.Errorf("pre-tool.sh hash lost: %q", got.DeployedFiles[".claude/hooks/pre-tool.sh"])
	}
}

func TestHasFlavor(t *testing.T) {
	cfg := config.Config{Flavors: []string{"tdd", "strict-review"}}
	if !cfg.HasFlavor("tdd") {
		t.Error("expected HasFlavor(tdd) = true")
	}
	if !cfg.HasFlavor("strict-review") {
		t.Error("expected HasFlavor(strict-review) = true")
	}
	if cfg.HasFlavor("no-research") {
		t.Error("expected HasFlavor(no-research) = false")
	}
	empty := config.Config{}
	if empty.HasFlavor("anything") {
		t.Error("expected HasFlavor on empty config = false")
	}
}

func TestDeployedFilesOmittedWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".forgecrate.yaml")

	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "p"}
	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "deployed_files") {
		t.Error("deployed_files should be omitted when empty")
	}
}
