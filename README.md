<div align="center">
  <img src="assets/banner.svg" alt="claude-setup — Reproduzierbares Claude-Setup" width="100%">
</div>

[![Latest Release](https://img.shields.io/github/v/release/jmt-labs/claude-setup)](https://github.com/jmt-labs/claude-setup/releases/latest)

# claude-setup

**claude-setup** deploys a reproducible [Claude Code](https://claude.ai/code) configuration to any repository. A single Go binary fetches profiles, skills, hooks, and MCP server definitions from GitHub and composes them into the target repo via a layered configuration system.

Stack: Go Binary · GitHub API · Layer-System · Hooks · Skills

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew tap jmt-labs/tap
brew install claude-setup
```

### Chocolatey (Windows)

```sh
choco install claude-setup
```

### apt (Ubuntu / Debian)

```sh
curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg
echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/jmt-labs.list
sudo apt update && sudo apt install claude-setup
```

### go install

```sh
go install github.com/jmt-labs/claude-setup/cmd/claude-setup@latest
```

### curl (manual install, no package manager)

```sh
curl -fsSL https://raw.githubusercontent.com/jmt-labs/claude-setup/main/install.sh | bash
```

Specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/jmt-labs/claude-setup/main/install.sh | bash -s v1.0.0
```

---

## Quick Start

Initialize a repository:

```sh
claude-setup init --profile backend --flavors tdd
```

This writes:

```
CLAUDE.md · AGENTS.md · .claude/settings.json · .claude/commands/ · .claude/hooks/
```

Update to the latest version:

```sh
claude-setup update
```

Switch profile:

```sh
claude-setup update --profile fullstack
```

---

## Dokumentation

| Thema | Dokument |
|---|---|
| Architektur und Komponenten | [docs/architecture.md](docs/architecture.md) |
| Abläufe (init, update, enforcement) | [docs/flows.md](docs/flows.md) |
| Layer-System | [docs/layer-system.md](docs/layer-system.md) |
| Hooks | [docs/hooks.md](docs/hooks.md) |
| Profile und Flavors | [docs/profiles-flavors.md](docs/profiles-flavors.md) |
| Entwicklung und Tests | [docs/development.md](docs/development.md) |

---

## Komponenten

| Pfad | Zweck |
|---|---|
| `base/` | Basis-Layer — immer deployt |
| `profiles/` | Profil-Layer — eines wählbar |
| `flavors/` | Flavor-Layer — mehrere kombinierbar |
| `cmd/claude-setup/` | Go-Binary (init, update) |
| `internal/compose/` | Markdown-, JSON- und Skills-Merge-Logik |
| `internal/github/` | GitHub API Client |
| `internal/config/` | `.claude-setup.yaml` Lesen/Schreiben |
| `internal/deploy/` | Deployment-Koordination |
| `e2e/` | End-to-End-Tests |

---

## Anpassung

Lokale Overrides werden nie überschrieben:

```
.claude/commands/overrides/   # eigene Skills
.claude/overrides/settings.override.json  # settings.json Erweiterungen
```

In `CLAUDE.md` und `AGENTS.md`:

```markdown
<!-- CUSTOM:BEGIN -->
Eigene Anweisungen hier
<!-- CUSTOM:END -->
```

> Vollständige Doku: [docs/layer-system.md](docs/layer-system.md)
