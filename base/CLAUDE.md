# Claude-Konfiguration

Dieses Repository nutzt eine reproduzierbare forgecrate-Konfiguration. Die hier
beschriebenen Regeln gelten für alle Agenten (Claude Code, Codex, …) die im Repo
arbeiten. Die generierten Abschnitte werden bei `forgecrate update` überschrieben —
eigene Anpassungen gehören in den CUSTOM-Abschnitt der Root-`CLAUDE.md`.

## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Nach Brainstorming | `forgecrate-roadmap-triage` | MUSS aufgerufen werden — entscheidet ob jetzt oder Future Feature |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor jeder nicht-trivialen Änderung | `forgecrate-research` | MUSS aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
| Bug gefunden (nach Debug) | `superpowers:test-driven-development` | Regressionstest schreiben, BEVOR der Fix committed wird |
| Session-Start | `mcp__memory__read_graph` | Projektübergreifendes Wissen laden |
| Architekturentscheidung / Debugging-Ergebnis | `mcp__memory__add_observations` | In memory MCP schreiben |

**Codegraph-Pflicht** (wenn codegraph-Flavor aktiv): Vor jeder nicht-trivialen Änderung `codegraph_node` + `codegraph_callers` für betroffene Symbole ausführen — kein Edit/Write ohne vorherige Codegraph-Abfrage.

## Recherche-Pflicht

**Alle** Rollen MÜSSEN vor jeder nicht-trivialen Code-Änderung mindestens ein
Recherche-Tool nutzen — statt aus gelerntem Wissen zu arbeiten. Raten ist verboten;
Quellen werden referenziert. Der `pre-tool.sh`-Hook **warnt** bei fehlender Recherche,
blockiert aber nicht.

| Frage-Typ | Tool | Beispiele |
|---|---|---|
| Library-/Framework-Doku | `context7` | API-Syntax, Migrationen, Versions-Updates |
| Spezifische URL aus Issue/Ticket | `fetch` MCP | RFCs, MDN, Changelogs |
| Allgemeine Web-Recherche | `WebSearch` | Best Practices, Vergleiche, aktuelle Probleme |

**Regeln:**

- Mindestens eine Quelle pro nicht-trivialer Entscheidung; eine Recherche pro Session
  schaltet weitere Warnungen für die Session ab
- Quellen im Plan-Dokument (`docs/superpowers/plans/*.md`) referenzieren
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
5. **PR & Abschluss** — Vor `gh pr create` diese Sequenz vollständig ausführen:
   1. `forgecrate-doc-sync` — Doku mit Code abgleichen
   2. `forgecrate-handoff` — memory-bank aktualisieren (`activeContext.md`, `progress.md`)
   3. `forgecrate-db-migration` — Migrations-Review
   4. `accessibility-audit` — A11y-Prüfung
   5. `ui-ux-audit` — UX-Review
   6. `forgecrate-pr-checklist` — Abschluss-Checkliste

   Dann PR erstellen, Issue im PR-Body verlinken ("Closes #N").
   Issue wird nach Merge automatisch geschlossen.

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

Der `pre-tool.sh`-Hook **warnt** bei destruktiven Bash-Befehlen und fehlender
Recherche — er blockiert nie. Die Verantwortung liegt beim Agenten: Warnungen
bewusst wahrnehmen, einschätzen und eine informierte Entscheidung treffen.

Für serverseitigen Schutz auf `main`: GitHub Branch Protection Rules konfigurieren.

Bei fehlender Binary, fehlendem oder kaputtem Transcript verhält sich der Hook
**fail-open** (keine Warnung).

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
| Entwickler | `superpowers:test-driven-development` | `sonnet` | Pflicht |
| Implementierer (mechanisch) | `superpowers:subagent-driven-development` | `haiku` | Pflicht |
| Reviewer | `superpowers:requesting-code-review` | `sonnet` | Pflicht |
| QA / Abschluss | `superpowers:verification-before-completion` | `sonnet` | Pflicht |
| Debugger | `superpowers:systematic-debugging` | `sonnet` | Pflicht |

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
| `github` | stdio (`npx`) | Issues, PRs, Code-Suche, Branches, Labels |
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

## Claude Plugins

Vier Plugins werden automatisch via `forgecrate deploy` installiert (`claude plugin install --scope project`).

| Plugin | Zweck |
|---|---|
| `superpowers` | Skill-System: Workflows für TDD, Brainstorming, Debugging, Reviews |
| `commit-commands` | Slash-Commands für standardisierte Commits und PRs |
| `security-guidance` | Sicherheitshinweise und Best-Practices für Code-Reviews |
| `claude-md-management` | Verwaltung und Verbesserung von CLAUDE.md-Dateien |

Plugins stellen Slash-Commands und Skills bereit — sie sind nicht über MCP aufrufbar.

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
