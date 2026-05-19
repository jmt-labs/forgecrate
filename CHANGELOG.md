# Changelog

## [Unreleased]

### Added
- feat(frontend): Skill `ui-ux-audit` für tiefgehende UI/UX-Reviews mit automatischer Erstellung kleinteiliger GitHub-Issues. Ergänzt den bestehenden `accessibility-audit`-Skill um einen ganzheitlichen Audit über alle UI-Bereiche (Auth, Hauptworkflow, Wizards/Dialoge, Settings) mit Severity-Bewertung.
- feat: handoff- und issue-resolver-Skills sowie MCP-Konfiguration (`.mcp.json` cleanup, `MEMORY_FILE_PATH` konfiguriert)

### Changed
- **Renamed from `claude-setup` to `forgecrate`** — see [MIGRATION.md](MIGRATION.md)
  - Binary: `claude-setup` → `forgecrate` (old name still works with deprecation warning for one minor version)
  - Config file: `.claude-setup.yaml` → `.forgecrate.yaml` (auto-migrated on first run)
  - All skill commands: `claude-setup-*` → `forgecrate-*`
  - Go module: `github.com/jmt-labs/claude-setup` → `github.com/jmt-labs/forgecrate`

## [v0.0.3] - 2026-05-17

### Added
- feat: gitops-Flavor und describe-Subcommand (#39)
  - Neues `gitops`-Flavor: ArgoCD-App-Topologie, clusterweite Regeln (Kyverno/Gatekeeper/RULES.md), Deployments nur über ArgoCD
  - Skill `/claude-setup-gitops-status`: Drift-Check, Policy-Validierung, Deployment-Status
  - Neues CLI-Subcommand `claude-setup describe <profile|flavor> <name>`
- feat(base): /handoff-Command für portablen Repo-Kontext (#38)

## [v0.0.2] - 2026-05-16

### Added
- Homebrew, Chocolatey und apt distribution via GoReleaser (#37)

## [v0.0.1] - 2026-05-14

- Initiales Release
