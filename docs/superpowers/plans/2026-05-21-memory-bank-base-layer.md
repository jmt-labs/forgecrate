# memory-bank base layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** memory-bank MCP-Server und Scaffold-Dateien als verpflichtenden Bestandteil des base layers einführen und den forgecrate-repo-onboarding Skill um automatisches Befüllen der memory-bank erweitern.

**Architecture:** `base/extensions.yaml` bekommt den MCP-Eintrag, `base/memory-bank/` enthält 5 Template-Dateien, `deploy.go` bekommt `scaffoldMemoryBank()` das fehlende Dateien idempotent anlegt. Der Skill schreibt direkt ohne Rückfrage.

**Tech Stack:** Go 1.24, `os`/`filepath`/`io`, YAML, Markdown

**Spec:** `docs/superpowers/specs/2026-05-21-memory-bank-base-layer-design.md`

---

## Dateiübersicht

| Datei | Aktion |
|---|---|
| `base/extensions.yaml` | Modify: memory-bank MCP-Eintrag hinzufügen |
| `base/memory-bank/projectbrief.md` | Create: Template |
| `base/memory-bank/techContext.md` | Create: Template |
| `base/memory-bank/systemPatterns.md` | Create: Template |
| `base/memory-bank/activeContext.md` | Create: Template |
| `base/memory-bank/progress.md` | Create: Template |
| `internal/deploy/deploy.go` | Modify: `scaffoldMemoryBank()` hinzufügen |
| `internal/deploy/deploy_test.go` | Create: Tests für `scaffoldMemoryBank()` |
| `base/CLAUDE.md` | Modify: Dokumentation ergänzen |
| `base/skills/forgecrate-repo-onboarding/skill.md` | Modify: memory-bank-Schritt hinzufügen |

---

### Task 1: MCP-Eintrag in `base/extensions.yaml`

**Files:**
- Modify: `base/extensions.yaml`

- [ ] **Step 1: Eintrag hinzufügen**

Am Ende des `mcp:`-Abschnitts in `base/extensions.yaml` einfügen:

```yaml
  - name: memory-bank
    command: npx
    args: ["-y", "@allpepper/memory-bank-mcp"]
    env:
      MEMORY_BANK_ROOT: "memory-bank"
```

- [ ] **Step 2: YAML-Syntax prüfen**

```bash
go run ./cmd/forgecrate describe 2>&1 | head -5
```

Erwartet: kein Parse-Fehler, normale Ausgabe.

- [ ] **Step 3: Commit**

```bash
git add base/extensions.yaml
git commit -m "feat(base): add memory-bank MCP server to extensions"
```

---

### Task 2: Template-Dateien in `base/memory-bank/`

**Files:**
- Create: `base/memory-bank/projectbrief.md`
- Create: `base/memory-bank/techContext.md`
- Create: `base/memory-bank/systemPatterns.md`
- Create: `base/memory-bank/activeContext.md`
- Create: `base/memory-bank/progress.md`

- [ ] **Step 1: Verzeichnis anlegen und Dateien erstellen**

```bash
mkdir -p base/memory-bank
```

Inhalt `base/memory-bank/projectbrief.md`:
```markdown
# Project Brief

## Was ist dieses Projekt?

<!-- Kurze Beschreibung: was tut dieses Projekt, für wen, warum. -->

## Ziele

<!-- Welche Probleme löst es? Was ist der Mehrwert? -->

## Nicht-Ziele

<!-- Was ist explizit out of scope? -->
```

Inhalt `base/memory-bank/techContext.md`:
```markdown
# Tech Context

## Stack

<!-- Programmiersprachen, Frameworks, wichtige Libraries. -->

## Tools & Infrastruktur

<!-- CI/CD, Linting, Test-Runner, Deployment. -->

## Constraints

<!-- Technische Einschränkungen, die Entscheidungen beeinflussen. -->
```

Inhalt `base/memory-bank/systemPatterns.md`:
```markdown
# System Patterns

## Architektur-Entscheidungen

<!-- Wichtige ADRs: Was wurde entschieden und warum? -->

## Wiederkehrende Muster

<!-- Patterns die im Projekt konsistent verwendet werden. -->

## Anti-Patterns

<!-- Was soll vermieden werden und warum? -->
```

Inhalt `base/memory-bank/activeContext.md`:
```markdown
# Active Context

## Aktueller Fokus

<!-- Woran wird gerade gearbeitet? Welches Feature, welcher Bug? -->

## Offene Fragen

<!-- Ungeklärte Punkte, ausstehende Entscheidungen. -->

## Bekannte Blocker

<!-- Was hält den Fortschritt auf? -->
```

Inhalt `base/memory-bank/progress.md`:
```markdown
# Progress

## Fertig

<!-- Abgeschlossene Features und Meilensteine. -->

## In Arbeit

<!-- Aktuell laufende Arbeiten. -->

## Nächste Schritte

<!-- Was kommt als nächstes dran? -->
```

- [ ] **Step 2: Dateien prüfen**

```bash
ls base/memory-bank/
```

Erwartet: `activeContext.md  progress.md  projectbrief.md  systemPatterns.md  techContext.md`

- [ ] **Step 3: Commit**

```bash
git add base/memory-bank/
git commit -m "feat(base): add memory-bank template files"
```

---

### Task 3: `scaffoldMemoryBank()` in `deploy.go` — Test zuerst

**Files:**
- Create: `internal/deploy/scaffold_test.go`
- Create: `internal/deploy/scaffold.go`

- [ ] **Step 1: Failing-Test schreiben**

Neue Datei `internal/deploy/scaffold_test.go`:

```go
package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldMemoryBank_CreatesFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Quelldatei anlegen
	srcMB := filepath.Join(src, "base", "memory-bank")
	os.MkdirAll(srcMB, 0755)
	os.WriteFile(filepath.Join(srcMB, "projectbrief.md"), []byte("# Project Brief\n"), 0644)
	os.WriteFile(filepath.Join(srcMB, "techContext.md"), []byte("# Tech Context\n"), 0644)

	if err := scaffoldMemoryBank(src, dst); err != nil {
		t.Fatalf("scaffoldMemoryBank: %v", err)
	}

	for _, name := range []string{"projectbrief.md", "techContext.md"} {
		path := filepath.Join(dst, "memory-bank", name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}
}

func TestScaffoldMemoryBank_DoesNotOverwriteExisting(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcMB := filepath.Join(src, "base", "memory-bank")
	os.MkdirAll(srcMB, 0755)
	os.WriteFile(filepath.Join(srcMB, "projectbrief.md"), []byte("template\n"), 0644)

	// Vorhandene Datei mit eigenem Inhalt
	dstMB := filepath.Join(dst, "memory-bank")
	os.MkdirAll(dstMB, 0755)
	os.WriteFile(filepath.Join(dstMB, "projectbrief.md"), []byte("custom content\n"), 0644)

	if err := scaffoldMemoryBank(src, dst); err != nil {
		t.Fatalf("scaffoldMemoryBank: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dstMB, "projectbrief.md"))
	if string(data) != "custom content\n" {
		t.Errorf("existing file must not be overwritten, got: %q", data)
	}
}

func TestScaffoldMemoryBank_NoSourceDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	// Kein base/memory-bank/ im src → kein Fehler, kein Absturz

	if err := scaffoldMemoryBank(src, dst); err != nil {
		t.Fatalf("expected no error when source missing, got: %v", err)
	}
}
```

- [ ] **Step 2: Tests ausführen — müssen fehlschlagen**

```bash
go test ./internal/deploy/... -run TestScaffoldMemoryBank -v
```

Erwartet: `FAIL` mit `undefined: scaffoldMemoryBank`

- [ ] **Step 3: Implementierung schreiben**

Neue Datei `internal/deploy/scaffold.go`:

```go
package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// scaffoldMemoryBank kopiert base/memory-bank/* nach memory-bank/ im Ziel-Repo.
// Vorhandene Dateien werden nicht überschrieben.
func scaffoldMemoryBank(sourceDir, destDir string) error {
	srcDir := filepath.Join(sourceDir, "base", "memory-bank")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	dstDir := filepath.Join(destDir, "memory-bank")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("mkdir memory-bank: %w", err)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read memory-bank source: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		dst := filepath.Join(dstDir, entry.Name())
		if _, err := os.Stat(dst); err == nil {
			continue // existiert bereits → überspringen
		}
		if err := copyFile(filepath.Join(srcDir, entry.Name()), dst); err != nil {
			return fmt.Errorf("scaffold %s: %w", entry.Name(), err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Tests ausführen — müssen bestehen**

```bash
go test ./internal/deploy/... -run TestScaffoldMemoryBank -v
```

Erwartet:
```
--- PASS: TestScaffoldMemoryBank_CreatesFiles
--- PASS: TestScaffoldMemoryBank_DoesNotOverwriteExisting
--- PASS: TestScaffoldMemoryBank_NoSourceDir
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/deploy/scaffold.go internal/deploy/scaffold_test.go
git commit -m "feat(deploy): add scaffoldMemoryBank with idempotent file copy"
```

---

### Task 4: `scaffoldMemoryBank()` in `RunWithClaude()` einbinden

**Files:**
- Modify: `internal/deploy/deploy.go:53-59`

- [ ] **Step 1: Aufruf nach `installExtensions` einfügen**

In `RunWithClaude()`, nach dem `installExtensions`-Block (Zeile ~55):

```go
	if err := installExtensions(sourceDir, destDir, cfg, claudeBin, out); err != nil {
		return fmt.Errorf("extensions: %w", err)
	}

	if err := scaffoldMemoryBank(sourceDir, destDir); err != nil {
		return fmt.Errorf("memory-bank scaffold: %w", err)
	}

	if err := copySkills(sourceDir, destDir, cfg, out); err != nil {
		return fmt.Errorf("skills: %w", err)
	}
```

- [ ] **Step 2: Alle bestehenden Tests laufen lassen**

```bash
go test ./internal/deploy/... -v 2>&1 | tail -20
```

Erwartet: alle Tests grün, kein neuer Fehler.

- [ ] **Step 3: Commit**

```bash
git add internal/deploy/deploy.go
git commit -m "feat(deploy): call scaffoldMemoryBank during deploy"
```

---

### Task 5: `base/CLAUDE.md` — Dokumentation ergänzen

**Files:**
- Modify: `base/CLAUDE.md:109` (Einleitungssatz)
- Modify: `base/CLAUDE.md:155` (nach context7-Abschnitt)

- [ ] **Step 1: Einleitungssatz anpassen**

Zeile 109 — `Vier` → `Fünf`:

```
Fünf MCP-Server sind im base layer deklariert und stehen automatisch zur Verfügung.
```

- [ ] **Step 2: Neuen Abschnitt nach dem context7-Block einfügen**

Nach Zeile 155 (nach `**Keine Konfiguration nötig** — ...`) und vor `## MCP-Konfiguration`:

```markdown
### Memory-Bank (`memory-bank`)

Strukturiertes, dateibasiertes Projektgedächtnis in `memory-bank/`. Persistiert
kontextuelles Wissen über Sessions hinweg.

**Schreiben nach:** Projektbeschreibung, erkannter Tech-Stack, Architektur-Entscheidungen,
aktueller Fokus, offene Fragen.

**Lesen am:** Sessionbeginn, nach Context-Kompaktierung, wenn Kontext zur Projekt-Geschichte
fehlt.

**Dateien:**
- `projectbrief.md` — Projektziel und Scope
- `techContext.md` — Stack, Tools, Constraints
- `systemPatterns.md` — ADRs und wiederkehrende Muster
- `activeContext.md` — Aktueller Fokus und Blocker
- `progress.md` — Fortschritt und nächste Schritte

**Abgrenzung zu `memory`:** `memory` ist graph-basiert und projektübergreifend (`.claude/memory.json`). `memory-bank` ist dateibasiert und repo-spezifisch — ideal für strukturierten Langzeit-Kontext.
```

- [ ] **Step 3: Commit**

```bash
git add base/CLAUDE.md
git commit -m "docs(base): document memory-bank MCP server in CLAUDE.md"
```

---

### Task 6: `forgecrate-repo-onboarding` Skill — memory-bank befüllen

**Files:**
- Modify: `base/skills/forgecrate-repo-onboarding/skill.md`

- [ ] **Step 1: Neuen Schritt nach dem CLAUDE.md-Schritt einfügen**

Am Ende der Datei nach Schritt 6 (`**Übergabe**`) einfügen:

```markdown
7. **memory-bank befüllen** — schreibe direkt (keine Rückfrage) auf Basis der Analyse:

   - **`memory-bank/projectbrief.md`** — ersetze die Platzhalter:
     - *Was ist dieses Projekt?* → Kurze Beschreibung aus README/go.mod/package.json
     - *Ziele* → Erkannter Mehrwert (CLI-Tool, API, Library, ...)
     - *Nicht-Ziele* → Was bewusst fehlt (z.B. kein Frontend, kein Auth)

   - **`memory-bank/techContext.md`** — ersetze die Platzhalter:
     - *Stack* → Sprache + Version, Framework, wichtige Libraries (aus go.mod/package.json/pyproject.toml)
     - *Tools & Infrastruktur* → Build-, Test-, Lint-Kommandos (aus Makefile/scripts)
     - *Constraints* → Erkannte Einschränkungen (z.B. Go-Version, Node-Version, kein CGO)

   - **`memory-bank/systemPatterns.md`** — ersetze die Platzhalter:
     - *Architektur-Entscheidungen* → Erkannte Struktur (z.B. `internal/` für Business-Logik, `cmd/` für Einstiegspunkte)
     - *Wiederkehrende Muster* → Coding-Conventions aus dem Code (z.B. error wrapping, interface-Design)
     - *Anti-Patterns* → Lass diesen Abschnitt leer wenn nichts klar erkennbar

   - **`memory-bank/activeContext.md`** und **`memory-bank/progress.md`** — nicht anfassen, Template bleibt leer.

   Schreibe die Dateien mit den Read/Write-Tools direkt. Kein Prompt an den Nutzer.
```

- [ ] **Step 2: Prüfen, dass Struktur kohärent ist**

```bash
cat base/skills/forgecrate-repo-onboarding/skill.md | grep -c "^[0-9]\."
```

Erwartet: `7` (sieben nummerierte Schritte)

- [ ] **Step 3: Commit**

```bash
git add base/skills/forgecrate-repo-onboarding/skill.md
git commit -m "feat(skill): fill memory-bank automatically in repo-onboarding"
```

---

### Task 7: Gesamte Testsuite + Build

- [ ] **Step 1: Alle Tests**

```bash
go test ./... 2>&1 | tail -20
```

Erwartet: `ok` für alle Packages, kein `FAIL`.

- [ ] **Step 2: Build**

```bash
go build ./cmd/forgecrate/...
```

Erwartet: kein Fehler, Binary erzeugt.

- [ ] **Step 3: Smoke-Test `forgecrate init` in temp-Verzeichnis**

```bash
tmp=$(mktemp -d) && cd "$tmp" && /Users/markus/repo/forgecrate/forgecrate init --profile backend 2>&1 | head -20
```

Erwartet: `Done.` und `memory-bank/` im Verzeichnis mit 5 Dateien.

```bash
ls "$tmp/memory-bank/"
```

Erwartet: `activeContext.md  progress.md  projectbrief.md  systemPatterns.md  techContext.md`

- [ ] **Step 4: Prüfen dass bestehende memory-bank nicht überschrieben wird**

```bash
echo "custom" > "$tmp/memory-bank/projectbrief.md"
/Users/markus/repo/forgecrate/forgecrate init --profile backend 2>&1 | head -5
cat "$tmp/memory-bank/projectbrief.md"
```

Erwartet: Ausgabe enthält `custom`, nicht den Template-Inhalt.

- [ ] **Step 5: Feature-Branch mergen / PR erstellen**

```bash
git log --oneline -6
```

Alle Commits prüfen, dann PR erstellen:

```bash
gh pr create --title "feat: memory-bank als verpflichtender base layer" \
  --body "$(cat <<'EOF'
## Summary
- `base/extensions.yaml`: memory-bank MCP-Server hinzugefügt
- `base/memory-bank/`: 5 Template-Dateien angelegt
- `deploy.go`: `scaffoldMemoryBank()` — idempotentes Scaffold bei jedem Deploy
- `base/CLAUDE.md`: Dokumentation für memory-bank ergänzt
- `forgecrate-repo-onboarding`: befüllt memory-bank automatisch nach Analyse

## Test plan
- [ ] `go test ./...` grün
- [ ] `go build ./cmd/forgecrate/...` erfolgreich
- [ ] `forgecrate init` legt `memory-bank/` mit 5 Dateien an
- [ ] Zweiter `forgecrate init` überschreibt keine bestehenden Dateien

Closes #<issue>
EOF
)"
```
