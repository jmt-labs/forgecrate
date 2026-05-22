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
- Deaktivierbar via Flavor `no-research` (siehe `flavors/no-research/`)

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
5. **PR & Abschluss** — PR erstellen, Issue im PR-Body verlinken ("Closes #N");
   Issue wird erst nach Merge des PR geschlossen (GitHub macht das automatisch)

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

## Konfliktbehandlung beim Deploy (`forgecrate update`)

Ein Konflikt entsteht nur, wenn **beides** gleichzeitig zutrifft: die lokale
Datei wurde seit dem letzten Deploy geändert, **und** die neue Upstream-Version
unterscheidet sich von der lokalen Version. Stimmt die lokale Änderung zufällig
mit dem Upstream überein, wird kein Konflikt ausgelöst.

> **Wichtig:** Dateien ohne gespeicherten Hash (z. B. beim ersten Update nach
> Einführung des Hash-Trackings) werden ohne Rückfrage überschrieben.

Das Tool zeigt bei einem echten Konflikt:

```
KONFLIKT: .claude/settings.json
  Deine Version: <erste Zeile der lokalen Datei, max. 80 Zeichen>
  Neue Version:  <erste Zeile des Upstream>
  [o]verwrite / [k]eep (default: keep):
```

**Entscheidung:**

- `o` — Upstream-Version übernehmen, lokale Änderungen gehen verloren
- `k` oder Enter — lokale Version behalten; der Hash der lokalen Version wird als
  neue Basis gespeichert. Beim nächsten Update entsteht erneut ein Konflikt, falls
  Upstream sich weiter ändert
- `ü` oder `u` — wie `o` (Backwards-Kompatibilität)
- `b` — wie `k` (Backwards-Kompatibilität)

**Faustregel:**

- Für `settings.json` und `CLAUDE.md`: Overrides in die CUSTOM-Sektion auslagern
- Für Hooks (`.claude/hooks/**`): eigene, nicht-verwaltete Hook-Dateien verwenden

## Team-Rollen & Subagent-Konfiguration

Der Hauptagent koordiniert als Team-Lead. Subagenten übernehmen Rollen
entsprechend ihrer Aufgabe. Der Hauptagent kann bei Bedarf eigenständig von
diesen Empfehlungen abweichen.

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

Sechs MCP-Server sind im base layer deklariert und stehen automatisch zur
Verfügung. Quelle der Wahrheit ist `base/extensions.yaml` — die Datei `.mcp.json`
wird daraus generiert (siehe [MCP-Konfiguration](#mcp-konfiguration-single-source-of-truth)).

| Server | Transport | Zweck |
|---|---|---|
| `github` | HTTP (GitHub Copilot) | Issues, PRs, Code-Suche, Branches, Labels |
| `fetch` | stdio (`npx`) | Externe Webinhalte: Docs, RFCs, Changelogs |
| `memory` | stdio (`npx`) | Projektübergreifende Architektur-Entscheidungen |
| `memory-bank` | stdio (`npx`) | Repo-spezifischer Projektkontext (laufender Stand) |
| `context-mode` | stdio (`npx`) | Automatisches Context-Budget und Session-History-Suche |
| `context7` | stdio (`npx`) | Aktuelle Bibliotheks-Dokumentation aus Source-Repos |

### GitHub (`github`)

Für alle Operationen mit GitHub: Issues, PRs, Code-Suche, Branches, Checks,
Labels.

**Verwende es für:** Issues lesen/erstellen/kommentieren, PRs öffnen/reviewen/
mergen, Code repo-übergreifend suchen, Workflow-Labels setzen.

**Verwende es NICHT für:** Lokale Dateioperationen (→ Read/Edit/Bash), lokale
Git-Kommandos (→ Bash mit git).

**Voraussetzung:** `GITHUB_PERSONAL_ACCESS_TOKEN` als Umgebungsvariable.

### Fetch (`fetch`)

Externe Webinhalte abrufen: Dokumentation, MDN, RFCs, Changelogs, Release Notes,
URLs aus Issues.

**Verwende es NICHT für:** GitHub-Inhalte (→ github MCP), lokale Dateien
(→ Read).

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

### Context-Mode (`context-mode`)

Sandboxt Tool-Output automatisch — kein expliziter Aufruf nötig.

**Explizit aufrufen:**

- `ctx_search` — nach Context-Kompaktierung: relevante Infos aus der
  Session-History finden (BM25-Suche)
- `ctx_insight` — Überblick über bisherigen Session-Verlauf
- `ctx_stats` — gespartes Context-Budget prüfen
- `ctx_doctor` — bei Problemen mit dem Server

### context7

Aktuelle Bibliotheks-Dokumentation direkt aus den Source-Repositories abrufen.

**Verwende es für:** Aktuelle API-Dokumentation, Versionsmigration,
Framework-spezifisches Debugging, Changelog-Inhalte — überall wo Trainingsdaten
veraltet sein könnten.

**Verwende es NICHT für:** GitHub-Inhalte (→ github MCP), lokale Dateien
(→ Read), allgemeine Programmierkonzepte.

## MCP-Konfiguration: Single Source of Truth

Die Datei `.mcp.json` wird aus `base/extensions.yaml` generiert. Änderungen an
MCP-Servern (Kommandos, Umgebungsvariablen wie `MEMORY_FILE_PATH` oder
`MEMORY_BANK_ROOT`) immer dort vornehmen, nicht direkt in `.mcp.json`.
