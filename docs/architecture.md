# Architektur

## Komponenten

```
┌─────────────────────────────────────────────────────┐
│                   forgecrate Repo                  │
│  base/ · profiles/ · flavors/ · cmd/forgecrate/   │
└───────────────────────┬─────────────────────────────┘
                        │ GitHub API (tarball)
                        ▼
              ┌─────────────────┐
              │  forgecrate   │  ← globales Go-Binary
              │    Binary       │
              └────────┬────────┘
                       │ compose + deploy
                       ▼
        ┌──────────────────────────────┐
        │          Ziel-Repo           │
        │  CLAUDE.md  AGENTS.md        │
        │  .claude/settings.json       │
        │  .claude/commands/           │
        │  .claude/hooks/              │
        │  .forgecrate.yaml          │
        └──────────────────────────────┘
```

## Layer-System

```
Layer 1: base/          → immer aktiv
Layer 2: profiles/<p>/  → eines wählbar
Layer 2: flavors/<f>/   → mehrere kombinierbar
Layer 3: overrides/     → lokal, nie überschrieben
```
