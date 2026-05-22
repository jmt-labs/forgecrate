# Fullstack-Vererbung und Frontend-Plugins Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fullstack-Profil erbt automatisch Backend + Frontend via `extends`-Feld in `profile.yaml`; Frontend-Profil bekommt fünf neue Plugins mit Nutzungssituationen in CLAUDE.md.

**Architecture:** Ein neues `ProfileConfig`-Struct (im `compose`-Paket) liest `profiles/<profile>/profile.yaml` und expandiert die Layer-Reihenfolge in `compose.go` und `deploy.go`. Der `Plugin`-Typ bekommt ein optionales `method`-Feld; `install.go` wählt anhand davon zwischen `plugin marketplace add` und `plugin install --scope project`.

**Tech Stack:** Go, `gopkg.in/yaml.v3` (bereits vorhanden), YAML-Profil-Dateien.

---

## Datei-Übersicht

| Aktion | Datei | Zweck |
|---|---|---|
| Create | `internal/compose/profile.go` | `ProfileConfig` + `LoadProfileConfig` |
| Create | `internal/compose/profile_test.go` | Tests für profile.go |
| Modify | `internal/compose/compose.go` | `collectMarkdownLayers` + `composeSkills` um `extends` erweitern |
| Modify | `internal/compose/compose_test.go` | Test: fullstack erbt backend+frontend Layer |
| Modify | `internal/deploy/deploy.go` | `installExtensions` + `copySkills` um `extends` erweitern |
| Modify | `internal/deploy/deploy_test.go` | Test: fullstack erbt extensions via extends |
| Modify | `internal/extensions/extensions.go` | `Method`-Feld zu `Plugin` hinzufügen |
| Modify | `internal/extensions/install.go` | Marketplace-Zweig für `method: marketplace` |
| Modify | `internal/extensions/install_test.go` | Test: marketplace-Methode |
| Create | `profiles/fullstack/profile.yaml` | `extends: [backend, frontend]` |
| Modify | `profiles/frontend/extensions.yaml` | Fünf neue Plugins |
| Modify | `profiles/frontend/CLAUDE.md` | Plugin-Nutzungssituationen |
| Modify | `profiles/fullstack/CLAUDE.md` | Playwright-Abschnitt auf fullstack-spezifisch kürzen |

---

## Task 1: `ProfileConfig` und `LoadProfileConfig`

**Files:**
- Create: `internal/compose/profile.go`
- Create: `internal/compose/profile_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

```go
// internal/compose/profile_test.go
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
```

- [ ] **Schritt 2: Test ausführen (muss fehlschlagen)**

```bash
go test ./internal/compose/ -run TestLoadProfileConfig -v
```

Erwartung: FAIL — `compose.LoadProfileConfig` und `compose.ProfileConfig` nicht definiert.

- [ ] **Schritt 3: Implementation schreiben**

```go
// internal/compose/profile.go
package compose

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProfileConfig struct {
	Extends []string `yaml:"extends"`
}

func LoadProfileConfig(sourceDir, profile string) ProfileConfig {
	path := filepath.Join(sourceDir, "profiles", profile, "profile.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return ProfileConfig{}
	}
	var cfg ProfileConfig
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}
```

- [ ] **Schritt 4: Test ausführen (muss bestehen)**

```bash
go test ./internal/compose/ -run TestLoadProfileConfig -v
```

Erwartung: PASS

- [ ] **Schritt 5: Commit**

```bash
git add internal/compose/profile.go internal/compose/profile_test.go
git commit -m "feat(compose): ProfileConfig mit extends-Feld und LoadProfileConfig"
```

---

## Task 2: `collectMarkdownLayers` und `composeSkills` um `extends` erweitern

**Files:**
- Modify: `internal/compose/compose.go:61-155`
- Modify: `internal/compose/compose_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

Füge folgenden Test in `internal/compose/compose_test.go` ein (vor der `writeFile`-Hilfsfunktion):

```go
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
```

- [ ] **Schritt 2: Tests ausführen (müssen fehlschlagen)**

```bash
go test ./internal/compose/ -run TestComposeFullstack -v
```

Erwartung: FAIL — Backend- und Frontend-Layer fehlen.

- [ ] **Schritt 3: `collectMarkdownLayers` anpassen**

Ersetze die Funktion `collectMarkdownLayers` in `internal/compose/compose.go`:

```go
func collectMarkdownLayers(req Request, filename string) []string {
	var layers []string
	profileCfg := LoadProfileConfig(req.SourceDir, req.Profile)

	candidates := []string{
		filepath.Join(req.SourceDir, "base", filename),
	}
	for _, parent := range profileCfg.Extends {
		candidates = append(candidates, filepath.Join(req.SourceDir, "profiles", parent, filename))
	}
	candidates = append(candidates, filepath.Join(req.SourceDir, "profiles", req.Profile, filename))
	for _, flavor := range req.Flavors {
		candidates = append(candidates, filepath.Join(req.SourceDir, "flavors", flavor, filename))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			layers = append(layers, string(data))
		}
	}
	return layers
}
```

- [ ] **Schritt 4: `composeSkills` anpassen**

Ersetze die Funktion `composeSkills` in `internal/compose/compose.go`:

```go
func composeSkills(req Request) error {
	skillsDest := filepath.Join(req.DestDir, ".claude", "commands")
	if err := os.MkdirAll(skillsDest, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	profileCfg := LoadProfileConfig(req.SourceDir, req.Profile)

	srcDirs := []string{
		filepath.Join(req.SourceDir, "base", ".claude", "commands"),
	}
	for _, parent := range profileCfg.Extends {
		srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "profiles", parent, ".claude", "commands"))
	}
	srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "commands"))
	for _, flavor := range req.Flavors {
		srcDirs = append(srcDirs, filepath.Join(req.SourceDir, "flavors", flavor, ".claude", "commands"))
	}

	var existing []string
	for _, d := range srcDirs {
		if _, err := os.Stat(d); err == nil {
			existing = append(existing, d)
		}
	}
	return MergeSkills(existing, skillsDest)
}
```

- [ ] **Schritt 5: Tests ausführen (müssen bestehen)**

```bash
go test ./internal/compose/ -v
```

Erwartung: alle Tests PASS

- [ ] **Schritt 6: Commit**

```bash
git add internal/compose/compose.go internal/compose/compose_test.go
git commit -m "feat(compose): extends-Vererbung in collectMarkdownLayers und composeSkills"
```

---

## Task 3: `deploy.go` um `extends` erweitern

**Files:**
- Modify: `internal/deploy/deploy.go:101-155`
- Modify: `internal/deploy/deploy_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

Füge folgenden Test in `internal/deploy/deploy_test.go` ein (vor der `writeFile`-Hilfsfunktion):

```go
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
```

- [ ] **Schritt 2: Test ausführen (muss fehlschlagen)**

```bash
go test ./internal/deploy/ -run TestRunInstallsExtensionsFromExtendsProfiles -v
```

Erwartung: FAIL — Backend- und Frontend-Extensions nicht installiert.

- [ ] **Schritt 3: `installExtensions` anpassen**

Ersetze die Funktion `installExtensions` in `internal/deploy/deploy.go`:

```go
func installExtensions(sourceDir, destDir string, cfg config.Config, claudeBin string, out io.Writer) error {
	profileCfg := compose.LoadProfileConfig(sourceDir, cfg.Profile)

	paths := []string{
		filepath.Join(sourceDir, "base", "extensions.yaml"),
	}
	for _, parent := range profileCfg.Extends {
		paths = append(paths, filepath.Join(sourceDir, "profiles", parent, "extensions.yaml"))
	}
	paths = append(paths, filepath.Join(sourceDir, "profiles", cfg.Profile, "extensions.yaml"))
	for _, flavor := range cfg.Flavors {
		paths = append(paths, filepath.Join(sourceDir, "flavors", flavor, "extensions.yaml"))
	}

	var layers []extensions.Extensions
	for _, path := range paths {
		ext, err := extensions.Load(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		layers = append(layers, ext)
	}

	merged := extensions.Merge(layers)
	if err := extensions.WriteMCPJson(destDir, merged); err != nil {
		return fmt.Errorf("write .mcp.json: %w", err)
	}
	return extensions.Installer{Claude: claudeBin, Dir: destDir, Out: out}.Install(merged)
}
```

- [ ] **Schritt 4: `copySkills` anpassen**

Lese zunächst den vollständigen Inhalt von `copySkills` in `deploy.go` (ab Zeile 129) und ersetze den Beginn der Funktion bis zur `seen`-Map:

```go
func copySkills(sourceDir, destDir string, cfg config.Config, out io.Writer) error {
	profileCfg := compose.LoadProfileConfig(sourceDir, cfg.Profile)

	dirs := []string{
		filepath.Join(sourceDir, "base", "skills"),
	}
	for _, parent := range profileCfg.Extends {
		dirs = append(dirs, filepath.Join(sourceDir, "profiles", parent, "skills"))
	}
	dirs = append(dirs, filepath.Join(sourceDir, "profiles", cfg.Profile, "skills"))
	for _, flavor := range cfg.Flavors {
		dirs = append(dirs, filepath.Join(sourceDir, "flavors", flavor, "skills"))
	}
	// Rest der Funktion bleibt unverändert (seen-Map, ReadDir-Schleife, etc.)
```

- [ ] **Schritt 5: Tests ausführen (müssen bestehen)**

```bash
go test ./internal/deploy/ -v
```

Erwartung: alle Tests PASS

- [ ] **Schritt 6: Commit**

```bash
git add internal/deploy/deploy.go internal/deploy/deploy_test.go
git commit -m "feat(deploy): extends-Vererbung in installExtensions und copySkills"
```

---

## Task 4: `method`-Feld für Marketplace-Plugins

**Files:**
- Modify: `internal/extensions/extensions.go:9-12`
- Modify: `internal/extensions/install.go:30-43`
- Modify: `internal/extensions/install_test.go`

- [ ] **Schritt 1: Failing Test schreiben**

Füge folgenden Test in `internal/extensions/install_test.go` ein:

```go
func TestInstallerMarketplacePlugin(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "ui-ux-pro-max-skill", Source: "nextlevelbuilder/ui-ux-pro-max-skill", Method: "marketplace"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin marketplace add nextlevelbuilder/ui-ux-pro-max-skill") {
		t.Errorf("expected marketplace add call, got: %q", string(data))
	}
}

func TestInstallerDefaultMethodUsesInstall(t *testing.T) {
	claude, argsFile := fakeClaude(t)

	inst := extensions.Installer{Claude: claude}
	ext := extensions.Extensions{
		Plugins: []extensions.Plugin{
			{Name: "refactoring-ui-skill", Source: "https://github.com/LovroPodobnik/refactoring-ui-skill"},
		},
	}

	if err := inst.Install(ext); err != nil {
		t.Fatalf("Install: %v", err)
	}

	data, _ := os.ReadFile(argsFile)
	if !strings.Contains(string(data), "plugin install --scope project") {
		t.Errorf("expected project-scoped install, got: %q", string(data))
	}
}
```

- [ ] **Schritt 2: Tests ausführen (müssen fehlschlagen)**

```bash
go test ./internal/extensions/ -run TestInstallerMarketplace -v
go test ./internal/extensions/ -run TestInstallerDefaultMethod -v
```

Erwartung: FAIL — `Method`-Feld nicht definiert.

- [ ] **Schritt 3: `Method`-Feld zu `Plugin` hinzufügen**

Ersetze den `Plugin`-Struct in `internal/extensions/extensions.go`:

```go
type Plugin struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	Method string `yaml:"method"` // "marketplace" → plugin marketplace add; default → plugin install --scope project
}
```

- [ ] **Schritt 4: `install.go` anpassen**

Ersetze die Plugin-Schleife in `internal/extensions/install.go`:

```go
for _, p := range ext.Plugins {
	var cmd *exec.Cmd
	if p.Method == "marketplace" {
		cmd = exec.Command(claude, "plugin", "marketplace", "add", p.Source)
	} else {
		cmd = exec.Command(claude, "plugin", "install", "--scope", "project", p.Source)
	}
	cmd.Dir = i.Dir
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		msg := string(cmdOut)
		if strings.Contains(msg, "not found in any configured marketplace") {
			_, _ = fmt.Fprintf(out, "❌ plugin:%s  (not found in marketplace)\n", p.Name)
			return fmt.Errorf("plugin %s not found in marketplace: %w", p.Name, err)
		}
		_, _ = fmt.Fprintf(out, "❌ plugin:%s  (%v)\n", p.Name, err)
		return fmt.Errorf("plugin install %s: %w", p.Name, err)
	}
	_, _ = fmt.Fprintf(out, "✅ plugin:%s\n", p.Name)
}
```

- [ ] **Schritt 5: Tests ausführen (müssen bestehen)**

```bash
go test ./internal/extensions/ -v
```

Erwartung: alle Tests PASS

- [ ] **Schritt 6: Commit**

```bash
git add internal/extensions/extensions.go internal/extensions/install.go internal/extensions/install_test.go
git commit -m "feat(extensions): method-Feld für marketplace vs install Plugin-Installations-Methode"
```

---

## Task 5: `profiles/fullstack/profile.yaml` anlegen

**Files:**
- Create: `profiles/fullstack/profile.yaml`

- [ ] **Schritt 1: Datei anlegen**

```yaml
# profiles/fullstack/profile.yaml
extends:
  - backend
  - frontend
```

- [ ] **Schritt 2: Manuell prüfen**

```bash
go run ./cmd/forgecrate --help
```

Erwartung: kein Fehler, CLI startet normal.

- [ ] **Schritt 3: Commit**

```bash
git add profiles/fullstack/profile.yaml
git commit -m "feat(profiles): fullstack erbt backend und frontend via extends"
```

---

## Task 6: Frontend-Plugins und CLAUDE.md aktualisieren

**Files:**
- Modify: `profiles/frontend/extensions.yaml`
- Modify: `profiles/frontend/CLAUDE.md`

- [ ] **Schritt 1: `profiles/frontend/extensions.yaml` aktualisieren**

Ersetze den gesamten Inhalt:

```yaml
plugins:
  - name: frontend-design
    source: frontend-design

  - name: typescript-lsp
    source: typescript-lsp

  - name: playwright
    source: playwright

  - name: ui-ux-pro-max-skill
    source: nextlevelbuilder/ui-ux-pro-max-skill
    method: marketplace

  - name: interface-design
    source: Dammyjay93/interface-design
    method: marketplace

  - name: agent-skills
    source: vercel-labs/agent-skills
    method: marketplace

  - name: wondelai-skills
    source: wondelai/skills
    method: marketplace

  - name: refactoring-ui-skill
    source: https://github.com/LovroPodobnik/refactoring-ui-skill

mcp:
  - name: playwright
    command: npx
    args: ["-y", "@playwright/mcp"]
```

- [ ] **Schritt 2: `profiles/frontend/CLAUDE.md` aktualisieren**

Ersetze den gesamten Inhalt:

```markdown
## Frontend-Profil

- Komponenten: klein, fokussiert, eine Verantwortlichkeit
- State: lokal wenn möglich, global nur wenn nötig
- Kein CSS-in-JS ohne explizite Anforderung
- Barrierefreiheit: semantisches HTML, ARIA-Attribute wo nötig
- Tests: Behavior-Tests (was der Nutzer sieht), keine Implementierungsdetails

## UI-Reviews

- **`accessibility-audit`** — schnelle statische A11y-Checks pro geänderter Datei (alt, label, aria-*). Eignet sich für Pre-Commit / PR-Reviews.
- **`ui-ux-audit`** — tiefgehender Audit der gesamten UI, gruppiert nach Bereichen, mit Severity-Bewertung und automatischer Erstellung kleinteiliger GitHub-Issues. Für Major-Releases oder größere UI-Refactorings.

## Design-Plugins

| Plugin | Optimal wenn… |
|---|---|
| `ui-ux-pro-max-skill` | Neue Komponente oder Seite designen — generiert automatisch ein vollständiges Design-System (Farben, Typografie, Spacing) passend zum Produkt; unterstützt React, Next.js, Vue, Tailwind, Flutter u.v.m. |
| `interface-design` | UI über mehrere Sessions konsistent halten — speichert Design-Entscheidungen (Spacing, Elevation, Farben) in `.interface-design/system.md` und wendet sie session-übergreifend an |
| `refactoring-ui-skill` | Bestehende UI überarbeiten — `/ui-refactor` verbessert Hierarchie, Spacing (8px-Raster), HSL-Farben und Schatten nach Refactoring-UI-Prinzipien |
| `agent-skills` | Vercel-Deployments oder React Composition Patterns — auto-detects 40+ Frameworks, hilft bei Compound Components, State-Lifting und Edge-Funktionen |
| `wondelai-skills` | UX-Strategie und Produktentscheidungen — 25 Skills nach Norman, Cialdini, Ries; deckt UX Design, Conversion-Optimierung und Produktstrategie ab |

## Playwright MCP

Browser-Automatisierung direkt aus Claude heraus. Automatisch konfiguriert via `profiles/frontend/extensions.yaml`.

**Verwende es für:** UI-Tests, Screenshots, Formular-Interaktionen, visuelle Regressionstests, Debugging von Rendering-Problemen.

**Verwende es NICHT für:** API-Tests ohne UI-Beteiligung (→ direkte HTTP-Calls), GitHub-Operationen (→ github MCP).
```

- [ ] **Schritt 3: Commit**

```bash
git add profiles/frontend/extensions.yaml profiles/frontend/CLAUDE.md
git commit -m "feat(profiles/frontend): fünf neue Design-Plugins mit Nutzungssituationen"
```

---

## Task 7: Abschluss-Validierung

- [ ] **Schritt 1: Alle Tests ausführen**

```bash
make test
```

Erwartung: alle Tests PASS, kein Fehler.

- [ ] **Schritt 2: Build prüfen**

```bash
make build
```

Erwartung: Binary wird ohne Fehler gebaut.

- [ ] **Schritt 3: Quality-Check**

```bash
make quality
```

Erwartung: `go vet` und `go build` ohne Fehler.

- [ ] **Schritt 4: Manueller Smoke-Test für `extends`**

Überprüfe, dass `profiles/fullstack/profile.yaml` korrekt geladen wird:

```bash
go test ./internal/compose/ -run TestLoadProfileConfig -v
go test ./internal/compose/ -run TestComposeFullstack -v
go test ./internal/deploy/ -run TestRunInstallsExtensionsFromExtendsProfiles -v
```

Erwartung: alle PASS.
