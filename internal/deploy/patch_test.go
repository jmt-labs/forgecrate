package deploy_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

func TestPatchPermissionMode(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	initial := `{"model":"claude-sonnet-4-6","permissions":{"allow":["Bash"]}}` + "\n"
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := config.Config{
		DeployedFiles: map[string]string{".claude/settings.json": "sha256:old"},
	}

	if err := deploy.PatchPermissionMode(dir, "bypass", &cfg); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v, want bypass", m["permissionMode"])
	}
	if m["model"] != "claude-sonnet-4-6" {
		t.Error("model should be preserved")
	}
	if cfg.DeployedFiles[".claude/settings.json"] == "sha256:old" {
		t.Error("hash should be updated after patch")
	}
}

func TestPatchPermissionModeRemovesKey(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	initial := `{"permissionMode":"bypass","model":"claude-sonnet-4-6"}` + "\n"
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := config.Config{}
	if err := deploy.PatchPermissionMode(dir, "", &cfg); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	if strings.Contains(string(data), "permissionMode") {
		t.Error("permissionMode should be removed when mode is empty")
	}
}

func TestPatchPermissionModeMissingSettings(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	err := deploy.PatchPermissionMode(dir, "bypass", &cfg)
	if err == nil {
		t.Error("expected error for missing settings.json")
	}
}
