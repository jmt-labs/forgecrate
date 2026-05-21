# memory-bank-mcp Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `@allpepper/memory-bank-mcp` als git-tauglichen, team-fähigen Memory-Layer in forgecrate integrieren und mem0 ersetzen.

**Architecture:** Der MCP-Server `@allpepper/memory-bank-mcp` wird in `base/extensions.yaml` als neuer `memory-bank`-Eintrag konfiguriert. Fünf Cline-Pattern-Markdown-Dateien unter `./memory-bank/` dienen als versionierter, team-geteilter Projektkontext. Mem0 ist ein globales Claude Code Plugin (in `~/.claude/settings.json`), nicht in forgecrate selbst — die Deaktivierung ist ein manueller Schritt für den Nutzer.

**Tech Stack:** Go (forgecrate CLI), YAML (extensions.yaml), Markdown, npx (`@allpepper/memory-bank-mcp@0.2.2`)

---

## Dateien die sich ändern

| Datei | Aktion |
|---|---|
| `base/extensions.yaml` | Modify: memory-bank MCP-Eintrag hinzufügen |
| `CLAUDE.md` | Modify: Memory-Abschnitt um memory-bank ergänzen |
| `memory-bank/projectbrief.md` | Create: Seed-Datei |
| `memory-bank/activeContext.md` | Create: Seed-Datei |
| `memory-bank/progress.md` | Create: Seed-Datei |
| `memory-bank/systemPatterns.md` | Create: Seed-Datei |
| `memory-bank/techContext.md` | Create: Seed-Datei |

---

## Task 1: memory-bank MCP-Eintrag in base/extensions.yaml

**Files:**
- Modify: `base/extensions.yaml`

- [ ] **Schritt 1: Aktuellen Stand lesen**

```bash
cat base/extensions.yaml
```

- [ ] **Schritt 2: memory-bank-Eintrag nach dem memory-Eintrag einfügen**

In `base/extensions.yaml` den Block nach dem bestehenden `memory`-Eintrag ergänzen:

```yaml
  - name: memory-bank
    command: npx
    args: ["-y", "@allpepper/memory-bank-mcp"]
    env:
      MEMORY_BANK_ROOT: "./memory-bank"
```

Die Datei soll danach so aussehen (MCP-Abschnitt):

```yaml
mcp:
  - name: github
    transport: http
    url: https://api.githubcopilot.com/mcp/

  - name: fetch
    command: npx
    args: ["-y", "@modelcontextprotocol/server-fetch"]

  - name: memory
    command: npx
    args: ["-y", "@modelcontextprotocol/server-memory"]
    env:
      MEMORY_FILE_PATH: ".claude/memory.json"

  - name: memory-bank
    command: npx
    args: ["-y", "@allpepper/memory-bank-mcp"]
    env:
      MEMORY_BANK_ROOT: "./memory-bank"

  - name: context-mode
    command: npx
    args: ["-y", "context-mode"]

  - name: context7
    command: npx
    args: ["-y", "@upstash/context7-mcp"]
```

- [ ] **Schritt 3: Datei prüfen**

```bash
cat base/extensions.yaml
```

Erwartung: `memory-bank`-Block ist vorhanden, alle anderen Einträge unverändert.

- [ ] **Schritt 4: Committen**

```bash
git add base/extensions.yaml
git commit -m "feat(extensions): memory-bank MCP-Server hinzufügen"
```

---

## Task 2: Seed-Dateien für memory-bank anlegen

**Files:**
- Create: `memory-bank/projectbrief.md`
- Create: `memory-bank/activeContext.md`
- Create: `memory-bank/progress.md`
- Create: `memory-bank/systemPatterns.md`
- Create: `memory-bank/techContext.md`

- [ ] **Schritt 1: Verzeichnis anlegen**

```bash
mkdir -p memory-bank
```

- [ ] **Schritt 2: projectbrief.md anlegen**

Inhalt für `memory-bank/projectbrief.md`:

```markdown
# Project Brief

## Was ist dieses Projekt?

<!-- Kurze Beschreibung: was tut dieses Projekt, für wen, warum. -->

## Ziele

<!-- Welche Probleme löst es? Was ist der Mehrwert? -->

## Nicht-Ziele

<!-- Was ist explizit out of scope? -->
```

- [ ] **Schritt 3: activeContext.md anlegen**

Inhalt für `memory-bank/activeContext.md`:

```markdown
# Active Context

## Aktueller Fokus

<!-- Woran wird gerade gearbeitet? Welches Feature, welcher Bug? -->

## Offene Fragen

<!-- Ungeklärte Punkte, ausstehende Entscheidungen. -->

## Bekannte Blocker

<!-- Was hält den Fortschritt auf? -->
```

- [ ] **Schritt 4: progress.md anlegen**

Inhalt für `memory-bank/progress.md`:

```markdown
# Progress

## Fertig

<!-- Abgeschlossene Features und Meilensteine. -->

## In Arbeit

<!-- Aktuell laufende Arbeiten. -->

## Nächste Schritte

<!-- Was kommt als nächstes dran? -->
```

- [ ] **Schritt 5: systemPatterns.md anlegen**

Inhalt für `memory-bank/systemPatterns.md`:

```markdown
# System Patterns

## Architektur-Entscheidungen

<!-- Wichtige ADRs: Was wurde entschieden und warum? -->

## Wiederkehrende Muster

<!-- Patterns die im Projekt konsistent verwendet werden. -->

## Anti-Patterns

<!-- Was soll vermieden werden und warum? -->
```

- [ ] **Schritt 6: techContext.md anlegen**

Inhalt für `memory-bank/techContext.md`:

```markdown
# Tech Context

## Stack

<!-- Programmiersprachen, Frameworks, wichtige Libraries. -->

## Tools & Infrastruktur

<!-- CI/CD, Linting, Test-Runner, Deployment. -->

## Constraints

<!-- Technische Einschränkungen, die Entscheidungen beeinflussen. -->
```

- [ ] **Schritt 7: Dateien prüfen**

```bash
ls -la memory-bank/
```

Erwartung: 5 Markdown-Dateien vorhanden.

- [ ] **Schritt 8: Committen**

```bash
git add memory-bank/
git commit -m "feat(memory-bank): Cline-Pattern Seed-Dateien anlegen"
```

---

## Task 3: CLAUDE.md Memory-Abschnitt aktualisieren

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Schritt 1: Bestehenden Memory-Abschnitt lesen**

```bash
grep -n "memory\|Memory\|mem0" CLAUDE.md
```

- [ ] **Schritt 2: Neuen memory-bank-Abschnitt nach dem bestehenden Memory-Abschnitt einfügen**

Den folgenden Block nach der `### Memory (`memory`)` Sektion (nach Zeile ~136) in `CLAUDE.md` einfügen:

```markdown
### Memory Bank (`memory-bank`)

Team-geteilter Projektkontext. Verzeichnis: `./memory-bank/` (versioniert, committed).

**Dateien:**
- `projectbrief.md` — Was & Warum des Projekts
- `activeContext.md` — Aktueller Fokus, offene Fragen, Blocker
- `progress.md` — Was fertig ist, was läuft, was als nächstes kommt
- `systemPatterns.md` — Architektur-Entscheidungen, ADRs, Anti-Patterns
- `techContext.md` — Stack, Tools, technische Constraints

**Schreiben:** Wenn sich Fokus, Fortschritt oder Architektur-Kontext ändert.

**Lesen:** Am Session-Start, um den aktuellen Team-Kontext zu verstehen.

**Abgrenzung zu `memory`:** `memory-bank` ist für laufenden Projekt-Kontext (was passiert gerade). `memory` (`.claude/memory.json`) ist für zeitlose Architektur-Entscheidungen mit Begründung.
```

- [ ] **Schritt 3: MCP-Konfigurationshinweis aktualisieren**

In der Zeile `(inkl. Umgebungsvariablen wie `MEMORY_FILE_PATH`)` in Zeile ~160 ergänzen:

Alt:
```
(inkl. Umgebungsvariablen wie `MEMORY_FILE_PATH`)
```

Neu:
```
(inkl. Umgebungsvariablen wie `MEMORY_FILE_PATH`, `MEMORY_BANK_ROOT`)
```

- [ ] **Schritt 4: Änderungen prüfen**

```bash
grep -A 15 "Memory Bank" CLAUDE.md
```

Erwartung: Neuer Abschnitt mit allen 5 Dateien dokumentiert.

- [ ] **Schritt 5: Committen**

```bash
git add CLAUDE.md
git commit -m "docs(claude-md): memory-bank MCP-Server dokumentieren"
```

---

## Task 4: Manuelle mem0-Deaktivierung dokumentieren

Mem0 ist kein forgecrate-Bestandteil — es ist ein globales Claude Code Plugin in `~/.claude/settings.json`. Die Deaktivierung kann nicht automatisiert werden, muss aber kommuniziert werden.

**Files:**
- Modify: `CLAUDE.md` (Hinweis ergänzen)
- Modify: `MIGRATION.md` (Migrationsanleitung)

- [ ] **Schritt 1: MIGRATION.md lesen**

```bash
head -50 MIGRATION.md
```

- [ ] **Schritt 2: Hinweis in MIGRATION.md ergänzen**

Am Ende von `MIGRATION.md` folgenden Abschnitt hinzufügen:

```markdown
## mem0 → memory-bank (ab Version X.Y)

forgecrate verwendet jetzt `@allpepper/memory-bank-mcp` statt mem0.

**Manuelle Schritte nach `forgecrate update`:**

1. mem0 Plugin in Claude Code deaktivieren:
   - Claude Code öffnen → Settings → Plugins → `mem0` deaktivieren
   - Oder in `~/.claude/settings.json`: `"mem0@mem0-plugins": false` setzen

2. `memory-bank/` Dateien mit eurem Projektkontext befüllen
   (Seed-Dateien mit leeren Abschnitten sind bereits vorhanden)

**Bestehende memory.json bleibt unverändert** — sie wird weiterhin für projektübergreifende Architektur-Entscheidungen genutzt.
```

- [ ] **Schritt 3: Prüfen**

```bash
tail -25 MIGRATION.md
```

Erwartung: Migrations-Abschnitt ist sichtbar.

- [ ] **Schritt 4: Committen**

```bash
git add MIGRATION.md
git commit -m "docs(migration): mem0 zu memory-bank Migrationsanleitung"
```

---

## Task 5: Smoke-Test — npx Package verfügbar

**Files:** keine

- [ ] **Schritt 1: Package-Verfügbarkeit prüfen**

```bash
npx -y @allpepper/memory-bank-mcp --version 2>&1 || echo "kein --version flag"
```

Erwartung: Kein Fehler, Ausgabe zeigt Version oder "kein --version flag".

- [ ] **Schritt 2: Prüfen ob MEMORY_BANK_ROOT respektiert wird**

```bash
MEMORY_BANK_ROOT="./memory-bank" npx -y @allpepper/memory-bank-mcp &
sleep 2
kill %1 2>/dev/null
echo "Server gestartet ohne Fehler"
```

Erwartung: Server startet und läuft 2 Sekunden ohne Fehler.

- [ ] **Schritt 3: git status prüfen — alles committed**

```bash
git status
```

Erwartung: `nothing to commit, working tree clean`

---

## Abschluss

Nach allen Tasks:

```bash
git log --oneline -5
```

Erwartetes Log:
```
... docs(migration): mem0 zu memory-bank Migrationsanleitung
... docs(claude-md): memory-bank MCP-Server dokumentieren
... feat(memory-bank): Cline-Pattern Seed-Dateien anlegen
... feat(extensions): memory-bank MCP-Server hinzufügen
```

Danach PR erstellen und Issue verlinken.
