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

func TestRunCopiesSkillsFromBase(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "base/skills/release/SKILL.md", "# Release Skill")

	cfg := config.Config{Profile: "backend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, ".claude", "skills", "release", "SKILL.md"))
	if err != nil {
		t.Fatalf("skill not copied: %v", err)
	}
	if string(got) != "# Release Skill" {
		t.Errorf("content: %q", got)
	}
}

func TestCopySkillsFirstWins(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "base/skills/release/SKILL.md", "base-content")
	writeFile(t, src, "profiles/frontend/skills/release/SKILL.md", "profile-content")

	cfg := config.Config{Profile: "frontend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(dst, ".claude", "skills", "release", "SKILL.md"))
	if string(got) != "base-content" {
		t.Errorf("first-wins failed: got %q, want %q", string(got), "base-content")
	}
}

func TestCopySkillsMissingDirOK(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)

	cfg := config.Config{Profile: "backend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("no error expected when skills dir missing: %v", err)
	}
}

func TestCopySkillsProfileAndFlavor(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "profiles/frontend/skills/frontend-tips/SKILL.md", "frontend-tips")
	writeFile(t, src, "flavors/strict-review/skills/review-tips/SKILL.md", "review-tips")

	cfg := config.Config{Profile: "frontend", Flavors: []string{"strict-review"}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(dst, ".claude", "skills", "frontend-tips", "SKILL.md")); err != nil {
		t.Errorf("frontend-tips skill missing")
	}
	if _, err := os.ReadFile(filepath.Join(dst, ".claude", "skills", "review-tips", "SKILL.md")); err != nil {
		t.Errorf("review-tips skill missing")
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
