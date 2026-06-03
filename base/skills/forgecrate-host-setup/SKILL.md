---
name: forgecrate-host-setup
description: Richtet eine Maschine vollständig für forgecrate + Claude ein — installiert die Union aller Plugins und MCP-Server (host-global oder projektweit) und fehlende Prerequisites.
---

# Host Setup

Bereitet einen Rechner darauf vor, Claude mit forgecrate zu nutzen: installiert
alle benötigten Plugins, MCP-Server und prüft/installiert Prerequisites.

## Ablauf

1. **Scope bestimmen** — `host` (user-global, alle Projekte) oder `project` (nur
   dieses Repo). Interaktiv per Dropdown oder via `--scope`.
2. **Prerequisites prüfen** — `claude`, `node`/`npx`, `codegraph`.
   - `claude` fehlt → Abbruch mit Install-Hinweis.
   - `npx` fehlt → Warnung (alle npx-basierten MCP-Server betroffen).
   - `codegraph` fehlt → nach Bestätigung via offiziellem `install.sh` installieren.
3. **Union sammeln** — alle Plugins/MCP-Server über base + alle Profile + alle Flavors.
4. **Plugins installieren** — `claude plugin install --scope user` (host) bzw. `--scope project`.
5. **MCP-Server registrieren** — host: `claude mcp add --scope user`; project: `.mcp.json`.
6. **Zusammenfassung** ausgeben.

## Aufruf

```bash
forgecrate host-setup --scope host          # host-global, interaktiv bei Bedarf
forgecrate host-setup --scope project -y    # CI, projektweit, ohne Rückfragen
forgecrate host-setup --dry-run             # nur anzeigen, nichts ausführen
forgecrate host-setup --skip-prereqs        # ohne Binary-Prüfung/-Installation
```

## Hinweise

- Env-abhängige MCP-Server (z. B. `github` → `GITHUB_PERSONAL_ACCESS_TOKEN`) erfordern,
  dass die Variablen gesetzt sind; projekt-relative Server (`memory`, `memory-bank`)
  sind im Projekt-Scope am sinnvollsten.
- `--yes` führt ggf. den remote codegraph `install.sh` ohne Rückfrage aus.
- Der Befehl ist idempotent: bereits installierte Plugins / registrierte MCP-Server
  werden übersprungen.
