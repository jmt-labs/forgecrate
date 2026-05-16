# Claude-Konfiguration

Dieses Repository verwendet ein reproduzierbares Claude-Setup.
Die generierten Abschnitte dieser Datei werden bei `claude-setup update` überschrieben.
Eigene Anpassungen gehören in den CUSTOM-Abschnitt.

## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |

## Entwicklungs-Workflow

Für alle Features, Bugfixes und Änderungen:

1. **Brainstorming** — `superpowers:brainstorming` aufrufen, Design abstimmen
2. **Spec** — Branch anlegen (`git checkout -b feat/<thema>`); Spec in `docs/superpowers/specs/YYYY-MM-DD-<thema>-design.md` schreiben und committen; GitHub-Issue anlegen oder verlinken; Branch-Name im Issue vermerken; Kommentar im Issue: "Spec fertig"
3. **Plan** — in `docs/superpowers/plans/YYYY-MM-DD-<thema>.md` schreiben und committen; Plan-Pfad im Issue ergänzen; Kommentar: "Plan fertig"
4. **Implementierung** — nach jedem Task kurzer Kommentar im Issue
5. **PR & Abschluss** — PR erstellen, Issue im PR-Body verlinken ("Closes #N"); Issue wird erst nach Merge des PR geschlossen (GitHub macht das automatisch)

Ticket-Kommentare immer kurz (ein Satz): Fortschritt, Pfad, oder Ergebnis.

## Verhalten

- Antworte auf Deutsch
- Keine unnötigen Kommentare im Code
- YAGNI: keine ungefragten Features
- Änderungen immer über Branch + PR, nie direkt auf `main`

## Team-Rollen & Subagent-Konfiguration

Der Hauptagent koordiniert als Team-Lead. Subagenten übernehmen Rollen entsprechend ihrer Aufgabe.
Der Hauptagent kann bei Bedarf eigenständig von diesen Empfehlungen abweichen.

| Rolle | Superpowers-Skill | Modell | Effort |
|---|---|---|---|
| Analyst / Product Owner | `superpowers:brainstorming` | `claude-opus-4-7` | high |
| Tech Lead / Architekt | `superpowers:writing-plans` | `claude-opus-4-7` | high |
| Entwickler | `superpowers:test-driven-development` | `claude-sonnet-4-6` | medium |
| Implementierer (mechanisch) | `superpowers:subagent-driven-development` | `claude-haiku-4-5-20251001` | low |
| Reviewer | `superpowers:requesting-code-review` | `claude-sonnet-4-6` | medium |
| QA / Abschluss | `superpowers:verification-before-completion` | `claude-sonnet-4-6` | medium |
| Debugger | `superpowers:systematic-debugging` | `claude-sonnet-4-6` | medium |

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

Vier MCP-Server sind im base layer deklariert und stehen automatisch zur Verfügung.

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

**Keine Konfiguration nötig** — wird beim ersten `claude-setup init/update` automatisch als Projekt-MCP-Server eingerichtet.
