<!-- GENERATED:BEGIN -->
<!-- GENERATED:BEGIN -->
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
<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
<!-- CUSTOM:END -->


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

<!-- GENERATED:END -->

<!-- CUSTOM:BEGIN -->
<!-- CUSTOM:END -->
