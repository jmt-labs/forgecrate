package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/deploy"
)

// localSource gibt den Pfad zum lokalen forgecrate Repo zurück.
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
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd"},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	requiredFiles := []string{
		".forgecrate.yaml",
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
		Source:  "github.com/jmt-labs/forgecrate",
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
		Source:  "github.com/jmt-labs/forgecrate",
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

	cfg := config.Config{Version: "1.0", Source: "github.com/jmt-labs/forgecrate", Ref: "main", Profile: "backend", Flavors: []string{}}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("init backend: %v", err)
	}

	cfg.Profile = "frontend"
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("update to frontend: %v", err)
	}

	yamlContent, _ := os.ReadFile(filepath.Join(dst, ".forgecrate.yaml"))
	if !strings.Contains(string(yamlContent), "frontend") {
		t.Error("profile not updated in .forgecrate.yaml")
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
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	skills := []string{"release", "repo-onboarding", "repo-health", "forgecrate-advisor"}
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
		Source:  "github.com/jmt-labs/forgecrate",
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
		Source:  "github.com/jmt-labs/forgecrate",
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
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	baseCommands := []string{
		"forgecrate-advisor.md",
		"forgecrate-release.md",
		"forgecrate-repo-health.md",
		"forgecrate-repo-onboarding.md",
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
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"tdd", "strict-review"},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	expectedCommands := []string{
		"forgecrate-db-migration.md",
		"forgecrate-test-coverage.md",
		"forgecrate-pr-checklist.md",
	}

	for _, f := range expectedCommands {
		path := filepath.Join(dst, ".claude", "commands", f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing command: %s", f)
		}
	}

	// frontend-Command darf bei backend-Profil nicht vorhanden sein
	frontendOnly := filepath.Join(dst, ".claude", "commands", "forgecrate-accessibility-audit.md")
	if _, err := os.Stat(frontendOnly); err == nil {
		t.Error("frontend-only command should not be present for backend profile")
	}
}

func TestConflictDetectionTracksHashes(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	// Erster Deploy — befüllt deployed_files
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("first deploy: %v", err)
	}

	yamlPath := filepath.Join(dst, ".forgecrate.yaml")
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
		Source:  "github.com/jmt-labs/forgecrate",
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
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{"getbetter"},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	path := filepath.Join(dst, ".claude", "commands", "forgecrate-getbetter.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("getbetter command missing: forgecrate-getbetter.md")
	}
}

func TestHookCommandsUseAbsolutePaths(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings.json missing: %v", err)
	}
	s := string(content)
	if strings.Contains(s, `"bash .claude/hooks/`) {
		t.Error("settings.json contains relative hook path — should use git rev-parse for absolute resolution")
	}
}

func TestDeployIncludesContext7MCP(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	mcpJson, err := os.ReadFile(filepath.Join(dst, ".mcp.json"))
	if err != nil {
		t.Fatalf(".mcp.json missing: %v", err)
	}
	if !strings.Contains(string(mcpJson), `"context7"`) {
		t.Error(".mcp.json missing context7 MCP server entry")
	}
	if !strings.Contains(string(mcpJson), `@upstash/context7-mcp`) {
		t.Error(".mcp.json missing context7 npx command")
	}
}

func TestDeployIncludesPlaywrightMCPFrontend(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "frontend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	mcpJson, err := os.ReadFile(filepath.Join(dst, ".mcp.json"))
	if err != nil {
		t.Fatalf(".mcp.json missing: %v", err)
	}
	if !strings.Contains(string(mcpJson), `"playwright"`) {
		t.Error(".mcp.json missing playwright MCP server entry for frontend profile")
	}
	if !strings.Contains(string(mcpJson), `@playwright/mcp`) {
		t.Error(".mcp.json missing @playwright/mcp npx command for frontend profile")
	}
}

func TestDeployIncludesPlaywrightMCPFullstack(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "fullstack",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	mcpJson, err := os.ReadFile(filepath.Join(dst, ".mcp.json"))
	if err != nil {
		t.Fatalf(".mcp.json missing: %v", err)
	}
	if !strings.Contains(string(mcpJson), `"playwright"`) {
		t.Error(".mcp.json missing playwright MCP server entry for fullstack profile")
	}
	if !strings.Contains(string(mcpJson), `@playwright/mcp`) {
		t.Error(".mcp.json missing @playwright/mcp npx command for fullstack profile")
	}
}

func TestDeployIncludesParallelisierungSection(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}
	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dst, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md missing: %v", err)
	}
	if !strings.Contains(string(content), "Parallelisierung & Isolation") {
		t.Error("CLAUDE.md missing section: Parallelisierung & Isolation")
	}
	if !strings.Contains(string(content), "run_in_background") {
		t.Error("CLAUDE.md missing: run_in_background")
	}
	if !strings.Contains(string(content), `isolation: "worktree"`) {
		t.Error(`CLAUDE.md missing: isolation: "worktree"`)
	}
}

func TestPermissionModeInSettings(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version:        "1.0",
		Source:         "github.com/jmt-labs/forgecrate",
		Ref:            "main",
		Profile:        "backend",
		Flavors:        []string{},
		PermissionMode: "bypass",
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings.json: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["permissionMode"] != "bypass" {
		t.Errorf("permissionMode: got %v, want bypass", m["permissionMode"])
	}
}

func TestSetPermissionModeE2E(t *testing.T) {
	dst := t.TempDir()
	cfg := config.Config{
		Version: "1.0",
		Source:  "github.com/jmt-labs/forgecrate",
		Ref:     "main",
		Profile: "backend",
		Flavors: []string{},
	}

	if err := deploy.Run(localSource(t), dst, cfg); err != nil {
		t.Fatalf("deploy.Run: %v", err)
	}

	cfgPath := filepath.Join(dst, ".forgecrate.yaml")
	got, err := config.Read(cfgPath)
	if err != nil {
		t.Fatalf("config.Read: %v", err)
	}

	if err := deploy.PatchPermissionMode(dst, "plan", &got); err != nil {
		t.Fatalf("PatchPermissionMode: %v", err)
	}
	got.PermissionMode = "plan"
	if err := config.Write(cfgPath, got); err != nil {
		t.Fatalf("config.Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings after patch: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON after patch: %v", err)
	}
	if m["permissionMode"] != "plan" {
		t.Errorf("after patch permissionMode: got %v, want plan", m["permissionMode"])
	}

	got2, err := config.Read(cfgPath)
	if err != nil {
		t.Fatalf("config.Read after patch: %v", err)
	}
	if err := deploy.Run(localSource(t), dst, got2); err != nil {
		t.Fatalf("second deploy.Run: %v", err)
	}
	data2, err := os.ReadFile(filepath.Join(dst, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings after redeploy: %v", err)
	}
	var m2 map[string]any
	if err := json.Unmarshal(data2, &m2); err != nil {
		t.Fatalf("invalid JSON after redeploy: %v", err)
	}
	if m2["permissionMode"] != "plan" {
		t.Errorf("after redeploy permissionMode: got %v, want plan", m2["permissionMode"])
	}
}
