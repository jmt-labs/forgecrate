# System Patterns

## Architektur-Entscheidungen

<!-- Wichtige ADRs: Was wurde entschieden und warum? -->

### ADR: codegraph-Integration als opt-in Flavor (2026-05-22)

**Kontext:** codegraph ist ein semantischer Code-Knowledge-Graph-MCP-Server (MIT-Lizenz,
https://github.com/colbymchenry/codegraph). Er verspricht laut Anbieter ~35% günstigere
API-Calls und ~70% weniger Tool-Aufrufe durch direkten Zugriff auf die Projektstruktur
(`codegraph_search`, `codegraph_context`, `codegraph_impact`).

**Entscheidung: Flavor, nicht Base-Layer.**

Begründung: codegraph benötigt Pro-Projekt-Initialisierung (lokale SQLite-Datenbank unter
`.codegraph/`). Die Base-Layer wird für alle Projekte deployed — auch kleine Repos oder
Konfig-Repos, für die codegraph keinen Mehrwert bietet. Opt-in via Flavor passt zur
forgecrate-Philosophie.

**MCP-Server-Konfiguration (`flavors/codegraph/extensions.yaml`):**
```yaml
mcp:
  - name: codegraph
    command: codegraph
    args: ["serve", "--mcp"]
```
`npx @colbymchenry/codegraph` startet den interaktiven Installer — nicht den Server.
Der korrekte Start-Befehl ist `codegraph serve --mcp`.

**Init ist vollständig skriptbar:**
- `codegraph install --yes` — Installation ohne interaktive Prompts (ausgeführt vom Session-Start-Hook)
- `codegraph init [path]` — Initialisierung ohne Prompts
- `codegraph init [path] --index` — Init + sofort indexieren (`-i` = `--index`, nicht "interaktiv")

**Session-Start-Hook statt CLAUDE.md-Hinweis:**
Der Hook prüft ob `.codegraph/` existiert und führt automatisch `codegraph install --yes`
+ `codegraph init . --index` aus. Der Nutzer muss nichts manuell tun.

**Post-Commit-Hook für Index-Aktualität:**
`codegraph sync .` nach jedem Commit hält den Index aktuell, auch wenn der MCP-Server
gerade nicht läuft (der Watch-Modus deckt nur die Laufzeit ab).

**`.codegraph/` ins `.gitignore`:**
Der Index ist projektlokal und wird von jedem Clone via Session-Start-Hook neu aufgebaut.
Nicht ins Repository committen.

## Wiederkehrende Muster

<!-- Patterns die im Projekt konsistent verwendet werden. -->

## Anti-Patterns

<!-- Was soll vermieden werden und warum? -->
