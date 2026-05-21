# Design: mem0-Ersatz durch memory-bank-mcp

**Datum:** 2026-05-21  
**Status:** Approved  
**Branch:** feat/memory-bank-mcp

---

## Kontext & Problem

forgecrate konfiguriert aktuell `plugin_mem0_mem0` als MCP-Server für persistentes Memory in Claude Code. mem0 ist ein Cloud-SaaS-Dienst mit folgenden Problemen für verteilte Teams:

- Jeder Nutzer braucht einen eigenen mem0-Account
- Memories sind nicht git-versioniert und nicht PR-reviewbar
- Vendor-Lock-in und Datenschutz-Bedenken
- Kein Team-Sharing über das Repository

## Ziel

mem0 ersetzen durch eine vollständig lokale, git-taugliche, team-fähige Memory-Lösung, die als MCP-Server in Claude Code funktioniert.

## Entscheidung

**`@allpepper/memory-bank-mcp`** (`alioshr/memory-bank-mcp`) wird als primäre Team-Memory verwendet.

Bewertete Alternativen:
- Basic Memory (Python/uvx) — git-tauglich, semantische Suche, aber Python-Abhängigkeit
- `@modelcontextprotocol/server-memory` (Status quo schärfen) — JSONL, merge-anfällig bei Teams
- Vector-DBs — binäre Indizes, nicht git-tauglich

## Architektur

### Komponenten

```
Claude Code
  └── MCP: memory-bank  →  ./memory-bank/*.md  (git-tracked)
  └── MCP: memory       →  .claude/memory.json  (git-tracked, Knowledge-Graph)
```

- **memory-bank**: Team-geteilter, dateibasierter Projektkontext (Cline-Pattern)
- **memory**: Bleibt für projektübergreifende Architektur-Entscheidungen (Knowledge-Graph)
- **mem0**: Wird komplett entfernt

### Verzeichnisstruktur nach Migration

```
./memory-bank/
├── projectbrief.md       # Was & Warum des Projekts
├── activeContext.md      # Aktueller Fokus, offene Fragen
├── progress.md           # Was fertig, was noch offen
├── systemPatterns.md     # Architektur-Entscheidungen, ADRs
└── techContext.md        # Stack, Tools, Constraints
```

Alle Dateien werden in git committed. Kein Index, keine Binärdateien — reines Markdown.

### MCP-Konfiguration

```yaml
# base/extensions.yaml (nach Migration)
mcp:
  - name: memory-bank
    command: npx
    args: ["-y", "@allpepper/memory-bank-mcp"]
    env:
      MEMORY_BANK_ROOT: "./memory-bank"

  - name: memory           # bleibt unverändert
    command: npx
    args: ["-y", "@modelcontextprotocol/server-memory"]
    env:
      MEMORY_FILE_PATH: ".claude/memory.json"
```

mem0-Block wird **ersatzlos gestrichen**.

### Exponierte MCP-Tools

`@allpepper/memory-bank-mcp` stellt bereit:
- `memory_bank_read` — Datei lesen
- `memory_bank_write` — Datei schreiben/anlegen
- `memory_bank_update` — Datei aktualisieren
- `list_projects` — Projekte auflisten
- `list_project_files` — Dateien im Projekt auflisten

## Änderungen im Überblick

| Datei | Änderung |
|---|---|
| `base/extensions.yaml` | mem0-Block entfernen, memory-bank-Block hinzufügen |
| `CLAUDE.md` | Memory-Abschnitt aktualisieren (memory-bank vs. memory) |
| `./memory-bank/*.md` | Initiale Seed-Dateien anlegen (5 Cline-Pattern-Dateien) |
| `profiles/*/CLAUDE.md` | Ggf. Memory-Abschnitt anpassen |

## Nutzungskonvention (CLAUDE.md)

Nach Migration gelten folgende Schreib-Regeln:

- **memory-bank** (`./memory-bank/`) — Team-geteilter Kontext: aktuelle Aufgaben, Architektur, Stack, offene Fragen. Wird von allen Team-Mitgliedern gelesen und geschrieben.
- **memory** (`.claude/memory.json`) — Projektübergreifendes Wissen: Architekturentscheidungen mit Begründung, Debugging-Ergebnisse, nicht-offensichtliche Lösungen.
- **Nie speichern:** API-Keys, Tokens, temporärer Zwischenstand, Code-Details die direkt aus dem Code lesbar sind.

## Migrationspfad

1. Bestehende `.claude/memory.json` bleibt unverändert
2. mem0 aus `base/extensions.yaml` entfernen
3. memory-bank-Eintrag in `base/extensions.yaml` hinzufügen
4. Initiale `./memory-bank/*.md`-Dateien anlegen (leer mit Headings)
5. CLAUDE.md-Abschnitt "Memory" aktualisieren
6. `forgecrate update` für bestehende Nutzer: mem0-Block wird beim nächsten Update entfernt

## Nicht im Scope

- Migration bestehender mem0-Inhalte (Cloud-Daten sind nicht automatisch migrierbar)
- Semantische Suche (kein Embedding — bewusste Entscheidung für Einfachheit)
- Globale User-Memory (scope ist pro-Projekt)
