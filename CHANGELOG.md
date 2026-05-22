# Changelog

## [Unreleased]

### Added

- `forgecrate config`: interaktive Profil-/Flavor-Konfiguration via TUI
- `forgecrate set-permission-mode`: setzt den Agent-Berechtigungsmodus
  (`bypass`/`plan`/`ask`/`auto`) und patcht `.claude/settings.json`
- `forgecrate hook prompt-submit`: Helper für den `UserPromptSubmit`-Hook
  (ersetzt das frühere Shell-Skript-basierte Auslesen der Konfiguration)
- Flavor `getbetter`: kontinuierliche Verbesserung über Sessions via
  `.claude/GETBETTER.md`
- Flavor `github`: GitHub-native Releases, CI-Checks vor Tag, proaktive
  Parallelisierung — neue Skills `/forgecrate-issue-resolver` und
  `/forgecrate-github-release`
- Flavor `no-research`: deaktiviert die Recherche-Pflicht aus dem base layer
- Profile-Skill `ui-ux-audit` (frontend) für tiefgehende UI/UX-Reviews mit
  automatischer Erstellung kleinteiliger GitHub-Issues
- Profile-Skill `accessibility-audit` (frontend) für statische A11y-Checks
- Skill `forgecrate-handoff`: portabler Projekt-Kontext für AI-Modellwechsel
  oder Session-Übergabe
- MCP-Server `memory-bank` im base layer (`@allpepper/memory-bank-mcp`) —
  strukturierter Projektkontext über `memory-bank/*.md`

### Changed

- **Renamed from `claude-setup` to `forgecrate`** — siehe [MIGRATION.md](MIGRATION.md)
  - Binary: `claude-setup` → `forgecrate` (alter Name funktioniert mit Deprecation-Warning für eine Minor-Version weiter)
  - Konfigdatei: `.claude-setup.yaml` → `.forgecrate.yaml` (Auto-Migration beim ersten Run)
  - Alle Skill-Kommandos: `claude-setup-*` → `forgecrate-*`
  - Go-Modul: `github.com/jmt-labs/claude-setup` → `github.com/jmt-labs/forgecrate`
- Memory-Bank ersetzt mem0 als Projektkontext-Quelle. Lesen/Schreiben
  ausschließlich über `mcp__memory-bank__*` Tools — direkte Datei-Tools auf
  `memory-bank/` sind verboten
- MCP-Konfiguration: `base/extensions.yaml` ist Single Source of Truth,
  `.mcp.json` wird daraus generiert

## [v0.0.3] - 2026-05-17

### Added

- feat: gitops-Flavor und describe-Subcommand (#39)
  - Neues `gitops`-Flavor: ArgoCD-App-Topologie, clusterweite Regeln
    (Kyverno/Gatekeeper/`RULES.md`), Deployments nur über ArgoCD
  - Skill `/claude-setup-gitops-status`: Drift-Check, Policy-Validierung,
    Deployment-Status
  - Neues CLI-Subcommand `claude-setup describe <profile|flavor> <name>`
- feat(base): `/handoff`-Command für portablen Repo-Kontext (#38)

## [v0.0.2] - 2026-05-16

### Added

- Homebrew-, Chocolatey- und apt-Distribution via GoReleaser (#37)

## [v0.0.1] - 2026-05-14

- Initiales Release
