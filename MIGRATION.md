# Migration: claude-setup → forgecrate

This project was renamed from `claude-setup` to `forgecrate`. This document
describes what existing users need to do.

## Binary name

The CLI binary is now called `forgecrate`. The old name `claude-setup` still
works for **one minor version** and prints a deprecation warning:

```
Warning: 'claude-setup' is deprecated, use 'forgecrate' instead.
```

**Action required:** Update any scripts, CI pipelines, or aliases that call
`claude-setup` to use `forgecrate`.

## Config file

The deployment state file was renamed:

| Old | New |
|---|---|
| `.claude-setup.yaml` | `.forgecrate.yaml` |

**Automatic migration:** On the first `forgecrate init` or `forgecrate update`
run, the CLI automatically renames `.claude-setup.yaml` to `.forgecrate.yaml`
and prints:

```
Notice: migrating .claude-setup.yaml → .forgecrate.yaml
```

No manual action needed for the config file — the CLI handles it.

## Skill and command names

All built-in slash-commands were renamed:

| Old | New |
|---|---|
| `/claude-setup-advisor` | `/forgecrate-advisor` |
| `/claude-setup-release` | `/forgecrate-release` |
| `/claude-setup-repo-health` | `/forgecrate-repo-health` |
| `/claude-setup-repo-onboarding` | `/forgecrate-repo-onboarding` |
| `/claude-setup-test-coverage` | `/forgecrate-test-coverage` |
| `/claude-setup-pr-checklist` | `/forgecrate-pr-checklist` |
| `/claude-setup-db-migration` | `/forgecrate-db-migration` |
| `/claude-setup-getbetter` | `/forgecrate-getbetter` |
| `/claude-setup-handoff` | `/forgecrate-handoff` |
| `/claude-setup-gitops-status` | `/forgecrate-gitops-status` |

**Action required:** Run `forgecrate update` to deploy the renamed commands into
existing repositories.

## Installation

Update your package-manager install command:

```sh
# Homebrew
brew upgrade forgecrate

# apt
sudo apt update && sudo apt install forgecrate

# go install
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest

# curl
curl -fsSL https://raw.githubusercontent.com/jmt-labs/forgecrate/main/install.sh | bash
```

## Go module path (if imported as a library)

```go
// Old
import "github.com/jmt-labs/claude-setup/internal/..."

// New
import "github.com/jmt-labs/forgecrate/internal/..."
```

## GitHub repository

The repository URL changes to `https://github.com/jmt-labs/forgecrate` after the
admin rename. GitHub automatically redirects
`https://github.com/jmt-labs/claude-setup` — existing clone URLs continue to
work.

## mem0 → memory-bank

forgecrate uses the [`@allpepper/memory-bank-mcp`](https://www.npmjs.com/package/@allpepper/memory-bank-mcp)
MCP server instead of the previous mem0 plugin.

Project context now lives in `memory-bank/*.md` (versioned, committed) and is
read/written exclusively via MCP tools (`mcp__memory-bank__memory_bank_read`,
`mcp__memory-bank__memory_bank_write`, `mcp__memory-bank__memory_bank_update`).

**Manual steps after `forgecrate update`:**

1. Disable the mem0 plugin in Claude Code if it was previously enabled:
   - Open Claude Code → Settings → Plugins → disable `mem0`
   - Or in `~/.claude/settings.json`: set `"mem0@mem0-plugins": false`
2. Populate `memory-bank/` files with your project context. Seed files with
   empty sections are deployed automatically.

The cross-project `memory` MCP server (file: `.claude/memory.json`) is
**unrelated** and continues to be used for timeless architecture decisions.
