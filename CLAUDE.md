<!-- GENERATED:BEGIN -->
# Claude-Konfiguration

Dieses Repository nutzt eine reproduzierbare forgecrate-Konfiguration. Die hier
beschriebenen Regeln gelten für alle Agenten (Claude Code, Codex, …) die im Repo
arbeiten. Die generierten Abschnitte werden bei `forgecrate update` überschrieben —
eigene Anpassungen gehören in den CUSTOM-Abschnitt der Root-`CLAUDE.md`.

## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |

## Recherche-Pflicht beim Planen

Planungs-Rollen (Analyst, Tech Lead, Debugger, Reviewer) MÜSSEN vor jedem Plan
mindestens ein Recherche-Tool nutzen. Raten ist verboten — Quellen werden im Plan
referenziert.

| Frage-Typ | Tool | Beispiele |
|---|---|---|
| Library-/Framework-Doku | `context7` | API-Syntax, Migrationen, Versions-Updates |
| Spezifische URL aus Issue/Ticket | `fetch` MCP | RFCs, MDN, Changelogs |
| Allgemeine Web-Recherche | `WebSearch` | Best Practices, Vergleiche, aktuelle Probleme |

**Regeln:**

- Mindestens eine Quelle pro nicht-trivialer Planungsentscheidung
- Quellen im Plan-Dokument (`docs/superpowers/plans/*.md`) referenzieren
- Bei rein mechanischen Tasks (Rename, Typo, einzeiliger Fix) entfällt die Pflicht
- Deaktivierbar via Flavor `no-research`

## Entwicklungs-Workflow

Für alle Features, Bugfixes und Änderungen:

1. **Brainstorming** — `superpowers:brainstorming` aufrufen, Design abstimmen
2. **Spec** — Branch anlegen (`git checkout -b feat/<thema>`); Spec in
   `docs/superpowers/specs/YYYY-MM-DD-<thema>-design.md` schreiben und committen;
   GitHub-Issue anlegen oder verlinken; Branch-Name im Issue vermerken; Kommentar
   im Issue: "Spec fertig"
3. **Plan** — in `docs/superpowers/plans/YYYY-MM-DD-<thema>.md` schreiben und
   committen; Plan-Pfad im Issue ergänzen; Kommentar: "Plan fertig"
4. **Implementierung** — nach jedem Task kurzer Kommentar im Issue
5. **PR & Abschluss** — Vor dem PR: memory-bank aktualisieren
   (`activeContext.md`, `progress.md`) und Inhalt in die PR-Beschreibung
   einbeziehen. Existiert noch kein memory-bank-Inhalt, zuerst
   `/forgecrate-repo-onboarding` ausführen. Dann PR erstellen, Issue im
   PR-Body verlinken ("Closes #N"); Issue wird erst nach Merge des PR
   geschlossen (GitHub macht das automatisch)

Ticket-Kommentare immer kurz (ein Satz): Fortschritt, Pfad oder Ergebnis.

## Session-Start

Beim Session-Start: aktuellen Projektkontext aus der memory-bank lesen.
**Pflicht:** `mcp__memory-bank__memory_bank_read` verwenden — direktes Lesen via
Read-Tool auf `memory-bank/`-Dateien ist verboten.

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer über Branch + PR, nie direkt auf `main`

## Hook-Schutz: Hinweis

Der `pre-tool.sh`-Hook blockt destruktive Bash-Befehle auf `main` (z. B.
`git commit`, `git push`, `git reset --hard`, Schreib-Redirectionen). Er ist
jedoch **keine alleinige Schutzschicht** — GitHub Branch Protection Rules müssen
zusätzlich konfiguriert werden, damit direkte Pushes auch serverseitig verhindert
werden.

## Team-Rollen & Subagent-Konfiguration

Der Hauptagent koordiniert als Team-Lead. Subagenten übernehmen Rollen
entsprechend ihrer Aufgabe. Der Hauptagent kann bei Bedarf eigenständig von
diesen Empfehlungen abweichen.

Das Hauptmodell der Session ist global (in `.claude/settings.json`). Die
`Modell`-Spalte nennt den empfohlenen Wert für den `model`-Parameter beim
Dispatch eines Subagenten über das Agent-Tool — gültig sind nur die Family-Aliase
`opus`/`sonnet`/`haiku`.

| Rolle | Superpowers-Skill | Modell | Recherche |
|---|---|---|---|
| Analyst / Product Owner | `superpowers:brainstorming` | `opus` | Pflicht |
| Tech Lead / Architekt | `superpowers:writing-plans` | `opus` | Pflicht |
| Entwickler | `superpowers:test-driven-development` | `sonnet` | optional |
| Implementierer (mechanisch) | `superpowers:subagent-driven-development` | `haiku` | nein |
| Reviewer | `superpowers:requesting-code-review` | `sonnet` | Pflicht bei Architektur-Fragen |
| QA / Abschluss | `superpowers:verification-before-completion` | `sonnet` | nein |
| Debugger | `superpowers:systematic-debugging` | `sonnet` | Pflicht (CVE, Lib-Issues, Stack-Overflow) |

## Parallelisierung & Isolation

Subagenten werden proaktiv parallelisiert und isoliert — ohne explizite
Aufforderung.

| Situation | Mechanismus | Anleitung |
|---|---|---|
| Task dauert >1 min oder Ergebnis nicht sofort nötig | `run_in_background: true` | `superpowers:dispatching-parallel-agents` |
| Feature-Branch, Multi-File-Änderung, langer Plan | `isolation: "worktree"` | `superpowers:using-git-worktrees` |
| Mehrere unabhängige Tasks gleichzeitig | beide kombinieren | beide Skills |

Im Zweifelsfall Background nutzen — warten ist kein Default.

### Agenten-Identität

Jeder Subagent bekommt eindeutige Identifikation:

- **Eindeutigen Namen** — via `description`-Parameter im Agent-Tool-Aufruf
  (3–5 Wörter, Rolle + Aufgabe)
- **Eindeutige Farbe** — dynamisch durch FleetView-Dashboard zugewiesen; keine
  zwei gleichzeitig laufenden Agenten teilen eine Farbe

Dies ermöglicht einfaches Tracking und verhindert Verwechslungen bei parallelen
Läufen.

## MCP-Server

Sechs MCP-Server stehen automatisch zur Verfügung. `.mcp.json` wird von forgecrate
generiert — nicht von Hand editieren; MCP-Server-Änderungen über einen erneuten
forgecrate-Lauf.

| Server | Transport | Zweck |
|---|---|---|
| `github` | HTTP (GitHub Copilot) | Issues, PRs, Code-Suche, Branches, Labels |
| `fetch` | stdio (`npx`) | Externe Webinhalte: Docs, RFCs, Changelogs |
| `memory` | stdio (`npx`) | Projektübergreifende Architektur-Entscheidungen |
| `memory-bank` | stdio (`npx`) | Repo-spezifischer Projektkontext (laufender Stand) |
| `context-mode` | stdio (`npx`) | Automatisches Context-Budget und Session-History-Suche |
| `context7` | stdio (`npx`) | Aktuelle Bibliotheks-Dokumentation aus Source-Repos |

Routing-Grenzen (verhindern Falsch-Aufrufe):

- **`github`** — alle GitHub-Operationen (Issues, PRs, Code-Suche, Labels). NICHT für
  lokale Datei-/Git-Kommandos (→ Read/Edit/Bash). Voraussetzung:
  `GITHUB_PERSONAL_ACCESS_TOKEN`.
- **`fetch`** — externe Webinhalte (Docs, MDN, RFCs, Changelogs). NICHT für
  GitHub-Inhalte (→ `github`) oder lokale Dateien (→ Read).
- **`context-mode`** — sandboxt Tool-Output automatisch (kein Aufruf nötig). Explizit:
  `ctx_search` (History-Suche nach Kompaktierung), `ctx_stats`, `ctx_doctor`.
- **`context7`** — aktuelle Bibliotheks-Doku aus Source-Repos. NICHT für GitHub-Inhalte
  (→ `github`), lokale Dateien (→ Read) oder allgemeine Programmierkonzepte.

`memory` und `memory-bank` haben eigene Pflicht-Regeln — siehe unten.

### Memory (`memory`)

Projektübergreifendes Wissen persistent speichern. Datei: `.claude/memory.json`
(versioniert).

**Schreiben nach:** Architekturentscheidungen, Begründungen für nicht-
offensichtliche Lösungen, Debugging-Ergebnisse, Brainstorming-Ergebnisse.

**Lesen am:** Sessionbeginn, nach Context-Kompaktierung, wenn unklar warum etwas
so gebaut wurde.

**Niemals speichern:** API-Keys, Tokens, Passwörter, temporären Zwischenstand,
Code-Details die direkt aus dem Code lesbar sind.

### Memory-Bank (`memory-bank`)

Repo-spezifischer, strukturierter Projektkontext im Verzeichnis `memory-bank/`
(versioniert, committed). Persistiert kontextuelles Wissen über Sessions hinweg.

**Dateien:**

- `projectbrief.md` — Projektziel und Scope
- `techContext.md` — Stack, Tools, technische Constraints
- `systemPatterns.md` — Architektur-Entscheidungen, ADRs, Anti-Patterns
- `activeContext.md` — Aktueller Fokus, offene Fragen, Blocker
- `progress.md` — Was fertig ist, was läuft, was als nächstes kommt

**Lesen** am Session-Start und bei Bedarf — **ausschließlich** via
`mcp__memory-bank__memory_bank_read`.

**Schreiben** wenn sich Fokus, Fortschritt oder Architektur-Kontext ändert —
**ausschließlich** via `mcp__memory-bank__memory_bank_write` oder
`mcp__memory-bank__memory_bank_update`.

> **Direkte Datei-Tools (Read/Write/Edit) auf `memory-bank/`-Dateien sind
> verboten.**

**Abgrenzung zu `memory`:** `memory-bank` ist repo-spezifisch und dateibasiert —
ideal für laufenden Projekt-Kontext. `memory` (`.claude/memory.json`) ist
graph-basiert und projektübergreifend — ideal für zeitlose
Architektur-Entscheidungen mit Begründung.

## Backend-Profil

- API-Design: REST-First, klare Fehlercodes, keine unnötige Abstraktion
- Datenbankzugriffe: typsicher, keine Raw-Queries ohne Parametrisierung
- Tests: Integrationstests bevorzugt gegenüber reinen Unit-Tests mit Mocks
- Kein ORM-Magic: explizite Queries sind verständlicher

## Strict-Review-Flavor

- Vor jedem Commit: `superpowers:requesting-code-review` aufrufen
- Keine direkten Commits auf main/master
- PR-Beschreibung enthält: Was, Warum, Wie getestet
- Breaking Changes werden explizit kommuniziert

## TDD-Flavor

- Test schreiben → ausführen (muss fehlschlagen) → implementieren → ausführen (muss bestehen) → committen
- Kein Produktionscode ohne vorherigen Test
- Test-Namen beschreiben Verhalten, nicht Implementierung
- Mocks nur an Systemgrenzen (externe APIs, Datenbanken)
- Für jeden gefundenen Bug: Regressionstest vor dem Fix

## Codegraph-Flavor

Dieses Repo nutzt **codegraph** — einen semantischen Code-Wissensgraphen als MCP-Server.

### Was codegraph bietet

Der MCP-Server läuft lokal (`codegraph serve --mcp`) und stellt folgende Tools bereit:

| Tool | Zweck |
|---|---|
| `codegraph_search` | Semantische Code-Suche ohne exakte Schlüsselwörter |
| `codegraph_node` | Definition eines Symbols (Funktion, Typ, Variable) abrufen |
| `codegraph_callers` / `codegraph_callees` | Alle Aufrufer / Aufgerufenen eines Symbols |
| `codegraph_trace` | Aufrufpfad zwischen zwei Symbolen nachverfolgen |
| `codegraph_explore` | Abhängigkeiten und Nachbarn eines Symbols erkunden |
| `codegraph_context` | Code-Abschnitt mit Graph-Kontext erklären |
| `codegraph_impact` | Blast-Radius einer Änderung ermitteln |
| `codegraph_files` | Dateien im Index auflisten |
| `codegraph_status` | Index-Status prüfen |

### Wann nutzen

- **Vor jeder nicht-trivialen Änderung**: `codegraph_node` + `codegraph_callers` für betroffene Symbole
- **Beim Debuggen**: `codegraph_trace` um Aufrufkette nachzuvollziehen
- **Bei Refactoring**: `codegraph_callers` für Call-Sites + `codegraph_search` für Type-/Import-Referenzen
- **Code-Suche**: `codegraph_search` statt grep bei konzeptuellen Fragen
- **Impact-Analyse**: `codegraph_impact` vor größeren Umbauten

### Index-Aktualisierung

Der Index wird automatisch bei Session-Start im Hintergrund aktualisiert (einmal pro Commit-Stand).
Manuell: `codegraph index` im Repo-Root. Erstmalige Initialisierung: `codegraph init -i`.

### Voraussetzung

Installation (einmalig, kein Node.js erforderlich):

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex

# Alternativ via npm
npm i -g @colbymchenry/codegraph
```

Danach im Repo initialisieren:

```bash
codegraph init -i
```

Der MCP-Server wird über `.mcp.json` automatisch konfiguriert.
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
## Entwicklungskommandos

```bash
make build              # Binary bauen
make test               # Unit- und Integration-Tests
make test-e2e           # End-to-End-Tests
make quality            # go vet + go build sanity check
make check-model-ids    # Model-IDs nur in base/models.yaml
make check-readme-coverage  # jeder Flavor im README erwähnt
```

Manuell ohne `make`:

```bash
go build ./cmd/forgecrate          # Binary bauen
go test ./...                       # Alle Tests
golangci-lint run ./...             # Linting
go run ./cmd/forgecrate --help      # CLI lokal ausführen
```

## Architektur

| Pfad | Zweck |
|---|---|
| `cmd/forgecrate/` | CLI-Entry-Points (Cobra-Commands) |
| `internal/config/` | `.forgecrate.yaml` lesen/schreiben |
| `internal/compose/` | Layer-System: Markdown, JSON, Skills zusammenführen |
| `internal/deploy/` | Profil + Flavors nach Ziel-Repo deployen, Konflikt-Resolution |
| `internal/extensions/` | Plugin- und MCP-Server-Installation via `claude` CLI |
| `internal/github/` | Tarball-Download von GitHub-Releases |
| `base/` | Base-Layer: CLAUDE.md, Hooks, Skills, `extensions.yaml`, `models.yaml` |
| `profiles/` | Profil-Definitionen (`backend`, `frontend`, `fullstack`) |
| `flavors/` | Flavor-Definitionen (`tdd`, `strict-review`, `minimal`, `gitops`, `getbetter`, `github`, `no-research`) |
| `e2e/` | End-to-End-Tests (brauchen `claude plugin install superpowers` oder `CLAUDE_BIN`) |
<!-- CUSTOM:END -->
