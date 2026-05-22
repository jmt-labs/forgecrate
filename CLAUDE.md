<!-- GENERATED:BEGIN -->
# Claude-Konfiguration

Dieses Repository verwendet ein reproduzierbares forgecrate.
Die generierten Abschnitte dieser Datei werden bei `forgecrate update` überschrieben.
Eigene Anpassungen gehören in den CUSTOM-Abschnitt.

## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |

## Recherche-Pflicht beim Planen

Planungs-Rollen (Analyst, Tech Lead, Debugger, Reviewer) MÜSSEN vor jedem Plan mindestens
ein Recherche-Tool nutzen. Raten ist verboten — Quellen werden im Plan referenziert.

| Frage-Typ | Tool | Beispiele |
|---|---|---|
| Library-/Framework-Doku | `context7` | API-Syntax, Migrationen, Versions-Updates |
| Spezifische URL aus Issue/Ticket | `fetch` MCP | RFCs, MDN, Changelogs |
| Allgemeine Web-Recherche | `WebSearch` | Best Practices, Vergleiche, aktuelle Probleme |

**Regeln:**
- Mindestens eine Quelle pro nicht-trivialer Planungsentscheidung
- Quellen im Plan-Dokument (`docs/superpowers/plans/*.md`) referenzieren
- Bei reinen mechanischen Tasks (Rename, Typo, einzeiliger Fix) entfällt die Pflicht
- Deaktivierbar via Flavor `no-research` (siehe `flavors/no-research/`)

## Entwicklungs-Workflow

Für alle Features, Bugfixes und Änderungen:

1. **Brainstorming** — `superpowers:brainstorming` aufrufen, Design abstimmen
2. **Spec** — Branch anlegen (`git checkout -b feat/<thema>`); Spec in `docs/superpowers/specs/YYYY-MM-DD-<thema>-design.md` schreiben und committen; GitHub-Issue anlegen oder verlinken; Branch-Name im Issue vermerken; Kommentar im Issue: "Spec fertig"
3. **Plan** — in `docs/superpowers/plans/YYYY-MM-DD-<thema>.md` schreiben und committen; Plan-Pfad im Issue ergänzen; Kommentar: "Plan fertig"
4. **Implementierung** — nach jedem Task kurzer Kommentar im Issue
5. **PR & Abschluss** — Vor dem PR: memory-bank aktualisieren (`activeContext.md`, `progress.md`) und Inhalt in die PR-Beschreibung einbeziehen. Existiert noch kein memory-bank-Inhalt, zuerst `/forgecrate-repo-onboarding` ausführen. Dann PR erstellen, Issue im PR-Body verlinken ("Closes #N"); Issue wird erst nach Merge des PR geschlossen (GitHub macht das automatisch)

Ticket-Kommentare immer kurz (ein Satz): Fortschritt, Pfad, oder Ergebnis.

## Session-Start

Beim Session-Start: memory-bank via MCP lesen um den aktuellen Projektkontext zu verstehen.

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer über Branch + PR, nie direkt auf `main`

## Hook-Schutz: Hinweis

Der `pre-tool.sh`-Hook blockt destruktive Bash-Befehle auf `main` (z. B. `git commit`, `git push`, `git reset --hard`, Schreib-Redirectionen). Er ist jedoch **keine alleinige Schutzschicht** — GitHub Branch Protection Rules müssen zusätzlich konfiguriert werden, damit direkte Pushes auch serverseitig verhindert werden.

## Konfliktbehandlung beim Deploy (`forgecrate update`)

Ein Konflikt entsteht nur, wenn **beides** gleichzeitig zutrifft: die lokale Datei wurde seit dem letzten Deploy geändert, **und** die neue Upstream-Version unterscheidet sich von der lokalen Version. Stimmt die lokale Änderung zufällig mit dem Upstream überein, wird kein Konflikt ausgelöst.

> **Wichtig:** Dateien ohne gespeicherten Hash (z. B. beim ersten Update nach Einführung des Hash-Trackings) werden ohne Rückfrage überschrieben.

Das Tool zeigt bei einem echten Konflikt:

```
KONFLIKT: .claude/settings.json
  Deine Version: <erste Zeile der lokalen Datei, max. 80 Zeichen>
  Neue Version:  <erste Zeile des Upstream>
  [o]verwrite / [k]eep (default: keep):
```

**Entscheidung:**
- `o` — Upstream-Version übernehmen, lokale Änderungen gehen verloren
- `k` oder Enter — Lokale Version behalten, Upstream-Update wird übersprungen; der Hash der lokalen Version wird als neue Basis gespeichert — beim nächsten Update entsteht erneut ein Konflikt, falls Upstream sich weiter ändert
- `ü` oder `u` — wie `o` (Backwards-Kompatibilität)
- `b` — wie `k` (Backwards-Kompatibilität)

**Faustregel:**
- Für `settings.json` und CLAUDE.md: Overrides in die CUSTOM-Sektion auslagern
- Für Hooks (`.claude/hooks/**`): eigene, nicht-verwaltete Hook-Dateien verwenden

## Team-Rollen & Subagent-Konfiguration

Der Hauptagent koordiniert als Team-Lead. Subagenten übernehmen Rollen entsprechend ihrer Aufgabe.
Der Hauptagent kann bei Bedarf eigenständig von diesen Empfehlungen abweichen.

<!-- Modell-IDs werden zentral in base/models.yaml verwaltet. -->
<!-- Beim Upgrade: nur base/models.yaml ändern, dann forgecrate update ausführen. -->

| Rolle | Superpowers-Skill | Modell | Effort | Recherche |
|---|---|---|---|---|
| Analyst / Product Owner | `superpowers:brainstorming` | `claude-opus-4-7` (models.planning) | high | Pflicht |
| Tech Lead / Architekt | `superpowers:writing-plans` | `claude-opus-4-7` (models.planning) | high | Pflicht |
| Entwickler | `superpowers:test-driven-development` | `claude-sonnet-4-6` (models.default) | medium | optional |
| Implementierer (mechanisch) | `superpowers:subagent-driven-development` | `claude-haiku-4-5-20251001` (models.mechanical) | low | nein |
| Reviewer | `superpowers:requesting-code-review` | `claude-sonnet-4-6` (models.review) | medium | Pflicht bei Architektur-Fragen |
| QA / Abschluss | `superpowers:verification-before-completion` | `claude-sonnet-4-6` (models.review) | medium | nein |
| Debugger | `superpowers:systematic-debugging` | `claude-sonnet-4-6` (models.default) | medium | Pflicht (CVE, Lib-Issues, Stack-Overflow) |

## Parallelisierung & Isolation

Subagenten werden proaktiv parallelisiert und isoliert — ohne explizite Aufforderung.

| Situation | Mechanismus | Anleitung |
|---|---|---|
| Task dauert >1 min oder Ergebnis nicht sofort nötig | `run_in_background: true` | `superpowers:dispatching-parallel-agents` |
| Feature-Branch, Multi-File-Änderung, langer Plan | `isolation: "worktree"` | `superpowers:using-git-worktrees` |
| Mehrere unabhängige Tasks gleichzeitig | beide kombinieren | beide Skills |

Im Zweifelsfall Background nutzen — warten ist kein Default.

### Agenten-Identität

Jeder Subagent bekommt eindeutige Identifikation:
- **Eindeutigen Namen** — via `description`-Parameter im Agent-Tool-Aufruf (3–5 Wörter, Rolle + Aufgabe)
- **Eindeutige Farbe** — dynamisch durch FleetView-Dashboard zugewiesen; keine zwei gleichzeitig laufenden Agenten teilen eine Farbe

Dies ermöglicht einfaches Tracking und verhindert Verwechslungen bei parallelen Läufen.

## MCP Server

Fünf MCP-Server sind im base layer deklariert und stehen automatisch zur Verfügung.

### GitHub (`github`)

Für alle Operationen mit GitHub: Issues, PRs, Code-Suche, Branches, Checks, Labels.

**Verwende es für:** Issues lesen/erstellen/kommentieren, PRs öffnen/reviewen/mergen, Code repo-übergreifend suchen, Workflow-Labels setzen.

**Verwende es NICHT für:** Lokale Dateioperationen (→ Read/Edit/Bash), lokale Git-Kommandos (→ Bash mit git).

**Voraussetzung:** `GITHUB_PERSONAL_ACCESS_TOKEN` als Umgebungsvariable.

### Fetch (`fetch`)

Externe Webinhalte abrufen: Dokumentation, MDN, RFCs, Changelogs, Release Notes, URLs aus Issues.

**Verwende es NICHT für:** GitHub-Inhalte (→ github MCP), lokale Dateien (→ Read).

### Memory (`memory`)

Projektübergreifendes Wissen persistent speichern. Datei: `.claude/memory.json` (versioniert).

**Schreiben nach:** Architekturentscheidungen, Begründungen für nicht-offensichtliche Lösungen, Debugging-Ergebnisse, Brainstorming-Ergebnisse.

**Lesen am:** Sessionbeginn, nach Context-Kompaktierung, wenn unklar warum etwas so gebaut wurde.

**Niemals speichern:** API-Keys, Tokens, Passwörter, temporärer Zwischenstand, Code-Details die direkt aus dem Code lesbar sind.

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

### Context Mode (`context-mode`)

Sandboxt Tool-Output automatisch — kein expliziter Aufruf nötig.

**Explizit aufrufen:**
- `ctx_search` — nach Context-Kompaktierung: relevante Infos aus der Session-History finden (BM25-Suche)
- `ctx_insight` — Überblick über bisherigen Session-Verlauf
- `ctx_stats` — gespartes Context-Budget prüfen
- `ctx_doctor` — bei Problemen mit dem Server

### context7

Aktuelle Bibliotheks-Dokumentation direkt aus den Source-Repositories abrufen. Automatisch konfiguriert via `base/extensions.yaml`.

**Verwende es für:** Aktuelle API-Dokumentation, Versionsmigration, Framework-spezifisches Debugging, Changelog-Inhalte — überall wo Trainingsdaten veraltet sein könnten.

**Verwende es NICHT für:** GitHub-Inhalte (→ github MCP), lokale Dateien (→ Read), allgemeine Programmierkonzepte.

**Keine Konfiguration nötig** — wird beim ersten `forgecrate init/update` automatisch als Projekt-MCP-Server eingerichtet.

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

## MCP-Konfiguration: Single Source of Truth

Die Datei `.mcp.json` wird aus `base/extensions.yaml` generiert — `base/extensions.yaml` ist die Quelle der Wahrheit für MCP-Server-Konfigurationen (inkl. Umgebungsvariablen wie `MEMORY_FILE_PATH`, `MEMORY_BANK_ROOT`). Änderungen immer dort vornehmen, nicht direkt in `.mcp.json`.

## Projektkontext

Nutze den `memory-bank` MCP-Server um den aktuellen Projektkontext zu lesen.

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
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
## Entwicklungskommandos

```bash
go build ./cmd/forgecrate          # Binary bauen
go test ./...                       # Alle Tests ausführen
go test ./cmd/forgecrate/... -v     # Nur CLI-Tests mit Output
golangci-lint run ./...             # Linting
go run ./cmd/forgecrate --help      # CLI lokal ausführen
```

## Architektur

| Pfad | Zweck |
|---|---|
| `cmd/forgecrate/` | CLI-Entry-Points (cobra commands) |
| `internal/config/` | `.forgecrate.yaml` lesen/schreiben |
| `internal/deploy/` | Profil+Flavor nach Ziel-Repo deployen |
| `internal/extensions/` | Plugin/Skill-Installation |
| `internal/github/` | GitHub-Release-Download |
| `base/` | Base-Layer: CLAUDE.md-Template, Hooks, Skills |
| `profiles/` | Profil-Definitionen (backend, frontend, fullstack) |
| `flavors/` | Flavor-Definitionen (tdd, strict-review, ...) |
| `e2e/` | End-to-End-Tests (brauchen `plugin install superpowers`) |
<!-- CUSTOM:END -->
