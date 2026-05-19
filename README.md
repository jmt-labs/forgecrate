<div align="center">
  <img src="assets/banner.svg" alt="forgecrate — Reproducible Claude Code configuration" width="100%">
</div>

> **Formerly known as `claude-setup`.** See [MIGRATION.md](MIGRATION.md) for upgrade notes.

[![Latest Release](https://img.shields.io/github/v/release/jmt-labs/forgecrate)](https://github.com/jmt-labs/forgecrate/releases/latest)
[![CI](https://github.com/jmt-labs/forgecrate/actions/workflows/ci.yml/badge.svg)](https://github.com/jmt-labs/forgecrate/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# forgecrate

**forgecrate** installs a reproducible [Claude Code](https://claude.ai/code) configuration into any Git repository. A single Go binary downloads profiles, flavors, hooks and MCP server definitions from GitHub and composes them through a layering system into the target repository.

```sh
forgecrate init --profile backend --flavors tdd,strict-review
```

Claude Code is immediately ready to use: with workflow enforcement, slash-commands, branch protection and five pre-integrated MCP servers — no manual configuration needed.

> Part of the jmt-labs toolchain → [forgedeck](https://github.com/jmt-labs/forgedeck)

---

## Contents

- [What gets deployed?](#what-gets-deployed)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [The layer system](#the-layer-system)
- [Profiles](#profiles)
- [Flavors](#flavors)
- [CLI reference](#cli-reference)
- [Customization & local overrides](#customization--local-overrides)
- [Skills & slash-commands](#skills--slash-commands)
- [MCP servers](#mcp-servers)
- [Update conflicts](#update-conflicts)
- [Further documentation](#further-documentation)
- [Development & tests](#development--tests)

---

## What gets deployed?

After `forgecrate init`, the following files are present in the repository:

| File / Directory | Purpose |
|---|---|
| `CLAUDE.md` | Behavior rules and workflow for Claude Code |
| `AGENTS.md` | Guidelines for Claude Code agents |
| `.claude/settings.json` | Model, hooks, plugins, permissions, MCP servers |
| `.claude/commands/` | Slash-commands (skills) |
| `.claude/hooks/` | Pre-tool and UserPromptSubmit hooks |
| `.forgecrate.yaml` | Deployment state (profile, flavors, file hashes) |

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew tap jmt-labs/tap
brew install forgecrate
```

### Chocolatey (Windows)

```sh
choco install forgecrate
```

### apt (Ubuntu / Debian)

```sh
curl -fsSL https://jmt-labs.github.io/apt/KEY.gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/jmt-labs.gpg

echo "deb [signed-by=/etc/apt/keyrings/jmt-labs.gpg] https://jmt-labs.github.io/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/jmt-labs.list

sudo apt update && sudo apt install forgecrate
```

### go install

```sh
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest
```

### curl (without package manager)

```sh
curl -fsSL https://raw.githubusercontent.com/jmt-labs/forgecrate/main/install.sh | bash
```

Install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/jmt-labs/forgecrate/main/install.sh | bash -s v1.0.0
```

---

## Quickstart

**1. Initialize a repository**

```sh
cd my-project
forgecrate init --profile backend --flavors tdd,strict-review
```

**2. Update to the latest version**

```sh
forgecrate update
```

**3. Switch profile**

```sh
forgecrate update --profile fullstack
```

**4. List available options**

```sh
forgecrate list profile    # all profiles
forgecrate list flavor     # all flavors
forgecrate describe profile backend   # details for a profile
```

---

## The layer system

`forgecrate` composes configuration from three stacked layers. Each layer extends or overrides the one below:

```
┌─────────────────────────────┐
│  overrides (local, optional) │  Highest priority — never overwritten
├─────────────────────────────┤
│  flavors   (multiple)        │  e.g. tdd + strict-review
├─────────────────────────────┤
│  profile   (one)             │  e.g. backend
├─────────────────────────────┤
│  base      (always active)   │  Lowest priority — always deployed
└─────────────────────────────┘
```

**Composition rules:**

| File type | Strategy |
|---|---|
| `CLAUDE.md` / `AGENTS.md` | Text blocks from layers are concatenated |
| `.claude/settings.json` | Deep-merge — deeper layers win on conflicts |
| `.claude/commands/` | Additive copy — later layers overwrite earlier ones |
| `.claude/hooks/` | Additive copy |

Layer contents come directly from this repository under `base/`, `profiles/` and `flavors/`.

---

## Profiles

A profile sets the technical context of the project. Exactly **one** is active per repository.

### `backend`

Optimized for server applications, REST APIs and database access.

- API design: REST-first, clearly named error codes
- Database access: type-safe, exclusively parameterized queries
- Prefers integration tests over pure unit tests with mocks
- No ORM magic — explicit queries are more readable

### `frontend`

Optimized for component-based UI development.

- Component architecture and state management
- Accessibility as a first-class requirement
- Visual regression and snapshot tests

### `fullstack`

Combines backend and frontend context in one profile.

- Shared types between client and server
- End-to-end tests across the full stack
- Clear boundary between API contract and implementation

---

## Flavors

Flavors add cross-cutting practices to a profile. **Multiple flavors** can be active simultaneously.

### `tdd` — Test-Driven Development

Enforces the classic Red-Green-Refactor cycle:

1. Write test → run (must fail)
2. Minimal implementation → test must pass
3. Refactoring → tests must continue to pass
4. Commit

Rules:
- No production code without a prior test
- Test names describe behavior, not implementation
- Mocks only at system boundaries (external APIs, databases)
- Before every bug fix: write a regression test

### `strict-review` — Mandatory code review

Enforces structured review before every commit:

- Mandatory call of `superpowers:requesting-code-review` before every commit
- No direct commits to `main` / `master`
- PR description contains: what changed, why, how tested
- Breaking changes are explicitly communicated

### `minimal`

Only basic workflow enforcement. Suitable for projects that prefer a lightweight start.

### `gitops`

For Infrastructure-as-Code and GitOps workflows:

- ArgoCD app topology is respected
- Cluster-wide rules (Kyverno, Gatekeeper, RULES.md) are enforced
- Deployments exclusively via ArgoCD — no direct `kubectl apply` commands
- Skill: `/forgecrate-gitops-status` for cluster overview

### `getbetter`

Enables continuous learning across sessions: reads `.claude/GETBETTER.md` at session start and saves insights at session end via `/forgecrate-getbetter`.

### `github`

Adds GitHub-specific workflow rules: releases via `gh release create`, CI status checks before tagging, and proactive parallelization for multi-step GitHub tasks.

---

## CLI reference

### `forgecrate init`

Performs first-time configuration of the repository.

```sh
forgecrate init [--profile <name>] [--flavors <name,name,...>]
```

| Flag | Default | Description |
|---|---|---|
| `--profile` | `backend` | Active profile |
| `--flavors` | _(none)_ | Comma-separated list of active flavors |

**Flow:** Download tarball from GitHub → compose layers → write files → create `.forgecrate.yaml`.

---

### `forgecrate update`

Updates the configuration to the latest version of the upstream repository.

```sh
forgecrate update [--profile <name>]
```

| Flag | Default | Description |
|---|---|---|
| `--profile` | _(current)_ | Switch profile |

When conflicts exist between local changes and upstream, an interactive prompt appears (see [Update conflicts](#update-conflicts)).

---

### `forgecrate list`

Lists all available profiles or flavors.

```sh
forgecrate list profile
forgecrate list flavor
```

---

### `forgecrate describe`

Shows a detailed description of a profile or flavor.

```sh
forgecrate describe profile backend
forgecrate describe flavor tdd
```

---

## Customization & local overrides

All local customizations are **preserved** on `forgecrate update`. Three mechanisms are available:

### CUSTOM blocks in Markdown

In `CLAUDE.md` and `AGENTS.md`, custom instructions can be written in a protected block:

```markdown
<!-- CUSTOM:BEGIN -->
Project-specific instructions here,
never overwritten by updates.
<!-- CUSTOM:END -->
```

### Settings overrides

```json
// .claude/overrides/settings.override.json
{
  "permissions": {
    "allow": ["Bash(make:*)"]
  }
}
```

This file is merged with the highest priority.

### Custom slash-commands

```
.claude/commands/overrides/my-command.md
```

Custom skills in this directory complement the managed commands and are never overwritten.

---

## Skills & slash-commands

`forgecrate` deploys a set of predefined skills, available directly in Claude Code as slash-commands:

| Command | Description |
|---|---|
| `/forgecrate-advisor` | Analyzes the repo and recommends the right profile + flavors |
| `/forgecrate-repo-onboarding` | Creates a structured codebase overview for `CLAUDE.md` |
| `/forgecrate-repo-health` | Finds improvement potential and delivers a prioritized list |
| `/forgecrate-test-coverage` | Analyzes test coverage and suggests the next concrete test |
| `/forgecrate-pr-checklist` | Systematic review before `gh pr create` |
| `/forgecrate-db-migration` | Guides creation and review of a database migration |
| `/forgecrate-release` | Runs a complete release cycle |
| `/forgecrate-handoff` | Generates a `HANDOFF.md` with portable project context for AI model switches or session handoffs |

Via [Superpowers skills](https://github.com/anthropics/claude-code-superpowers), additional mandatory skills are available and automatically integrated into the development workflow (brainstorming, TDD, code review, debugging).

---

## MCP servers

The base configuration includes five pre-integrated MCP servers:

| Server | Transport | Purpose |
|---|---|---|
| `github` | HTTP (GitHub Copilot) | Issues, PRs, code search, branches, labels |
| `fetch` | stdio (`npx`) | External web content — docs, RFCs, changelogs |
| `memory` | stdio (`npx`) | Persistent cross-session knowledge (`.claude/memory.json`) |
| `context-mode` | stdio (`npx`) | Automatic context optimization and session history search |
| `context7` | stdio (`npx`) | Current library documentation directly from source repos |

**Requirement for `github`:** The `GITHUB_PERSONAL_ACCESS_TOKEN` environment variable must be set.

**Requirement for stdio servers:** Node.js / `npx` must be available in PATH.

---

## Update conflicts

During `forgecrate update`, the tool compares each managed file against a SHA256 hash stored at the last deploy.

A **conflict** only arises when **both** conditions are met:
- The local file was manually changed since the last deploy **and**
- the new upstream version differs from the local version

In that case, an interactive prompt appears:

```
KONFLIKT: .claude/settings.json
  Deine Version: { "model": "claude-opus-4-7", ...
  Neue Version:  { "model": "claude-sonnet-4-6", ...
  [ü]berschreiben / [b]ehalten (Standard: behalten):
```

| Input | Result |
|---|---|
| `ü` or `u` | Take upstream version, local changes are lost |
| `b` or Enter | Keep local version, local hash becomes the new baseline |

**Recommendation:** Put custom changes in CUSTOM blocks or override files — then no conflicts arise.

---

## Further documentation

| Topic | Document |
|---|---|
| Architecture and components | [docs/architecture.md](docs/architecture.md) |
| Layer system (composition rules) | [docs/layer-system.md](docs/layer-system.md) |
| Flows: init, update, enforcement | [docs/flows.md](docs/flows.md) |
| Hook reference | [docs/hooks.md](docs/hooks.md) |
| Profiles & flavors (details) | [docs/profiles-flavors.md](docs/profiles-flavors.md) |
| Development & tests | [docs/development.md](docs/development.md) |
| Migration from `claude-setup` | [MIGRATION.md](MIGRATION.md) |

---

## Development & tests

**Prerequisites:** Go 1.22+, `make`

```sh
# Unit tests
make test

# End-to-end tests
make test-e2e

# Code quality check
make quality

# Build binary locally
make build

# Release (requires GoReleaser + GitHub token)
make release
```

**Repository structure:**

| Path | Purpose |
|---|---|
| `cmd/forgecrate/` | CLI entry point (Cobra) |
| `internal/compose/` | Markdown, JSON and skills merge logic |
| `internal/deploy/` | Deployment orchestration and conflict handling |
| `internal/config/` | `.forgecrate.yaml` read/write |
| `internal/github/` | GitHub API client (tarball download) |
| `internal/extensions/` | Claude extension handling |
| `base/` | Base layer — always deployed |
| `base/models.yaml` | Canonical model IDs — single source of truth for all agent roles |
| `profiles/` | Profile layer — one selectable |
| `flavors/` | Flavor layer — multiple combinable |
| `e2e/` | End-to-end tests |

Contributions welcome — please open an issue and follow the development workflow from [CLAUDE.md](CLAUDE.md).

---

## License

[MIT](LICENSE) — Copyright (c) jmt-labs
