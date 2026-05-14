package deploy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markus/claude-setup/internal/config"
	"github.com/markus/claude-setup/internal/deploy"
)

func TestDeploy(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "# Base")
	writeFile(t, src, "base/.claude/settings.json", `{"hooks":{}}`)
	writeFile(t, src, "base/hooks/prompt-submit.sh", "#!/bin/bash\necho ok")
	writeFile(t, src, "base/hooks/pre-tool.sh", "#!/bin/bash\necho ok")

	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/markus/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude-setup.yaml")); err != nil {
		t.Errorf(".claude-setup.yaml missing")
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "hooks", "prompt-submit.sh")); err != nil {
		t.Errorf("prompt-submit.sh missing")
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "hooks", "pre-tool.sh")); err != nil {
		t.Errorf("pre-tool.sh missing")
	}
}

func TestRunInstallsExtensions(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	claudeDir := t.TempDir()
	argsFile := filepath.Join(claudeDir, "calls.txt")
	fakeClaude := filepath.Join(claudeDir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", argsFile)
	os.WriteFile(fakeClaude, []byte(script), 0755)

	baseDir := filepath.Join(src, "base")
	os.MkdirAll(baseDir, 0755)
	os.WriteFile(filepath.Join(baseDir, "CLAUDE.md"), []byte("<!-- GENERATED:BEGIN -->\n# Claude\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"), 0644)
	os.MkdirAll(filepath.Join(baseDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(baseDir, ".claude", "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)
	os.WriteFile(filepath.Join(baseDir, "extensions.yaml"), []byte("plugins:\n  - name: superpowers\n    source: claude-plugins-official/superpowers\n"), 0644)

	cfg := config.Config{Profile: "backend"}
	if err := deploy.RunWithClaude(src, dst, cfg, fakeClaude); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install claude-plugins-official/superpowers") {
		t.Errorf("plugin not installed, calls: %q", string(data))
	}
}

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}
