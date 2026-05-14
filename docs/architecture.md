# Architektur

## Komponenten

```
┌─────────────────────────────────────────────────────┐
│                   claude-setup Repo                  │
│  base/ · profiles/ · flavors/ · cmd/claude-setup/   │
└───────────────────────┬─────────────────────────────┘
                        │ GitHub API (tarball)
                        ▼
              ┌─────────────────┐
              │  claude-setup   │  ← globales Go-Binary
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
        │  .claude-setup.yaml          │
        └──────────────────────────────┘
```

## Layer-System

```
Layer 1: base/          → immer aktiv
Layer 2: profiles/<p>/  → eines wählbar
Layer 2: flavors/<f>/   → mehrere kombinierbar
Layer 3: overrides/     → lokal, nie überschrieben
```
