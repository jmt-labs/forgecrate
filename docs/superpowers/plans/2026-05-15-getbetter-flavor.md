# GETBETTER-Flavor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Einen opt-in Flavor `getbetter` anlegen, der einen `/forgecrate-getbetter`-Slash-Command und einen Skill bereitstellt, mit dem Claude Session-Erkenntnisse in `.claude/GETBETTER.md` reflektiert und persistent speichert.

**Architecture:** Rein dateibasiert — kein Go-Code. Der Flavor besteht aus drei Markdown-Dateien: `CLAUDE.md` (Lese-Instruktion), `SKILL.md` (Synthese-Ablauf), Command-Wrapper. Der bestehende `composeSkills()`-Mechanismus in `compose.go` deployt alles automatisch.

**Tech Stack:** Markdown, Go-Tests (`testing`, `os`, `path/filepath`), bestehende `deploy.Run()` E2E-Infrastruktur.

---

### Task 1: Fehlschlagende E2E-Tests schreiben

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Tests am Ende von `e2e/e2e_test.go` einfügen**

```go
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
```

- [ ] **Step 2: Tests ausführen — müssen fehlschlagen**

```bash
go test ./e2e/ -run "TestDeployIncludesGetbetter" -v
```

Erwartet: FAIL mit `getbetter flavor skill missing: getbetter`

---

### Task 2: Flavor-Verzeichnis und CLAUDE.md anlegen

**Files:**
- Create: `flavors/getbetter/CLAUDE.md`

- [ ] **Step 1: Verzeichnis anlegen**

```bash
mkdir -p flavors/getbetter/skills/getbetter
mkdir -p flavors/getbetter/.claude/commands
```

- [ ] **Step 2: `flavors/getbetter/CLAUDE.md` erstellen**

```markdown
## GETBETTER-Flavor

Falls `.claude/GETBETTER.md` existiert, MUSS sie vor allem anderen gelesen werden.
```

- [ ] **Step 3: Commit**

```bash
git add flavors/getbetter/CLAUDE.md
git commit -m "feat: getbetter flavor — CLAUDE.md-Instruktion"
```

---

### Task 3: Skill anlegen

**Files:**
- Create: `flavors/getbetter/skills/getbetter/SKILL.md`

- [ ] **Step 1: `flavors/getbetter/skills/getbetter/SKILL.md` erstellen**

```markdown
# GETBETTER

Reflektiert die aktuelle Session und aktualisiert `.claude/GETBETTER.md` mit synthetisierten Erkenntnissen.

## Ablauf

1. **Bestehende Erkenntnisse laden**

   Prüfe ob `.claude/GETBETTER.md` existiert:
   ```bash
   cat .claude/GETBETTER.md 2>/dev/null || echo "(keine bisherigen Erkenntnisse)"
   ```

2. **Aktuelle Session reflektieren**

   Analysiere die aktuelle Session und extrahiere Erkenntnisse in diesen Kategorien:

   - **Entscheidungen** — Was wurde entschieden und warum? Welche Alternativen wurden verworfen?
   - **Anti-Patterns** — Was lief schief? Was hätte früher erkannt werden sollen?
   - **Was funktioniert** — Welche Ansätze haben sich bewährt? Was sollte beibehalten werden?

   Formuliere frei — kein starres Format, aber bleib konkret und präzise.

3. **Synthetisieren und schreiben**

   Führe bestehende und neue Erkenntnisse zusammen:
   - Bestehende Punkte bleiben erhalten, sofern sie nicht durch neue überholt werden
   - Überschneidende Punkte werden verdichtet, nicht doppelt geführt
   - Neue Erkenntnisse werden eingearbeitet

   Schreibe das Ergebnis nach `.claude/GETBETTER.md`:

   ```markdown
   # GETBETTER

   _Letzte Aktualisierung: YYYY-MM-DD_

   ## Entscheidungen
   [synthetisierter Inhalt]

   ## Anti-Patterns
   [synthetisierter Inhalt]

   ## Was funktioniert
   [synthetisierter Inhalt]
   ```

4. **Bestätigen**

   Gib eine kurze Zusammenfassung: wie viele Punkte wurden hinzugefügt, geändert, verdichtet.
```

- [ ] **Step 2: Commit**

```bash
git add flavors/getbetter/skills/getbetter/SKILL.md
git commit -m "feat: getbetter flavor — Skill"
```

---

### Task 4: Slash-Command anlegen

**Files:**
- Create: `flavors/getbetter/.claude/commands/forgecrate-getbetter.md`

- [ ] **Step 1: Command-Datei erstellen**

```markdown
---
description: Aktuelle Session reflektieren und GETBETTER.md aktualisieren
---

Use the Skill tool to invoke the "getbetter" skill.
```

- [ ] **Step 2: Tests ausführen — müssen bestehen**

```bash
go test ./e2e/ -run "TestDeployIncludesGetbetter" -v
```

Erwartet: beide Tests PASS

- [ ] **Step 3: Alle Tests ausführen**

```bash
go test ./...
```

Erwartet: alle PASS, kein FAIL

- [ ] **Step 4: Commit**

```bash
git add flavors/getbetter/.claude/commands/forgecrate-getbetter.md
git commit -m "feat: getbetter flavor — Slash-Command"
```

---

### Task 5: Abschluss

- [ ] **Step 1: Alle Tests ein letztes Mal ausführen**

```bash
go test ./...
```

Erwartet: alle PASS

- [ ] **Step 2: PR erstellen**

```bash
gh pr create \
  --title "feat: getbetter flavor" \
  --body "$(cat <<'EOF'
## Was

Fügt einen opt-in Flavor \`getbetter\` hinzu mit:
- \`/forgecrate-getbetter\` Slash-Command
- Skill, der Session-Erkenntnisse in \`.claude/GETBETTER.md\` synthetisiert
- CLAUDE.md-Instruktion: GETBETTER.md am Sessionbeginn zwingend lesen

## Warum

Claude beginnt jede Session ohne Kontext aus vorherigen Sessions. Mit GETBETTER werden Entscheidungen, Anti-Patterns und bewährte Ansätze persistent gespeichert und automatisch am nächsten Sessionbeginn geladen.

## Wie getestet

- \`TestDeployIncludesGetbetterFlavorSkill\`: Skill nach deploy vorhanden
- \`TestDeployIncludesGetbetterCommand\`: Command nach deploy vorhanden
- Alle bestehenden E2E- und Unit-Tests unverändert grün

Closes #13
EOF
)"
```
