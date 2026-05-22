# Architektur

forgecrate ist ein Single-Binary-CLI in Go. Es lädt Profile, Flavors, Hooks und
MCP-Server-Definitionen aus einem GitHub-Release, komponiert sie über ein
Layer-System und schreibt das Ergebnis in das Ziel-Repo.

## Komponenten

```
┌─────────────────────────────────────────────────────┐
│                  forgecrate Repo                    │
│   base/  ·  profiles/  ·  flavors/  ·  cmd/         │
└───────────────────────┬─────────────────────────────┘
                        │ GitHub API (tarball)
                        ▼
              ┌─────────────────┐
              │   forgecrate    │  ← globales Go-Binary
              │   (single bin)  │
              └────────┬────────┘
                       │ download → compose → deploy
                       ▼
        ┌──────────────────────────────┐
        │          Ziel-Repo           │
        │  CLAUDE.md   AGENTS.md        │
        │  .claude/settings.json        │
        │  .claude/commands/            │
        │  .claude/hooks/               │
        │  .mcp.json                    │
        │  memory-bank/                 │
        │  .forgecrate.yaml             │
        └──────────────────────────────┘
```

## Interne Pakete

Die Geschäftslogik liegt in `internal/`. Die Cobra-Commands in `cmd/forgecrate/`
sind dünn — sie validieren Argumente und delegieren.

| Paket | Zweck |
|---|---|
| `internal/config` | `.forgecrate.yaml` lesen/schreiben, `permission_mode` validieren |
| `internal/github` | Tarball-Download via GitHub-API, Retry-Logik bei Netzwerkfehlern |
| `internal/compose` | Layer-Composition: Markdown-Merge, JSON-Deep-Merge, Skills-Kopie |
| `internal/deploy` | Datei-Deployment, SHA256-Hashes, Konflikt-Resolution, Permission-Mode-Patch |
| `internal/extensions` | Parsen von `extensions.yaml`, Installation von Plugins und MCP-Servern via `claude` CLI |

## Single Source of Truth

| Quelle | Versorgt |
|---|---|
| `base/extensions.yaml` (+ `profiles/*/extensions.yaml`, `flavors/*/extensions.yaml`) | Plugins, MCP-Server, generiert `.mcp.json` |
| `base/models.yaml` | Canonical Claude-Model-IDs (per CI `make check-model-ids` enforced) |
| `internal/config/config.go` (`Config`-Struct) | Felder von `.forgecrate.yaml` |

## Datenfluss (`forgecrate init` / `update`)

1. **Download** — `internal/github` zieht das Tarball für `github.com/jmt-labs/forgecrate@<ref>` in ein temporäres Verzeichnis
2. **Compose** — `internal/compose` führt base → profile → flavors zusammen:
   - Markdown-Dateien werden konkateniert, CUSTOM-Blöcke bleiben erhalten
   - `settings.json` per Deep-Merge zusammengeführt
   - Slash-Commands additiv kopiert
3. **Deploy** — `internal/deploy` schreibt die komponierten Dateien ins Ziel-Repo,
   speichert SHA256-Hashes in `.forgecrate.yaml`, behandelt Konflikte interaktiv
4. **Extensions installieren** — `internal/extensions` ruft das `claude`-CLI auf,
   um Plugins (`claude plugin install`) und MCP-Server (`claude mcp add`) zu
   registrieren

## Querverweise

- [Layer-System](layer-system.md) — wie die drei Layer zusammenspielen
- [Abläufe](flows.md) — init-, update- und enforcement-Flow im Detail
- [Hook-System](hooks.md) — UserPromptSubmit- und PreToolUse-Hooks
- [Profile & Flavors](profiles-flavors.md) — verfügbare Layer-Optionen
- [ADR: CLAUDE.md ownership](architecture-decisions/CLAUDE-md-ownership.md)
