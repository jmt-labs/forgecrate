package deploy_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
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
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".forgecrate.yaml")); err != nil {
		t.Errorf(".forgecrate.yaml missing")
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
	_ = os.WriteFile(fakeClaude, []byte(script), 0755)

	baseDir := filepath.Join(src, "base")
	_ = os.MkdirAll(baseDir, 0755)
	_ = os.WriteFile(filepath.Join(baseDir, "CLAUDE.md"), []byte("<!-- GENERATED:BEGIN -->\n# Claude\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"), 0644)
	_ = os.MkdirAll(filepath.Join(baseDir, ".claude"), 0755)
	_ = os.WriteFile(filepath.Join(baseDir, ".claude", "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)
	_ = os.WriteFile(filepath.Join(baseDir, "extensions.yaml"), []byte("plugins:\n  - name: superpowers\n    source: claude-plugins-official/superpowers\n"), 0644)

	cfg := config.Config{Profile: "backend"}
	if err := deploy.RunWithClaude(src, dst, cfg, fakeClaude, io.Discard, strings.NewReader("")); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install --scope project claude-plugins-official/superpowers") {
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

func setupMinimalSource(t *testing.T, src string) {
	t.Helper()
	settingsDir := filepath.Join(src, "base", ".claude")
	_ = os.MkdirAll(settingsDir, 0755)
	_ = os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(`{"model":"claude-sonnet-4-6"}`), 0644)
	_ = os.MkdirAll(filepath.Join(src, "base"), 0755)
	_ = os.WriteFile(filepath.Join(src, "base", "CLAUDE.md"), []byte("# Base\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"), 0644)
}

func TestDeployTracksSettingsHash(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	setupMinimalSource(t, src)

	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	written, err := config.Read(filepath.Join(dst, ".forgecrate.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if written.DeployedFiles[".claude/settings.json"] == "" {
		t.Error("settings.json hash not tracked after deploy")
	}
}

func TestDeploySecondRunIsStable(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	setupMinimalSource(t, src)

	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	written, _ := config.Read(filepath.Join(dst, ".forgecrate.yaml"))
	hashBefore := written.DeployedFiles[".claude/settings.json"]

	if err := deploy.Run(src, dst, written); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	after, _ := config.Read(filepath.Join(dst, ".forgecrate.yaml"))
	if after.DeployedFiles[".claude/settings.json"] != hashBefore {
		t.Error("hash changed on clean second deploy — should be stable")
	}
}

func TestCopyHooksMissingDirSucceedsWithoutHookFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// settings.json referenziert Hooks, aber base/hooks/ fehlt absichtlich
	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"bash .claude/hooks/pre-tool.sh"}]}]}}`)

	cfg := config.Config{Profile: "backend"}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run should succeed even when hooks dir missing: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "hooks", "pre-tool.sh")); err == nil {
		t.Error("pre-tool.sh should not exist when source has no hooks dir")
	}
}

func TestDeployConflictIsShown(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	setupMinimalSource(t, src)

	// Erster Deploy: Settings werden eingespielt und Hash gespeichert
	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	// Nutzer ändert die Datei lokal
	settingsPath := filepath.Join(dst, ".claude", "settings.json")
	_ = os.WriteFile(settingsPath, []byte(`{"model":"user-modified"}`), 0644)

	// Upstream bekommt eine andere Änderung
	_ = os.WriteFile(filepath.Join(src, "base", ".claude", "settings.json"), []byte(`{"model":"upstream-update"}`), 0644)

	// Zweiter Deploy: Konflikt erwartet — Nutzer wählt "behalten"
	cfg2, _ := config.Read(filepath.Join(dst, ".forgecrate.yaml"))
	var out strings.Builder
	if err := deploy.RunWithClaude(src, dst, cfg2, "claude", &out, strings.NewReader("b\n")); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "KONFLIKT") {
		t.Errorf("Konfliktmeldung fehlt in der Ausgabe:\n%s", output)
	}
	if !strings.Contains(output, ".claude/settings.json") {
		t.Errorf("Dateiname fehlt in der Konfliktmeldung:\n%s", output)
	}

	// Nutzer-Version muss erhalten bleiben
	kept, _ := os.ReadFile(settingsPath)
	if string(kept) != `{"model":"user-modified"}` {
		t.Errorf("Datei hätte erhalten bleiben sollen, aber: %q", string(kept))
	}
}

func TestDeployConflictOverwriteReplacesFile(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	setupMinimalSource(t, src)

	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	settingsPath := filepath.Join(dst, ".claude", "settings.json")
	_ = os.WriteFile(settingsPath, []byte(`{"model":"user-modified"}`), 0644)
	_ = os.WriteFile(filepath.Join(src, "base", ".claude", "settings.json"), []byte(`{"model":"upstream-update"}`), 0644)

	cfg2, _ := config.Read(filepath.Join(dst, ".forgecrate.yaml"))
	if err := deploy.RunWithClaude(src, dst, cfg2, "claude", io.Discard, strings.NewReader("ü\n")); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	got, _ := os.ReadFile(settingsPath)
	content := string(got)
	if !strings.Contains(content, "upstream-update") {
		t.Errorf("expected upstream content after überschreiben, got: %q", content)
	}
	if strings.Contains(content, "user-modified") {
		t.Errorf("user content should have been replaced after überschreiben, got: %q", content)
	}
}

func TestDeployConflictEmptyInputKeepsUserFile(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	setupMinimalSource(t, src)

	cfg := config.Config{Version: "1.0", Source: "s", Ref: "r", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	settingsPath := filepath.Join(dst, ".claude", "settings.json")
	_ = os.WriteFile(settingsPath, []byte(`{"model":"user-modified"}`), 0644)
	_ = os.WriteFile(filepath.Join(src, "base", ".claude", "settings.json"), []byte(`{"model":"upstream-update"}`), 0644)

	cfg2, _ := config.Read(filepath.Join(dst, ".forgecrate.yaml"))
	if err := deploy.RunWithClaude(src, dst, cfg2, "claude", io.Discard, strings.NewReader("\n")); err != nil {
		t.Fatalf("second deploy: %v", err)
	}

	kept, _ := os.ReadFile(settingsPath)
	if string(kept) != `{"model":"user-modified"}` {
		t.Errorf("expected user content after empty input (default behalten), got: %q", string(kept))
	}
}

func TestCopyHooksStatErrorPropagates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("circular symlinks not reliable on Windows")
	}
	src := t.TempDir()
	dst := t.TempDir()
	setupMinimalSource(t, src)

	hooksPath := filepath.Join(src, "base", "hooks")
	if err := os.Symlink(hooksPath, hooksPath); err != nil {
		t.Skipf("cannot create circular symlink: %v", err)
	}

	cfg := config.Config{Profile: "backend"}
	err := deploy.Run(src, dst, cfg)
	if err == nil {
		t.Fatal("expected error when hooks dir has unresolvable stat error, got nil")
	}
	if !strings.Contains(err.Error(), "hooks-Verzeichnis") {
		t.Errorf("expected specific stat error message containing 'hooks-Verzeichnis', got: %v", err)
	}
}

func TestRunInstallsExtensionsFromExtendsProfiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script nicht unterstützt auf Windows")
	}
	src := t.TempDir()
	dst := t.TempDir()

	claudeDir := t.TempDir()
	argsFile := filepath.Join(claudeDir, "calls.txt")
	fakeClaude := filepath.Join(claudeDir, "claude")
	script := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", argsFile)
	_ = os.WriteFile(fakeClaude, []byte(script), 0755)

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{"model":"claude-sonnet-4-6"}`)
	writeFile(t, src, "profiles/backend/extensions.yaml", "plugins:\n  - name: backend-plugin\n    source: backend-plugin\n")
	writeFile(t, src, "profiles/frontend/extensions.yaml", "plugins:\n  - name: frontend-plugin\n    source: frontend-plugin\n")
	writeFile(t, src, "profiles/fullstack/profile.yaml", "extends:\n  - backend\n  - frontend\n")

	cfg := config.Config{Profile: "fullstack"}
	if err := deploy.RunWithClaude(src, dst, cfg, fakeClaude, io.Discard, strings.NewReader("")); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	calls := string(data)
	if !strings.Contains(calls, "backend-plugin") {
		t.Errorf("backend-plugin nicht installiert, calls: %q", calls)
	}
	if !strings.Contains(calls, "frontend-plugin") {
		t.Errorf("frontend-plugin nicht installiert, calls: %q", calls)
	}
}

func TestCopyFlavorHooksDeployedToHooksDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "base/hooks/pre-tool.sh", "#!/bin/sh\necho base")
	writeFile(t, src, "flavors/myflavor/hooks/session-start.sh", "#!/bin/sh\necho hello")

	cfg := config.Config{Profile: "backend", Flavors: []string{"myflavor"}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, ".claude", "hooks", "session-start.sh")); err != nil {
		t.Errorf("flavor hook not copied to .claude/hooks/: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, ".claude", "hooks", "pre-tool.sh")); err != nil {
		t.Errorf("base hook missing after flavor hooks added: %v", err)
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
