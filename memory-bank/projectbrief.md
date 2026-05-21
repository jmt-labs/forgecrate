# Project Brief

## Was ist dieses Projekt?

**forgecrate** (früher `claude-setup`) ist ein Go-CLI-Tool, das eine reproduzierbare Claude Code-Konfiguration in beliebige Git-Repositories installiert. Ein einziges Binary lädt Profile, Flavors, Hooks, Skills und MCP-Server-Definitionen von GitHub und schreibt sie per Layering-System ins Ziel-Repo.

Zielgruppe: Entwickler und Teams, die Claude Code nutzen und eine konsistente, versionierte AI-Workflow-Konfiguration über mehrere Repos hinweg wollen.

## Ziele

- Reproduzierbare Claude Code-Konfiguration per `forgecrate init --profile backend --flavors tdd`
- Layering-System: base → profile → flavor → lokal (überschreibend, nicht überschreibend)
- Kein manueller Setup: Hooks, MCP-Server, Skills werden automatisch installiert
- Update-Mechanismus mit Konfliktbehandlung (`forgecrate update`)
- Teil des jmt-labs-Toolchains (→ forgedeck)

## Nicht-Ziele

- Kein Frontend oder Web-Interface
- Kein eigener MCP-Server — nur Deployment von MCP-Konfigurationen
- Keine Verwaltung von API-Keys oder Secrets
- Kein Hosting — das Binary läuft lokal oder in CI
