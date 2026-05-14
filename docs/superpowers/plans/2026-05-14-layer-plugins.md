# Layer Plugins Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Plugin-Deklarationen layer-spezifisch ergänzen — base, frontend-Profil und strict-review-Flavor erhalten ihre Plugins.

**Architecture:** Reine YAML-Konfiguration. Der Extension-Mechanismus (`internal/extensions/`) ist vollständig implementiert und muss nicht angefasst werden. Jeder Layer bekommt eine `extensions.yaml` mit seinen Plugin-Deklarationen.

**Tech Stack:** YAML, Go 1.24 (nur für Verifikation via `go test`)

**Abhängigkeit:** `docs/superpowers/plans/2026-05-14-extensions.md` muss implementiert sein (`internal/extensions/` muss existieren). ✓

---

## Dateistruktur

| Datei | Aktion | Zweck |
|---|---|---|
| `base/extensions.yaml` | Ändern | 3 weitere Plugins ergänzen |
| `profiles/frontend/extensions.yaml` | Neu | Frontend-spezifische Plugins |
| `flavors/strict-review/extensions.yaml` | Neu | Review-Qualitäts-Plugins |

---

### Task 1: `base/extensions.yaml` — 3 Plugins ergänzen

**Files:**
- Modify: `base/extensions.yaml`

Hinweis: Für reine YAML-Inhaltsänderungen gibt es keinen sinnvollen TDD-Zyklus — der Parsing-Mechanismus ist bereits getestet. Stattdessen: schreiben, YAML-Korrektheit via bestehender Test-Suite prüfen, committen.

- [ ] **Schritt 1: `base/extensions.yaml` aktualisieren**

Vollständiger neuer Inhalt:

```yaml
plugins:
  - name: superpowers
    source: claude-plugins-official/superpowers

  - name: commit-commands
    source: claude-plugins-official/commit-commands

  - name: security-guidance
    source: claude-plugins-official/security-guidance

  - name: claude-md-management
    source: claude-plugins-official/claude-md-management

mcp:
  - name: github
    transport: http
    url: https://api.githubcopilot.com/mcp/
    scope: local

  - name: fetch
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-fetch"]

  - name: memory
    scope: local
    command: npx
    args: ["-y", "@modelcontextprotocol/server-memory"]
    env:
      MEMORY_FILE_PATH: ".claude/memory.json"

  - name: context-mode
    scope: local
    command: npx
    args: ["-y", "context-mode"]
```

- [ ] **Schritt 2: Test-Suite ausführen**

```bash
go test ./...
```

Erwartet: alle Pakete `ok` (kein Regressions-Fehler durch YAML-Änderung).

- [ ] **Schritt 3: Commit**

```bash
git add base/extensions.yaml
git commit -m "feat(base): add commit-commands, security-guidance, claude-md-management plugins"
```

---

### Task 2: `profiles/frontend/extensions.yaml` anlegen

**Files:**
- Create: `profiles/frontend/extensions.yaml`

- [ ] **Schritt 1: Datei anlegen**

```yaml
plugins:
  - name: frontend-design
    source: claude-plugins-official/frontend-design

  - name: typescript-lsp
    source: claude-plugins-official/typescript-lsp

  - name: playwright
    source: claude-plugins-official/playwright
```

- [ ] **Schritt 2: Test-Suite ausführen**

```bash
go test ./...
```

Erwartet: alle Pakete `ok`.

- [ ] **Schritt 3: Commit**

```bash
git add profiles/frontend/extensions.yaml
git commit -m "feat(frontend): declare frontend-design, typescript-lsp, playwright plugins"
```

---

### Task 3: `flavors/strict-review/extensions.yaml` anlegen

**Files:**
- Create: `flavors/strict-review/extensions.yaml`

- [ ] **Schritt 1: Datei anlegen**

```yaml
plugins:
  - name: pr-review-toolkit
    source: claude-plugins-official/pr-review-toolkit

  - name: code-simplifier
    source: claude-plugins-official/code-simplifier
```

- [ ] **Schritt 2: Test-Suite ausführen**

```bash
go test ./...
```

Erwartet: alle Pakete `ok`.

- [ ] **Schritt 3: Commit**

```bash
git add flavors/strict-review/extensions.yaml
git commit -m "feat(strict-review): declare pr-review-toolkit, code-simplifier plugins"
```
