# codegraph-Flavor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Neuen Flavor `codegraph` implementieren, der den codegraph-MCP-Server in Ziel-Repos integriert — mit automatischer Initialisierung, Pflicht-Regeln in CLAUDE.md, Session-Hook und Post-Commit-Sync.

**Architecture:** Deploy-Logik wird um drei Mechanismen erweitert: (1) Flavor-spezifische Hooks werden in `.claude/hooks/` kopiert, (2) Flavor-eigene `.claude/settings.json`-Overlays werden in `ComposeSettings` gemergt, (3) eine neue `appendGitignore`-Funktion trägt Flavor-Gitignore-Einträge idempotent ein. Der `prompt-submit.sh`-Hook ruft `session-start.sh` automatisch auf, falls vorhanden. Issues #87 + #88.

**Tech Stack:** Go 1.22+, Shell (sh/bash), JSON (settings.json), YAML (extensions.yaml), Markdown (CLAUDE.md/SKILL.md)

**Referenzen:** Issue #85 (Epic), #86 (Recherche, closed), #87 (Flavor), #88 (Skill)

---

## Datei-Übersicht

| Datei | Aktion |
|---|---|
| `internal/deploy/deploy.go` | Modify: `copyHooks` + neue `appendGitignore` + `Run`-Aufruf |
| `internal/deploy/deploy_test.go` | Modify: Tests für Flavor-Hooks und Gitignore |
| `internal/compose/compose.go` | Modify: `ComposeSettings` mergt auch Flavor-Settings |
| `internal/compose/compose_test.go` | Modify: Test für Flavor-Settings-Merge |
| `base/hooks/prompt-submit.sh` | Modify: Auto-call `session-start.sh` wenn vorhanden |
| `flavors/codegraph/extensions.yaml` | Create: MCP-Server-Eintrag |
| `flavors/codegraph/CLAUDE.md` | Create: Pflicht-Regeln für 7 codegraph-Tools |
| `flavors/codegraph/hooks/session-start.sh` | Create: Auto-Init codegraph |
| `flavors/codegraph/hooks/post-commit.sh` | Create: Index-Sync nach git commit |
| `flavors/codegraph/.claude/settings.json` | Create: PostToolUse-Hook-Registrierung |
| `flavors/codegraph/gitignore.txt` | Create: `.codegraph/` Eintrag |
| `base/skills/forgecrate-repo-onboarding/SKILL.md` | Modify: codegraph-Abschnitt |

---

## Task 1: Flavor-Hooks in deploy.go kopieren (RED)

**Files:**
- Modify: `internal/deploy/deploy_test.go`
- Modify: `internal/deploy/deploy.go`

- [ ] **Step 1.1: Test schreiben (muss fehlschlagen)**

In `internal/deploy/deploy_test.go` nach `TestCopyHooksMissingDirSucceedsWithoutHookFiles` einfügen:

```go
func TestCopyFlavorHooksDeployedToHooksDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "flavors/myflavor/hooks/session-start.sh", "#!/bin/sh\necho hello")

	cfg := config.Config{Profile: "backend", Flavors: []string{"myflavor"}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	path := filepath.Join(dst, ".claude", "hooks", "session-start.sh")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("flavor hook not copied to .claude/hooks/: %v", err)
	}
}
```

- [ ] **Step 1.2: Test ausführen — muss FEHLSCHLAGEN**

```bash
go test ./internal/deploy/... -run TestCopyFlavorHooksDeployedToHooksDir -v
```

Erwartet: `FAIL — flavor hook not copied`

- [ ] **Step 1.3: Implementierung in `deploy.go`**

In `copyHooks` nach dem `filepath.Walk(hooksDir, ...)` Block die Flavor-Hooks hinzufügen:

```go
func copyHooks(src, dst string, cfg *config.Config, out io.Writer, in io.Reader) error {
	hooksDir := filepath.Join(src, "base", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(out, "🔵 hooks: kein Verzeichnis vorhanden, wird übersprungen\n")
			return nil
		}
		return fmt.Errorf("hooks-Verzeichnis prüfen (%s): %w", hooksDir, err)
	}

	dstHooks := filepath.Join(dst, ".claude", "hooks")
	if err := os.MkdirAll(dstHooks, 0755); err != nil {
		return fmt.Errorf("mkdir hooks: %w", err)
	}

	if err := walkHooksDir(hooksDir, dstHooks, cfg, out, in); err != nil {
		return err
	}

	for _, flavor := range cfg.Flavors {
		flavorHooksDir := filepath.Join(src, "flavors", flavor, "hooks")
		if _, err := os.Stat(flavorHooksDir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("flavor-hooks-Verzeichnis prüfen (%s): %w", flavorHooksDir, err)
		}
		if err := walkHooksDir(flavorHooksDir, dstHooks, cfg, out, in); err != nil {
			return err
		}
	}
	return nil
}

func walkHooksDir(hooksDir, dstHooks string, cfg *config.Config, out io.Writer, in io.Reader) error {
	return filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(hooksDir, path)
		dstPath := filepath.Join(dstHooks, rel)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read hook %s: %w", rel, err)
		}

		relKey := filepath.Join(".claude", "hooks", rel)
		if err := deployFile(dstPath, relKey, content, cfg, out, in); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		return os.Chmod(dstPath, 0755)
	})
}
```

- [ ] **Step 1.4: Test ausführen — muss BESTEHEN**

```bash
go test ./internal/deploy/... -run TestCopyFlavorHooksDeployedToHooksDir -v
```

Erwartet: `PASS`

- [ ] **Step 1.5: Alle Tests grün**

```bash
go test ./internal/deploy/...
```

Erwartet: alle PASS

- [ ] **Step 1.6: Commit**

```bash
git add internal/deploy/deploy.go internal/deploy/deploy_test.go
git commit -m "feat(deploy): flavor-spezifische Hooks in .claude/hooks/ kopieren"
```

---

## Task 2: ComposeSettings mergt Flavor-Settings (RED)

**Files:**
- Modify: `internal/compose/compose_test.go`
- Modify: `internal/compose/compose.go`

- [ ] **Step 2.1: Test schreiben (muss fehlschlagen)**

Datei `internal/compose/compose_test.go` lesen und folgenden Test am Ende einfügen:

```go
func TestComposeSettingsMergesFlavorSettings(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	baseSettings := `{
  "hooks": {
    "UserPromptSubmit": [{"matcher":"","hooks":[{"type":"command","command":"base-hook"}]}]
  }
}`
	flavorSettings := `{
  "hooks": {
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"flavor-hook"}]}]
  }
}`
	writeTestFile(t, src, "base/.claude/settings.json", baseSettings)
	writeTestFile(t, src, "flavors/myflavor/.claude/settings.json", flavorSettings)

	req := Request{
		SourceDir: src,
		DestDir:   dst,
		Profile:   "backend",
		Flavors:   []string{"myflavor"},
	}
	out, err := ComposeSettings(req)
	if err != nil {
		t.Fatalf("ComposeSettings: %v", err)
	}
	content := string(out)
	if !strings.Contains(content, "UserPromptSubmit") {
		t.Errorf("base UserPromptSubmit hook lost after flavor merge: %s", content)
	}
	if !strings.Contains(content, "PostToolUse") {
		t.Errorf("flavor PostToolUse hook not merged: %s", content)
	}
	if !strings.Contains(content, "base-hook") {
		t.Errorf("base-hook command lost: %s", content)
	}
	if !strings.Contains(content, "flavor-hook") {
		t.Errorf("flavor-hook command not merged: %s", content)
	}
}
```

Wenn `writeTestFile` in `compose_test.go` noch nicht vorhanden ist, erst `compose_test.go` lesen und nach einem passenden Helfer suchen. Alternativ:

```go
func writeTestFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
```

- [ ] **Step 2.2: Test ausführen — muss FEHLSCHLAGEN**

```bash
go test ./internal/compose/... -run TestComposeSettingsMergesFlavorSettings -v
```

Erwartet: `FAIL — flavor PostToolUse hook not merged`

- [ ] **Step 2.3: Implementierung in `compose.go`**

In `ComposeSettings` nach dem Override-Merge (overridePath-Block) die Flavor-Settings einfügen:

```go
func ComposeSettings(req Request) ([]byte, error) {
	basePath := filepath.Join(req.SourceDir, "base", ".claude", "settings.json")
	data, err := os.ReadFile(basePath)
	if err != nil {
		return nil, err
	}
	merged := string(data)

	profilePath := filepath.Join(req.SourceDir, "profiles", req.Profile, ".claude", "settings.json")
	if override, err := os.ReadFile(profilePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return nil, err
		}
	}

	for _, flavor := range req.Flavors {
		flavorPath := filepath.Join(req.SourceDir, "flavors", flavor, ".claude", "settings.json")
		if override, err := os.ReadFile(flavorPath); err == nil {
			merged, err = DeepMergeJSON(merged, string(override))
			if err != nil {
				return nil, err
			}
		}
	}

	overridePath := filepath.Join(req.DestDir, ".claude", "overrides", "settings.override.json")
	if override, err := os.ReadFile(overridePath); err == nil {
		merged, err = DeepMergeJSON(merged, string(override))
		if err != nil {
			return nil, err
		}
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(merged), &m); err != nil {
		return nil, fmt.Errorf("merged JSON invalid: %w", err)
	}
	if req.PermissionMode != "" {
		m["permissionMode"] = req.PermissionMode
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return append(out, '\n'), nil
}
```

- [ ] **Step 2.4: Test ausführen — muss BESTEHEN**

```bash
go test ./internal/compose/... -run TestComposeSettingsMergesFlavorSettings -v
```

Erwartet: `PASS`

- [ ] **Step 2.5: Alle compose-Tests grün**

```bash
go test ./internal/compose/...
```

Erwartet: alle PASS

- [ ] **Step 2.6: Commit**

```bash
git add internal/compose/compose.go internal/compose/compose_test.go
git commit -m "feat(compose): Flavor-spezifische settings.json in ComposeSettings mergen"
```

---

## Task 3: appendGitignore in deploy.go (RED)

**Files:**
- Modify: `internal/deploy/deploy_test.go`
- Modify: `internal/deploy/deploy.go`

- [ ] **Step 3.1: Test schreiben (muss fehlschlagen)**

```go
func TestAppendGitignoreFromFlavor(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "flavors/myflavor/gitignore.txt", ".myflavor/\n")

	cfg := config.Config{Profile: "backend", Flavors: []string{"myflavor"}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dst, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	if !strings.Contains(string(content), ".myflavor/") {
		t.Errorf(".gitignore does not contain .myflavor/: %s", content)
	}
}

func TestAppendGitignoreIdempotent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "base/CLAUDE.md", "<!-- GENERATED:BEGIN -->\n# Base\n<!-- GENERATED:END -->\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n")
	writeFile(t, src, "base/.claude/settings.json", `{}`)
	writeFile(t, src, "flavors/myflavor/gitignore.txt", ".myflavor/\n")
	writeFile(t, dst, ".gitignore", ".myflavor/\n")

	cfg := config.Config{Profile: "backend", Flavors: []string{"myflavor"}}
	if err := deploy.Run(src, dst, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dst, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not found: %v", err)
	}
	count := strings.Count(string(content), ".myflavor/")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of .myflavor/, got %d: %s", count, content)
	}
}
```

- [ ] **Step 3.2: Test ausführen — muss FEHLSCHLAGEN**

```bash
go test ./internal/deploy/... -run "TestAppendGitignore" -v
```

Erwartet: `FAIL — .gitignore not created`

- [ ] **Step 3.3: Implementierung — neue `appendGitignore`-Funktion in deploy.go**

Am Ende der Datei einfügen:

```go
func appendGitignore(sourceDir, destDir string, cfg config.Config) error {
	gitignorePath := filepath.Join(destDir, ".gitignore")

	existing := ""
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existing = string(data)
	}

	var additions []string
	for _, flavor := range cfg.Flavors {
		txtPath := filepath.Join(sourceDir, "flavors", flavor, "gitignore.txt")
		data, err := os.ReadFile(txtPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s gitignore.txt: %w", flavor, err)
		}
		for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.Contains(existing, line) {
				continue
			}
			already := false
			for _, a := range additions {
				if a == line {
					already = true
					break
				}
			}
			if !already {
				additions = append(additions, line)
			}
		}
	}

	if len(additions) == 0 {
		return nil
	}

	suffix := "\n" + strings.Join(additions, "\n") + "\n"
	f, err := os.OpenFile(gitignorePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(suffix)
	return err
}
```

Dazu `"strings"` zu den Imports in `deploy.go` hinzufügen.

- [ ] **Step 3.4: `appendGitignore` in `RunWithClaude` aufrufen**

In `RunWithClaude` nach `copySkills`:

```go
if err := appendGitignore(sourceDir, destDir, cfg); err != nil {
    return fmt.Errorf("gitignore: %w", err)
}
```

- [ ] **Step 3.5: Test ausführen — muss BESTEHEN**

```bash
go test ./internal/deploy/... -run "TestAppendGitignore" -v
```

Erwartet: `PASS`

- [ ] **Step 3.6: Alle deploy-Tests grün**

```bash
go test ./internal/deploy/...
```

Erwartet: alle PASS

- [ ] **Step 3.7: Commit**

```bash
git add internal/deploy/deploy.go internal/deploy/deploy_test.go
git commit -m "feat(deploy): .gitignore-Einträge aus Flavor-gitignore.txt idempotent eintragen"
```

---

## Task 4: prompt-submit.sh — Auto-call session-start.sh

**Files:**
- Modify: `base/hooks/prompt-submit.sh`

- [ ] **Step 4.1: prompt-submit.sh anpassen**

Datei `base/hooks/prompt-submit.sh` vollständig ersetzen mit:

```bash
#!/usr/bin/env bash
# Erinnerung an Pflicht-Skills — wird bei jeder User-Nachricht ausgegeben.
# Schlank halten: nur wenige Zeilen, vollständig cached nach erster Ausführung.

if command -v forgecrate >/dev/null 2>&1; then
  forgecrate hook prompt-submit
else
  echo "## forgecrate — Aktive Konfiguration"
  echo "Profil: unbekannt | Flavors: keine"
  echo ""
  echo "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs"
  echo "Recherche beim Planen: WebSearch/context7/fetch nutzen — nicht raten"
fi

# Flavor session-start hooks auto-discovery
root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
if [ -f "$root/.claude/hooks/session-start.sh" ]; then
  bash "$root/.claude/hooks/session-start.sh" 2>/dev/null || true
fi
```

- [ ] **Step 4.2: Manuelle Verifikation**

```bash
cat base/hooks/prompt-submit.sh
```

Erwartet: Datei enthält auto-discovery Block.

- [ ] **Step 4.3: Commit**

```bash
git add base/hooks/prompt-submit.sh
git commit -m "feat(hooks): session-start.sh auto-discovery in prompt-submit.sh"
```

---

## Task 5: codegraph-Flavor-Dateien erstellen

**Files:**
- Create: `flavors/codegraph/extensions.yaml`
- Create: `flavors/codegraph/CLAUDE.md`
- Create: `flavors/codegraph/hooks/session-start.sh`
- Create: `flavors/codegraph/hooks/post-commit.sh`
- Create: `flavors/codegraph/.claude/settings.json`
- Create: `flavors/codegraph/gitignore.txt`

- [ ] **Step 5.1: `flavors/codegraph/extensions.yaml`**

```yaml
mcp:
  - name: codegraph
    type: stdio
    command: codegraph
    args: ["serve", "--mcp"]
```

- [ ] **Step 5.2: `flavors/codegraph/CLAUDE.md`**

```markdown
## codegraph — Pflicht-Regeln

- Symbolsuche: `codegraph_search` MUSS statt `grep`/`Bash` verwendet werden
- Code-Kontext lesen: `codegraph_context` MUSS statt mehreren `Read`-Aufrufen verwendet werden
- Vor Refactoring/Breaking Changes: `codegraph_impact` MUSS ausgeführt werden
- Direktes `grep` auf Quellcode ist verboten, solange codegraph initialisiert ist

**Fallback:** Meldet `codegraph_status` dass der Index nicht bereit ist → Fallback auf Read/grep erlaubt und Init-Hook anstoßen.

## codegraph — Verfügbare Tools

| Tool | Wann nutzen |
|---|---|
| `codegraph_search` | Symbole nach Name finden (statt grep) |
| `codegraph_context` | Relevanten Code für Tasks zusammentragen (statt mehrere Reads) |
| `codegraph_callers` | Wer ruft diese Funktion auf? |
| `codegraph_callees` | Was ruft diese Funktion auf? |
| `codegraph_impact` | Auswirkungsanalyse vor Refactoring / Breaking Changes |
| `codegraph_node` | Details zu einzelnen Symbolen |
| `codegraph_files` | Dateistruktur abfragen |
| `codegraph_status` | Indexgesundheit prüfen (vor erstem Einsatz) |

**Initialisierung:** `codegraph install --yes && codegraph init . --index` (wird automatisch via Hook ausgeführt)
```

- [ ] **Step 5.3: `flavors/codegraph/hooks/session-start.sh`**

```bash
#!/bin/sh
# Codegraph automatisch installieren und initialisieren wenn noch nicht vorhanden
if [ -f ".forgecrate.yaml" ] && grep -q "codegraph" ".forgecrate.yaml" 2>/dev/null; then
  if [ ! -d ".codegraph" ]; then
    echo "codegraph: Index fehlt — initialisiere..."
    codegraph install --yes 2>/dev/null || true
    codegraph init . --index
    echo "codegraph: Index bereit."
  fi
fi
```

- [ ] **Step 5.4: `flavors/codegraph/hooks/post-commit.sh`**

```bash
#!/bin/sh
# Index nach jedem Commit aktualisieren
INPUT="${TOOL_INPUT:-}"
case "$INPUT" in
  *"git commit"*)
    if [ -d ".codegraph" ]; then
      codegraph sync . 2>/dev/null || true
    fi
    ;;
esac
```

- [ ] **Step 5.5: `flavors/codegraph/.claude/settings.json`**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'root=$(git rev-parse --show-toplevel 2>/dev/null || pwd) && exec bash \"$root/.claude/hooks/post-commit.sh\"'"
          }
        ]
      }
    ]
  }
}
```

- [ ] **Step 5.6: `flavors/codegraph/gitignore.txt`**

```
.codegraph/
```

- [ ] **Step 5.7: Alle Dateien vorhanden prüfen**

```bash
find flavors/codegraph -type f | sort
```

Erwartet:
```
flavors/codegraph/.claude/settings.json
flavors/codegraph/CLAUDE.md
flavors/codegraph/extensions.yaml
flavors/codegraph/gitignore.txt
flavors/codegraph/hooks/post-commit.sh
flavors/codegraph/hooks/session-start.sh
```

- [ ] **Step 5.8: Hook-Ausführbarkeit setzen**

```bash
chmod +x flavors/codegraph/hooks/session-start.sh flavors/codegraph/hooks/post-commit.sh
```

- [ ] **Step 5.9: Alle Tests grün**

```bash
go test ./...
```

Erwartet: alle PASS

- [ ] **Step 5.10: Commit**

```bash
git add flavors/codegraph/
git commit -m "feat(flavor): codegraph-Flavor — extensions, CLAUDE.md, Hooks, gitignore"
```

---

## Task 6: forgecrate-repo-onboarding Skill erweitern (Issue #88)

**Files:**
- Modify: `base/skills/forgecrate-repo-onboarding/SKILL.md`

- [ ] **Step 6.1: SKILL.md lesen und codegraph-Abschnitt hinzufügen**

Datei `base/skills/forgecrate-repo-onboarding/SKILL.md` lesen. Am Ende (nach Schritt 5) folgenden Abschnitt ergänzen:

```markdown
6. **codegraph-Flavor prüfen** — falls in `flavors` aktiv:

   Prüfe ob `codegraph` in `.forgecrate.yaml`-Flavors enthalten ist:
   - **Flavor aktiv + `.codegraph/` fehlt:** führe direkt aus:
     ```
     codegraph install --yes && codegraph init . --index
     ```
     Ausgabe: eine Statuszeile ("codegraph: Index wird erstellt...")
   - **Flavor aktiv + `.codegraph/` vorhanden:** rufe `codegraph_status` auf und füge Ergebnis in Onboarding-Output ein (kein erneuter Init)
   - **Flavor nicht aktiv:** keine Änderung im Output
```

- [ ] **Step 6.2: Manuelle Überprüfung**

```bash
cat base/skills/forgecrate-repo-onboarding/SKILL.md
```

Erwartet: Schritt 6 mit codegraph-Logik am Ende sichtbar.

- [ ] **Step 6.3: Commit**

```bash
git add base/skills/forgecrate-repo-onboarding/SKILL.md
git commit -m "feat(skill): forgecrate-repo-onboarding um codegraph-Flavor erweitert (closes #88)"
```

---

## Task 7: Abschluss-Validierung

- [ ] **Step 7.1: Alle Tests grün**

```bash
go test ./...
```

Erwartet: alle PASS, kein FAIL.

- [ ] **Step 7.2: Build-Check**

```bash
go build ./cmd/forgecrate
```

Erwartet: kein Fehler.

- [ ] **Step 7.3: Quality-Check**

```bash
make quality
```

Erwartet: kein Fehler.

- [ ] **Step 7.4: codegraph in flavor-Liste prüfen**

```bash
ls flavors/
```

Erwartet: `codegraph` erscheint neben den anderen Flavors.

- [ ] **Step 7.5: Integrations-Test — simulate deploy**

```bash
go test ./internal/deploy/... -run "TestCopyFlavorHooks|TestAppendGitignore|TestComposeSettingsMerges" -v
```

Erwartet: alle PASS.

---

## Self-Review Checkliste

- [x] Flavor-Hooks kopieren: Task 1
- [x] ComposeSettings mergt Flavor-Settings: Task 2
- [x] `.gitignore`-Appending: Task 3
- [x] prompt-submit.sh auto-discovery: Task 4
- [x] extensions.yaml (MCP `codegraph serve --mcp`): Task 5.1
- [x] CLAUDE.md mit Pflicht-Regeln für alle 7 Tools: Task 5.2
- [x] session-start.sh Auto-Init: Task 5.3
- [x] post-commit.sh Sync: Task 5.4
- [x] `.claude/settings.json` PostToolUse-Hook: Task 5.5
- [x] `gitignore.txt` `.codegraph/`: Task 5.6
- [x] `forgecrate-repo-onboarding` Skill: Task 6
- [x] Alle Tests: Task 7
