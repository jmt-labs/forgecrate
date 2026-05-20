package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func PatchPermissionMode(destDir string, mode string, cfg *config.Config) error {
	settingsPath := filepath.Join(destDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("settings.json lesen: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("settings.json parsen: %w", err)
	}

	if mode == "" {
		delete(m, "permissionMode")
	} else {
		m["permissionMode"] = mode
	}

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	out = append(out, '\n')

	if err := os.WriteFile(settingsPath, out, 0644); err != nil {
		return fmt.Errorf("settings.json schreiben: %w", err)
	}

	if cfg.DeployedFiles == nil {
		cfg.DeployedFiles = map[string]string{}
	}
	cfg.DeployedFiles[".claude/settings.json"] = hashBytes(out)
	return nil
}
