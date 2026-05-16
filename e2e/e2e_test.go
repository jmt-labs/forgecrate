package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/claude-setup/internal/config"
	"github.com/jmt-labs/claude-setup/internal/deploy"
)

// localSource gibt den Pfad zum lokalen claude-setup Repo zurück.
func localSource(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(wd) // e2e/ ist ein Level unterhalb des Repo-Root
}

func TestInitCreatesExpectedFiles(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd"},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	requiredFiles := []string{
		".claude-setup.yaml",
		"CLAUDE.md",
		"AGENTS.md",
		".claude/settings.json",
		".claude/hooks/prompt-submit.sh",
		".claude/hooks/pre-tool.sh",
	}

	for _, f := range requiredFiles {
		if _, err := os.Stat(filepath.Join(dst, f)); err != nil {
			t.Errorf("missing: %s", f)
		}
	}
}

func TestInitIsIdempotent(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("first run: %v", err)
	}

	claudeMD := filepath.Join(dst, "CLAUDE.md")
	content, _ := os.ReadFile(claudeMD)
	customized := strings.Replace(string(content),
		"<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->",
		"<!-- CUSTOM:BEGIN -->\n# Mein Custom\n<!-- CUSTOM:END -->", 1)
	os.WriteFile(claudeMD, []byte(customized), 0644)

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("second run: %v", err)
	}

	after, _ := os.ReadFile(claudeMD)
	if !strings.Contains(string(after), "# Mein Custom") {
		t.Error("custom section was lost on second run")
	}
}

func TestUpdatePreservesOverrides(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("init: %v", err)
	}

	overridesDir := filepath.Join(dst, ".claude", "commands", "overrides")
	os.MkdirAll(overridesDir, 0755)
	overrideSkill := filepath.Join(overridesDir, "my-skill.md")
	os.WriteFile(overrideSkill, []byte("# My Skill"), 0644)

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("update: %v", err)
	}

	content, err := os.ReadFile(overrideSkill)
	if err != nil {
		t.Fatalf("override skill missing after update")
	}
	if string(content) != "# My Skill" {
		t.Errorf("override was modified: %q", content)
	}
}

func TestProfileSwitch(t *testing.T) {
	dst := t.TempDir()

	cfg := config.Config{Version: "1.0", Source: "github.com/jmt-labs/claude-setup", Ref: "main", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("init backend: %v", err)
	}

	cfg.Profile = "frontend"
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("update to frontend: %v", err)
	}

	yamlContent, _ := os.ReadFile(filepath.Join(dst, ".claude-setup.yaml"))
	if !strings.Contains(string(yamlContent), "frontend") {
		t.Error("profile not updated in .claude-setup.yaml")
	}

	claudeMD, _ := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if !strings.Contains(string(claudeMD), "Frontend-Profil") {
		t.Error("frontend profile content missing in CLAUDE.md")
	}
}

func TestDeployIncludesBaseSkills(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	skills := []string{"release", "repo-onboarding", "repo-health", "claude-setup-advisor"}
	for _, s := range skills {
		path := filepath.Join(dst, ".claude", "skills", s, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("base skill missing: %s", s)
		}
	}
}

func TestDeployIncludesProfileSkill(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "frontend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "skills", "accessibility-audit", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("frontend profile skill missing: accessibility-audit")
	}
}

func TestDeployIncludesFlavorSkill(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"github"},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "skills", "github-release", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("github flavor skill missing: github-release")
	}
}

func TestBaseCommandsDeployed(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	baseCommands := []string{
		"claude-setup-advisor.md",
		"claude-setup-release.md",
		"claude-setup-repo-health.md",
		"claude-setup-repo-onboarding.md",
	}

	for _, f := range baseCommands {
		path := filepath.Join(dst, ".claude", "commands", f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing base command: %s", f)
		}
	}
}

func TestProfileFlavorCommandsDeployed(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd", "strict-review"},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	expectedCommands := []string{
		"claude-setup-db-migration.md",
		"claude-setup-test-coverage.md",
		"claude-setup-pr-checklist.md",
	}

	for _, f := range expectedCommands {
		path := filepath.Join(dst, ".claude", "commands", f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing command: %s", f)
		}
	}

	// frontend-Command darf bei backend-Profil nicht vorhanden sein
	frontendOnly := filepath.Join(dst, ".claude", "commands", "claude-setup-accessibility-audit.md")
	if _, err := os.Stat(frontendOnly); err == nil {
		t.Error("frontend-only command should not be present for backend profile")
	}
}

func TestConflictDetectionTracksHashes(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	// Erster Deploy — befüllt deployed_files
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	yamlPath := filepath.Join(dst, ".claude-setup.yaml")
	written, err := config.Read(yamlPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if written.DeployedFiles[".claude/settings.json"] == "" {
		t.Error("settings.json hash not tracked after first deploy")
	}
	if written.DeployedFiles[".claude/hooks/prompt-submit.sh"] == "" {
		t.Error("prompt-submit.sh hash not tracked after first deploy")
	}

	// Zweiter Deploy ohne Änderung — Hashes bleiben stabil
	if err := deploy.Run(localSource(t), dst, written); err != nil {
		t.Fatalf("second deploy: %v", err)
	}
	after, _ := config.Read(yamlPath)
	if after.DeployedFiles[".claude/settings.json"] != written.DeployedFiles[".claude/settings.json"] {
		t.Error("hash changed on clean second deploy")
	}
}


func TestDeployIncludesGetbetterFlavorSkill(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"getbetter"},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "skills", "getbetter", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("getbetter flavor skill missing: getbetter")
	}
}

func TestDeployIncludesGetbetterCommand(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/claude-setup",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"getbetter"},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "commands", "claude-setup-getbetter.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("getbetter command missing: claude-setup-getbetter.md")
	}
}
