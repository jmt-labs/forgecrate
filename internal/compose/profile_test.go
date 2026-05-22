package compose_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestLoadProfileConfigWithExtends(t *testing.T) {
	src := t.TempDir()
	dir := filepath.Join(src, "profiles", "fullstack")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "profile.yaml"), []byte("extends:\n  - backend\n  - frontend\n"), 0644)

	cfg := compose.LoadProfileConfig(src, "fullstack")
	if len(cfg.Extends) != 2 || cfg.Extends[0] != "backend" || cfg.Extends[1] != "frontend" {
		t.Errorf("Extends: got %v, want [backend frontend]", cfg.Extends)
	}
}

func TestLoadProfileConfigMissingFileReturnsEmpty(t *testing.T) {
	cfg := compose.LoadProfileConfig(t.TempDir(), "backend")
	if len(cfg.Extends) != 0 {
		t.Errorf("expected empty Extends, got %v", cfg.Extends)
	}
}
