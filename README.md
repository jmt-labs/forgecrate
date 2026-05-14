<div align="center">
  <img src="assets/banner.svg" alt="claude-setup — Reproduzierbares Claude-Setup" width="100%">
</div>

# claude-setup

claude-setup deployt ein reproduzierbares Claude-Setup in beliebige Repos. Ein globales Go-Binary holt Konfiguration, Skills und Hooks von GitHub und compositioniert sie per Layer-System ins Ziel-Repo.

Stack: Go Binary · GitHub API · Layer-System · Hooks · Skills

---

## Quick Start

**Installation (empfohlen):**

```sh
curl -fsSL https://raw.githubusercontent.com/aidun/claude-setup/main/install.sh | bash
```

Bestimmte Version:

```sh
curl -fsSL https://raw.githubusercontent.com/aidun/claude-setup/main/install.sh | bash -s v1.0.0
```

**Alternativ via Go:**

```sh
go install github.com/markus/claude-setup/cmd/claude-setup@latest
```

**Setup im Ziel-Repo:**

```sh
claude-setup init --profile backend --flavors tdd
```

Danach enthält das Repo:

```
CLAUDE.md · AGENTS.md · .claude/settings.json · .claude/commands/ · .claude/hooks/
```

Aktualisieren:

```sh
claude-setup update
```

Profil wechseln:

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
