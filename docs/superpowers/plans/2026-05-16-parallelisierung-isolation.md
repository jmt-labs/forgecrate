# Parallelisierung & Isolation — Implementierungsplan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Einen neuen Abschnitt "Parallelisierung & Isolation" in `base/CLAUDE.md` einfügen, der Claude mit einer Entscheidungsmatrix ausstattet, wann Subagenten im Hintergrund dispatcht oder in Worktrees isoliert werden sollen.

**Architecture:** Rein additiv — nur `base/CLAUDE.md` wird geändert. Der neue Abschnitt landet zwischen "Team-Rollen & Subagent-Konfiguration" und "MCP Server". Kein Go-Code, keine neuen Dateien.

**Tech Stack:** Markdown, Go-Tests (`testing`, `os`, `path/filepath`, `strings`), bestehende `deploy.Run()` E2E-Infrastruktur.

---

### Task 1: Fehlschlagenden E2E-Test schreiben

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Test am Ende von `e2e/e2e_test.go` einfügen**

```go
func TestDeployIncludesParallelisierungSection(t *testing.T) {
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
```

- [ ] **Step 2: Test ausführen — muss fehlschlagen**

```bash
go test ./e2e/ -run TestDeployIncludesParallelisierungSection -v
```

Erwartet: `FAIL` mit `CLAUDE.md missing section: Parallelisierung & Isolation`

- [ ] **Step 3: Commit**

```bash
git add e2e/e2e_test.go
git commit -m "test: fehlschlagender E2E-Test für Parallelisierung-Abschnitt"
```

---

### Task 2: Abschnitt in `base/CLAUDE.md` einfügen

**Files:**
- Modify: `base/CLAUDE.md` (nach Zeile 48, vor `## MCP Server`)

- [ ] **Step 1: Neuen Abschnitt einfügen**

In `base/CLAUDE.md` den folgenden Block **zwischen** der letzten Zeile der Team-Rollen-Tabelle und `## MCP Server` einfügen:

```markdown
## Parallelisierung & Isolation

Subagenten werden proaktiv parallelisiert und isoliert — ohne explizite Aufforderung.

| Situation | Mechanismus | Anleitung |
|---|---|---|
| Task dauert >1 min oder Ergebnis nicht sofort nötig | `run_in_background: true` | `superpowers:dispatching-parallel-agents` |
| Feature-Branch, Multi-File-Änderung, langer Plan | `isolation: "worktree"` | `superpowers:using-git-worktrees` |
| Mehrere unabhängige Tasks gleichzeitig | beide kombinieren | beide Skills |

Im Zweifelsfall Background nutzen — warten ist kein Default.
```

Die Datei `base/CLAUDE.md` sieht nach der Änderung so aus (relevanter Ausschnitt):

```markdown
| Debugger | `superpowers:systematic-debugging` | `claude-sonnet-4-6` | medium |

## Parallelisierung & Isolation

Subagenten werden proaktiv parallelisiert und isoliert — ohne explizite Aufforderung.

| Situation | Mechanismus | Anleitung |
|---|---|---|
| Task dauert >1 min oder Ergebnis nicht sofort nötig | `run_in_background: true` | `superpowers:dispatching-parallel-agents` |
| Feature-Branch, Multi-File-Änderung, langer Plan | `isolation: "worktree"` | `superpowers:using-git-worktrees` |
| Mehrere unabhängige Tasks gleichzeitig | beide kombinieren | beide Skills |

Im Zweifelsfall Background nutzen — warten ist kein Default.
## MCP Server
```

- [ ] **Step 2: E2E-Test ausführen — muss bestehen**

```bash
go test ./e2e/ -run TestDeployIncludesParallelisierungSection -v
```

Erwartet: `PASS`

- [ ] **Step 3: Alle Tests ausführen**

```bash
go test ./...
```

Erwartet: alle `PASS`, kein `FAIL`

- [ ] **Step 4: Commit**

```bash
git add base/CLAUDE.md
git commit -m "feat: Parallelisierung & Isolation — Entscheidungsmatrix in base/CLAUDE.md"
```

---

### Task 3: PR erstellen

- [ ] **Step 1: Alle Tests ein letztes Mal ausführen**

```bash
go test ./...
```

Erwartet: alle `PASS`

- [ ] **Step 2: PR erstellen**

```bash
gh pr create \
  --title "feat: Parallelisierung & Isolation in base/CLAUDE.md" \
  --body "$(cat <<'EOF'
## Was

Fügt einen neuen Abschnitt \"Parallelisierung & Isolation\" in \`base/CLAUDE.md\` ein mit einer Entscheidungsmatrix:
- \`run_in_background: true\` — für Tasks >1 min oder ohne sofortigen Bedarf
- \`isolation: \"worktree\"\` — für Feature-Branches und Multi-File-Änderungen
- Verweis auf \`superpowers:dispatching-parallel-agents\` und \`superpowers:using-git-worktrees\`

## Warum

Claude nutzt Background-Dispatch und Worktree-Isolation bisher nur auf explizite Aufforderung.
Die Entscheidungsmatrix verankert das proaktive Verhalten direkt im Kontext — kein Skill-Lookup nötig.

## Wie getestet

- \`TestDeployIncludesParallelisierungSection\`: prüft nach deploy ob CLAUDE.md den Abschnitt samt \`run_in_background\` und \`isolation: \"worktree\"\` enthält
- Alle bestehenden Tests unverändert grün
EOF
)"
```
