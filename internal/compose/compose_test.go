package compose_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/compose"
)

func TestCompose(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "# Base Claude")
	writeFile(t, src, "base/.claude/settings.json", `{"permissions":{"allow":["Bash"]}}`)
	writeFile(t, src, "base/.claude/commands/base-skill.md", "# Base Skill")
	writeFile(t, src, "profiles/backend/CLAUDE.md", "## Backend Profile")
	writeFile(t, src, "flavors/tdd/CLAUDE.md", "## TDD Flavor")

	req := compose.Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "backend",
		Flavors:   []string{"tdd"},
	}

	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	claudeMD, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md missing: %v", err)
	}
	for _, want := range []string{"# Base Claude", "## Backend Profile", "## TDD Flavor"} {
		if !strings.Contains(string(claudeMD), want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "settings.json")); err != nil {
		t.Errorf("settings.json missing: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "commands", "base-skill.md")); err != nil {
		t.Errorf("base-skill.md missing: %v", err)
	}
}

func TestComposeSettingsReturnsContent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	settingsDir := filepath.Join(src, "base", ".claude")
	_ = os.MkdirAll(settingsDir, 0755)
	_ = os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)

	req := compose.Request{SourceDir: src, DestDir: dst, Profile: "backend", Flavors: []string{}}
	content, err := compose.ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}
	if !strings.Contains(string(content), "claude-sonnet-4-6") {
		t.Errorf("expected model in content, got: %s", content)
	}
}

func TestComposeRunSkipsSettingsWhenFlagSet(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	settingsDir := filepath.Join(src, "base", ".claude")
	_ = os.MkdirAll(settingsDir, 0755)
	_ = os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)
	_ = os.WriteFile(filepath.Join(src, "base", "CLAUDE.md"), []byte("# Base\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"), 0644)

	req := compose.Request{SourceDir: src, DestDir: dst, Profile: "backend", Flavors: []string{}, SkipSettings: true}
	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "settings.json")); err == nil {
		t.Error("settings.json should not be written when SkipSettings is true")
	}
}

func TestComposeSettingsInjectsPermissionMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	settingsDir := filepath.Join(src, "base", ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"),
		[]byte(`{"model":"claude-sonnet-4-6"}`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	req := compose.Request{
		SourceDir:      src,
		DestDir:        dst,
		Profile:        "backend",
		Flavors:        []string{},
		PermissionMode: "bypass",
	}
	content, err := compose.ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(content, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v, want bypass", m["permissionMode"])
	}
}

func TestComposeSettingsNoPermissionModeWhenEmpty(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	settingsDir := filepath.Join(src, "base", ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"),
		[]byte(`{"model":"claude-sonnet-4-6"}`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	req := compose.Request{SourceDir: src, DestDir: dst, Profile: "backend", Flavors: []string{}}
	content, err := compose.ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}

	if strings.Contains(string(content), "permissionMode") {
		t.Error("permissionMode should not appear when PermissionMode is empty")
	}
}

func TestComposeNoResearchFlavorAppended(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "# Base\n## Recherche-Pflicht beim Planen\nPflicht.")
	writeFile(t, src, "base/.claude/settings.json", `{"permissions":{"allow":["Bash"]}}`)
	writeFile(t, src, "profiles/backend/CLAUDE.md", "## Backend")
	writeFile(t, src, "flavors/no-research/CLAUDE.md", "## No-Research-Flavor (Opt-out)\nDeaktiviert.")

	req := compose.Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "backend",
		Flavors:   []string{"no-research"},
	}
	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Recherche-Pflicht beim Planen") {
		t.Errorf("expected base research mandate in output, got: %s", content)
	}
	if !strings.Contains(content, "No-Research-Flavor (Opt-out)") {
		t.Errorf("expected no-research opt-out block in output, got: %s", content)
	}
}


func TestComposeFullstackExtendsBackendAndFrontend(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "# Base")
	writeFile(t, src, "base/.claude/settings.json", `{"permissions":{"allow":[]}}`)
	writeFile(t, src, "profiles/backend/CLAUDE.md", "## Backend-Profil")
	writeFile(t, src, "profiles/frontend/CLAUDE.md", "## Frontend-Profil")
	writeFile(t, src, "profiles/fullstack/CLAUDE.md", "## Fullstack-Profil")
	writeFile(t, src, "profiles/fullstack/profile.yaml", "extends:\n  - backend\n  - frontend\n")

	req := compose.Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "fullstack",
		Flavors:   []string{},
	}
	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md missing: %v", err)
	}
	content := string(data)
	for _, want := range []string{"# Base", "## Backend-Profil", "## Frontend-Profil", "## Fullstack-Profil"} {
		if !strings.Contains(content, want) {
			t.Errorf("CLAUDE.md fehlt %q", want)
		}
	}
}

func TestComposeFullstackExtendsSkills(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "# Base")
	writeFile(t, src, "base/.claude/settings.json", `{"permissions":{"allow":[]}}`)
	writeFile(t, src, "profiles/backend/.claude/commands/db-migration.md", "# DB Migration")
	writeFile(t, src, "profiles/frontend/.claude/commands/a11y-audit.md", "# A11y Audit")
	writeFile(t, src, "profiles/fullstack/profile.yaml", "extends:\n  - backend\n  - frontend\n")

	req := compose.Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "fullstack",
		Flavors:   []string{},
	}
	if err := compose.Run(req); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, skill := range []string{"db-migration.md", "a11y-audit.md"} {
		if _, err := os.Stat(filepath.Join(dst, ".claude", "commands", skill)); err != nil {
			t.Errorf("skill %s fehlt: %v", skill, err)
		}
	}
}

func TestComposeSettingsMergesFlavorSettings(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/.claude/settings.json", `{"model":"claude-sonnet-4-6"}`)
	writeFile(t, src, "flavors/myflavor/.claude/settings.json", `{"hooks":{"PostToolUse":[{"hooks":[{"type":"command","command":"echo done"}]}]}}`)

	req := compose.Request{SourceDir: src, DestDir: dst, Profile: "backend", Flavors: []string{"myflavor"}}
	content, err := compose.ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}
	if !strings.Contains(string(content), "PostToolUse") {
		t.Errorf("flavor settings not merged: PostToolUse missing in %s", content)
	}
	if !strings.Contains(string(content), "claude-sonnet-4-6") {
		t.Errorf("base settings lost after flavor merge: model missing in %s", content)
	}
}

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
