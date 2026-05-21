package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

func TestSetPermissionModeRoundtrip(t *testing.T) {
	dir := t.TempDir()

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
		DeployedFiles: map[string]string{
			".claude/settings.json": "sha256:old",
		},
	}
	cfgPath := filepath.Join(dir, ".forgecrate.yaml")
	if err := config.Write(cfgPath, cfg); err != nil {
		t.Fatalf("config.Write: %v", err)
	}

	settingsDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"),
		[]byte(`{"model":"claude-sonnet-4-6"}`+"\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := deploy.PatchPermissionMode(dir, "bypass", &cfg); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}
	if cfg.DeployedFiles[".claude/settings.json"] == "sha256:old" {
		t.Error("DeployedFiles hash was not updated")
	}

	cfg.PermissionMode = "bypass"
	if err := config.Write(cfgPath, cfg); err != nil {
		t.Fatalf("config.Write: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v", m["permissionMode"])
	}

	got, _ := config.Read(cfgPath)
	if got.PermissionMode != "bypass" {
		t.Errorf("config PermissionMode: got %q", got.PermissionMode)
	}
}

func TestSetPermissionModeValidation(t *testing.T) {
	for _, mode := range []string{"bypass", "plan", "ask", "auto"} {
		if err := config.ValidatePermissionMode(mode); err != nil {
			t.Errorf("mode %q should be valid: %v", mode, err)
		}
	}
	if err := config.ValidatePermissionMode("foo"); err == nil {
		t.Error("expected error for invalid mode")
	}
}
