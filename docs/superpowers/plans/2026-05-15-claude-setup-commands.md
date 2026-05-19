# forgecrate Slash-Commands Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Slash-Commands `/forgecrate-*` für alle forgecrate-Skills in `base/.claude/commands/` anlegen, damit sie nach `forgecrate run` sofort als `/forgecrate-<name>` verfügbar sind.

**Architecture:** Command-Wrapper-Dateien in `base/.claude/commands/` werden vom bestehenden `composeSkills()` in `compose.go` automatisch nach `~/.claude/commands/` gemergt. Jede Datei ist ein minimaler Wrapper, der den Skill-Tool-Aufruf beschreibt. Kein Umbau an Go-Code nötig.

**Tech Stack:** Markdown-Dateien, Go-Tests (`testing`, `os`, `path/filepath`), bestehende `deploy.Run()` E2E-Infrastruktur.

---

### Task 1: Fehlschlagender E2E-Test für base-Commands

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Fehlschlagenden Test schreiben**

Am Ende von `e2e/e2e_test.go` einfügen:

```go
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
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./e2e/ -run TestBaseCommandsDeployed -v
```

Erwartet: FAIL mit `missing base command: forgecrate-advisor.md`

---

### Task 2: Base-Command-Dateien anlegen

**Files:**
- Create: `base/.claude/commands/forgecrate-advisor.md`
- Create: `base/.claude/commands/forgecrate-release.md`
- Create: `base/.claude/commands/forgecrate-repo-health.md`
- Create: `base/.claude/commands/forgecrate-repo-onboarding.md`

- [ ] **Step 1: Verzeichnis anlegen**

```bash
mkdir -p base/.claude/commands
```

- [ ] **Step 2: `forgecrate-advisor.md` erstellen**

Inhalt von `base/.claude/commands/forgecrate-advisor.md`:

```markdown
---
description: Analysiere dieses Repo und empfehle das passende forgecrate-Profil und Flavors
---

Use the Skill tool to invoke the "forgecrate-advisor" skill.
```

- [ ] **Step 3: `claude-setup-release.md` erstellen**

Inhalt von `base/.claude/commands/forgecrate-release.md`:

```markdown
---
description: Führe einen vollständigen Release-Zyklus durch
---

Use the Skill tool to invoke the "release" skill.
```

- [ ] **Step 4: `claude-setup-repo-health.md` erstellen**

Inhalt von `base/.claude/commands/forgecrate-repo-health.md`:

```markdown
---
description: Analysiere das Repo auf Verbesserungspotenzial und gib eine priorisierte Liste zurück
---

Use the Skill tool to invoke the "repo-health" skill.
```

- [ ] **Step 5: `claude-setup-repo-onboarding.md` erstellen**

Inhalt von `base/.claude/commands/forgecrate-repo-onboarding.md`:

```markdown
---
description: Erkunde das Repo nach forgecrate run und erstelle einen strukturierten Überblick für CLAUDE.md
---

Use the Skill tool to invoke the "repo-onboarding" skill.
```

- [ ] **Step 6: Test ausführen — muss bestehen**

```bash
go test ./e2e/ -run TestBaseCommandsDeployed -v
```

Erwartet: PASS

- [ ] **Step 7: Commit**

```bash
git add base/.claude/commands/ e2e/e2e_test.go
git commit -m "feat: slash-commands für base-skills (forgecrate-*)"
```

---

### Task 3: Fehlschlagender Test für Profil/Flavor-Commands

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Test für profile/flavor-spezifische Commands schreiben**

Am Ende von `e2e/e2e_test.go` einfügen:

```go
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
}
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./e2e/ -run TestProfileFlavorCommandsDeployed -v
```

Erwartet: FAIL mit `missing command: claude-setup-db-migration.md`

---

### Task 4: Profil/Flavor-Command-Dateien anlegen

**Files:**
- Create: `profiles/backend/.claude/commands/forgecrate-db-migration.md`
- Create: `profiles/frontend/.claude/commands/forgecrate-accessibility-audit.md`
- Create: `flavors/tdd/.claude/commands/forgecrate-test-coverage.md`
- Create: `flavors/strict-review/.claude/commands/forgecrate-pr-checklist.md`
- Create: `flavors/github/.claude/commands/forgecrate-github-release.md`

- [ ] **Step 1: Verzeichnisse anlegen**

```bash
mkdir -p profiles/backend/.claude/commands
mkdir -p profiles/frontend/.claude/commands
mkdir -p flavors/tdd/.claude/commands
mkdir -p flavors/strict-review/.claude/commands
mkdir -p flavors/github/.claude/commands
```

- [ ] **Step 2: `claude-setup-db-migration.md` erstellen**

Inhalt von `profiles/backend/.claude/commands/forgecrate-db-migration.md`:

```markdown
---
description: Führe durch Erstellung und Review einer Datenbankmigrierung
---

Use the Skill tool to invoke the "db-migration" skill.
```

- [ ] **Step 3: `claude-setup-accessibility-audit.md` erstellen**

Inhalt von `profiles/frontend/.claude/commands/forgecrate-accessibility-audit.md`:

```markdown
---
description: Prüfe Barrierefreiheit in geänderten UI-Komponenten
---

Use the Skill tool to invoke the "accessibility-audit" skill.
```

- [ ] **Step 4: `claude-setup-test-coverage.md` erstellen**

Inhalt von `flavors/tdd/.claude/commands/forgecrate-test-coverage.md`:

```markdown
---
description: Analysiere Testabdeckung und schlage den nächsten konkreten Test vor
---

Use the Skill tool to invoke the "test-coverage" skill.
```

- [ ] **Step 5: `claude-setup-pr-checklist.md` erstellen**

Inhalt von `flavors/strict-review/.claude/commands/forgecrate-pr-checklist.md`:

```markdown
---
description: Systematische Überprüfung vor gh pr create
---

Use the Skill tool to invoke the "pr-checklist" skill.
```

- [ ] **Step 6: `claude-setup-github-release.md` erstellen**

Inhalt von `flavors/github/.claude/commands/forgecrate-github-release.md`:

```markdown
---
description: Erstelle ein GitHub Release für den soeben erstellten Tag
---

Use the Skill tool to invoke the "github-release" skill.
```

- [ ] **Step 7: Test ausführen — muss bestehen**

```bash
go test ./e2e/ -run TestProfileFlavorCommandsDeployed -v
```

Erwartet: PASS

- [ ] **Step 8: Alle Tests ausführen**

```bash
go test ./...
```

Erwartet: alle PASS

- [ ] **Step 9: Commit**

```bash
git add profiles/ flavors/ e2e/e2e_test.go
git commit -m "feat: slash-commands für profil/flavor-skills (forgecrate-*)"
```

---

### Task 5: Abschluss

- [ ] **Step 1: Alle Tests ein letztes Mal ausführen**

```bash
go test ./...
```

Erwartet: alle PASS, kein Skip

- [ ] **Step 2: PR erstellen**

```bash
gh pr create \
  --title "feat: slash-commands für forgecrate skills" \
  --body "$(cat <<'EOF'
## Was

Fügt `/forgecrate-*` Slash-Commands für alle Skills hinzu.

## Warum

Skills waren bisher nur über den Skill-Tool-Namen erreichbar. Mit den Commands sind sie direkt per Slash-Command aufrufbar, ohne den genauen Namen zu kennen.

## Wie getestet

- `TestBaseCommandsDeployed`: prüft dass base-Commands nach deploy vorhanden sind
- `TestProfileFlavorCommandsDeployed`: prüft dass profil/flavor-Commands korrekt gemergt werden
- Alle bestehenden E2E- und Unit-Tests unverändert grün

## Nächster Schritt

Plugin-Wrapper für echten `/forgecrate:<name>`-Namespace (separates Feature).

Closes #<issue>
EOF
)"
```
