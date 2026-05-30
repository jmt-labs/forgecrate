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

Claude Code is immediately ready to use: workflow enforcement, slash-commands, branch protection and six pre-integrated MCP servers — no manual configuration required.

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
| `.mcp.json` | Generated MCP server configuration |
| `memory-bank/` | Structured project memory (read/written via `memory-bank` MCP) |
| `.forgecrate.yaml` | Deployment state: profile, flavors, permission mode, file hashes |

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

**3. Reconfigure profile and flavors interactively**

```sh
forgecrate config
```

**4. List or describe available options**

```sh
forgecrate list                       # all profiles and flavors
forgecrate describe profile backend   # details for a profile
forgecrate describe flavor tdd        # details for a flavor
```

---

## The layer system

`forgecrate` composes configuration from four stacked layers. Each layer extends or overrides the one below:

```
┌──────────────────────────────────────────────┐
│  overrides (local, optional)                  │  Highest priority — never overwritten
├──────────────────────────────────────────────┤
│  flavors  (multiple, additive)                │  e.g. tdd + strict-review
├──────────────────────────────────────────────┤
│  profile  (exactly one)                       │  e.g. backend
├──────────────────────────────────────────────┤
│  base     (always active)                     │  Lowest priority — always deployed
└──────────────────────────────────────────────┘
```

**Composition rules:**

| File type | Strategy |
|---|---|
| `CLAUDE.md` / `AGENTS.md` | Text blocks from layers are concatenated; user-`CUSTOM` blocks are preserved |
| `.claude/settings.json` | Deep-merge — deeper layers win on conflicts |
| `.claude/commands/` | Additive copy — later layers overwrite earlier ones |
| `.claude/hooks/` | Additive copy |
| `extensions.yaml` (plugins + MCP) | Merged across layers, generates `.mcp.json` |

Layer contents come directly from this repository under `base/`, `profiles/` and `flavors/`.

---

## Profiles

A profile sets the technical context of the project. Exactly **one** profile is active per repository.

### `backend`

Optimized for server applications, REST APIs and database access.

- API design: REST-first, clearly named error codes
- Database access: type-safe, exclusively parameterized queries
- Prefers integration tests over pure unit tests with mocks
- No ORM magic — explicit queries are more readable
- Extra skill: `/forgecrate-db-migration`

### `frontend`

Optimized for component-based UI development.

- Small, focused components with a single responsibility
- Semantic HTML, ARIA attributes where required (accessibility first)
- Behavior tests over implementation tests
- Extra plugins: `frontend-design`, `typescript-lsp`, `playwright`
- Extra MCP: `playwright` (browser automation)
- Extra skills: `/accessibility-audit`, `/ui-ux-audit`

### `fullstack`

Combines backend and frontend context in one profile.

- Shared types between client and server
- End-to-end tests across the full stack
- Clear boundary between API contract and implementation
- Extra MCP: `playwright`

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

Extra skill: `/forgecrate-test-coverage`.

### `strict-review` — Mandatory code review

Enforces structured review before every commit:

- Mandatory call of `superpowers:requesting-code-review` before every commit
- No direct commits to `main` / `master`
- PR description contains: what changed, why, how tested
- Breaking changes are explicitly communicated

Extra plugins: `pr-review-toolkit`, `code-simplifier`. Extra skill: `/forgecrate-pr-checklist`.

### `minimal` — Lightweight start

Adds no extra mandatory skills. Suitable for prototypes, solo projects or early-stage exploration. The base layer remains fully active — `minimal` is purely an explicit "no extras" signal that combines cleanly with other flavors.

### `gitops` — Infrastructure-as-Code

For ArgoCD-driven GitOps workflows:

- ArgoCD app topology is loaded on session start; rules from `RULES.md`, Kyverno/OPA `ClusterPolicy` and `Constraint` manifests are enforced
- Deployments **exclusively** via ArgoCD — direct `kubectl apply`, `helm install/upgrade` etc. are blocked unless explicitly confirmed
- No `latest` image tags — pinned versions or digests required
- Extra skill: `/forgecrate-gitops-status` (drift check, policy validation)

### `getbetter` — Continuous learning across sessions

Reads `.claude/GETBETTER.md` at session start and saves insights via
`/forgecrate-getbetter` at session end. The file collects recurring mistakes,
patterns that worked well, and project-specific gotchas not visible in the code.

### `github` — GitHub-native workflow

Adds GitHub-specific rules: releases via `gh release create`, CI status checks
before tagging, proactive parallelization for multi-step GitHub tasks. Extra
skills: `/forgecrate-issue-resolver`, `/forgecrate-github-release`.

### `force-research` — Enforced web research

Hardens the base layer's research mandate. By default every role must use a research
tool (`WebSearch` / `WebFetch` / `context7` / `fetch`) before any non-trivial code
change, enforced by the `pre-tool.sh` hook (`forgecrate hook require-research`):
`Edit` / `Write` / `MultiEdit` are blocked until a research tool call is present
anywhere in the session transcript. This flavor extends that block to **file-writing
Bash commands** (`sed -i`, `tee`, `dd of=`, redirects outside `/tmp`), closing the
"write via shell instead of Edit/Write" loophole. One research call per session unblocks
all subsequent edits in that session.

### `no-research` — Opt-out from research mandate

Disables the default research mandate from the base layer, **including the hard
PreToolUse block**: `Edit` / `Write` / `MultiEdit` are no longer bound to a prior
research call. By default every role must use `WebSearch` / `context7` / `fetch`
before a non-trivial code change. Enable this flavor for air-gapped repositories,
strict compliance environments, or projects with purely internal logic where
external research is not applicable.

### `codegraph` — Semantic code knowledge graph

Adds a local [codegraph](https://github.com/colbymchenry/codegraph) MCP server that builds a semantic knowledge graph of the repository. Provides tools for AI agents:

| Tool | Purpose |
|---|---|
| `codegraph_search` | Semantic code search without exact keywords |
| `codegraph_node` | Retrieve definition of a symbol (function, type, variable) |
| `codegraph_callers` / `codegraph_callees` | Find all callers / callees of a symbol |
| `codegraph_trace` | Trace the call path between two symbols |
| `codegraph_explore` | Explore dependencies and neighbours of a symbol |
| `codegraph_context` | Explain a code section with graph context |
| `codegraph_impact` | Determine the blast radius of a change |
| `codegraph_files` | List files in the index |
| `codegraph_status` | Check index status |

The index is updated in the background at session start. The `.codegraph/` directory is automatically added to `.gitignore`.

**Prerequisite:** `codegraph` must be installed:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex

# Alternatively via npm
npm i -g @colbymchenry/codegraph
```

Then initialise in the repository: `codegraph init -i`

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

**Flow:** Download tarball from GitHub → compose layers → write files → install plugins and MCP servers via `claude` CLI → create `.forgecrate.yaml`.

The alias `forgecrate run` is accepted for backwards compatibility.

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

### `forgecrate config`

Reconfigures profile and flavors interactively, then re-runs deploy.

```sh
forgecrate config
```

Opens a TUI selector (powered by [Charmbracelet Huh](https://github.com/charmbracelet/huh)) with the currently active profile and flavors pre-selected. Requires an existing `.forgecrate.yaml` (run `forgecrate init` first).

---

### `forgecrate list`

Lists all available profiles and flavors fetched from the upstream repository.

```sh
forgecrate list
```

---

### `forgecrate describe`

Shows a detailed description of a profile or flavor — prints the layer's `CLAUDE.md` content plus the included skills.

```sh
forgecrate describe profile backend
forgecrate describe flavor tdd
```

---

### `forgecrate set-permission-mode`

Sets the Claude Code agent permission mode and patches `.claude/settings.json` accordingly.

```sh
forgecrate set-permission-mode <bypass|plan|ask|auto>
```

The setting is persisted to `.forgecrate.yaml` (`permission_mode:` key) and survives the next `forgecrate update`.

---

### `forgecrate hook prompt-submit`

Helper invoked by the `UserPromptSubmit` hook. Prints the active profile, active flavors and the mandatory-skill checklist. You will rarely call this manually — it is wired up by `.claude/hooks/prompt-submit.sh`.

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

`forgecrate` deploys a set of predefined skills, available directly in Claude Code as slash-commands. The set depends on which layers are active.

| Command | Layer | Description |
|---|---|---|
| `/forgecrate-advisor` | base | Analyzes the repo and recommends a profile + flavors |
| `/forgecrate-repo-onboarding` | base | Creates a structured codebase overview for `CLAUDE.md` |
| `/forgecrate-repo-health` | base | Finds improvement potential and delivers a prioritized list |
| `/forgecrate-release` | base | Runs a complete release cycle |
| `/forgecrate-handoff` | base | Generates a portable `HANDOFF.md` for AI model switches or session handoffs |
| `/forgecrate-catchup` | base | Short digest of recent activity (commits, context, GitHub) over the last N days |
| `/forgecrate-db-migration` | profile: `backend` | Guides creation and review of a database migration |
| `/accessibility-audit` | profile: `frontend` | Static A11y checks per changed file (alt, label, aria-\*) |
| `/ui-ux-audit` | profile: `frontend` | Deep UI/UX audit with severity grading and auto-created GitHub issues |
| `/forgecrate-test-coverage` | flavor: `tdd` | Analyzes test coverage and suggests the next concrete test |
| `/forgecrate-pr-checklist` | flavor: `strict-review` | Systematic review before `gh pr create` |
| `/forgecrate-issue-resolver` | flavor: `github` | End-to-end resolution of a GitHub issue up to merge-ready PR |
| `/forgecrate-batch-issues` | flavor: `github` | Assigns and resolves up to 5 issues in parallel via sub-agents, one PR each |
| `/forgecrate-github-release` | flavor: `github` | Creates a GitHub Release for the latest tag |
| `/forgecrate-gitops-status` | flavor: `gitops` | Drift check, policy validation, deployment status |
| `/forgecrate-getbetter` | flavor: `getbetter` | Saves session insights into `.claude/GETBETTER.md` |

Via [Superpowers skills](https://github.com/anthropics/claude-code-superpowers), additional mandatory skills are available and automatically integrated into the development workflow (brainstorming, TDD, code review, debugging).

---

## MCP servers

The base configuration includes six pre-integrated MCP servers. Profiles can add more (e.g. `playwright` for `frontend`/`fullstack`).

| Server | Transport | Purpose |
|---|---|---|
| `github` | HTTP (GitHub Copilot) | Issues, PRs, code search, branches, labels |
| `fetch` | stdio (`npx`) | External web content — docs, RFCs, changelogs |
| `memory` | stdio (`npx`) | Persistent cross-project architecture decisions (`.claude/memory.json`) |
| `memory-bank` | stdio (`npx`) | Repo-specific structured project memory (`memory-bank/*.md`) |
| `context-mode` | stdio (`npx`) | Automatic context optimization and session history search |
| `context7` | stdio (`npx`) | Current library documentation directly from source repos |

**Requirement for `github`:** the `GITHUB_PERSONAL_ACCESS_TOKEN` environment variable must be set.

**Requirement for stdio servers:** Node.js / `npx` must be available in `PATH`.

The single source of truth for MCP configuration is `base/extensions.yaml` (plus profile/flavor `extensions.yaml`). The deployed `.mcp.json` is generated from these — edit the YAML, then re-run `forgecrate update`.

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
  [o]verwrite / [k]eep (default: keep):
```

| Input | Result |
|---|---|
| `o` or `u` / `ü` | Take upstream version; local changes are lost |
| `k` or `b` or Enter | Keep local version; the local hash becomes the new baseline |

`u`/`ü` and `b` are kept for backwards compatibility with earlier versions.

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
| ADR: CLAUDE.md ownership | [docs/architecture-decisions/CLAUDE-md-ownership.md](docs/architecture-decisions/CLAUDE-md-ownership.md) |
| Migration from `claude-setup` | [MIGRATION.md](MIGRATION.md) |
| Naming rationale | [NAMING.md](NAMING.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |

---

## Development & tests

**Prerequisites:** Go 1.24+, `make`, optionally `goreleaser` for releases.

```sh
make test                # Unit + integration tests
make test-e2e            # End-to-end tests (uses CLAUDE_BIN or a fake binary)
make quality             # go vet + go build sanity check
make build               # Build ./forgecrate binary locally
make check-model-ids     # Enforce: model IDs live only in base/models.yaml
make check-readme-coverage  # Verify every flavor is mentioned in this README
make release             # GoReleaser release (requires GitHub token)
make clean               # Clean build artifacts and test cache
```

**Repository structure:**

| Path | Purpose |
|---|---|
| `cmd/forgecrate/` | CLI entry point (Cobra) — one file per subcommand |
| `internal/compose/` | Markdown, JSON and skills merge logic |
| `internal/config/` | `.forgecrate.yaml` read/write, permission-mode validation |
| `internal/deploy/` | Deployment orchestration and conflict handling |
| `internal/extensions/` | Plugin and MCP server installation via `claude` CLI |
| `internal/github/` | GitHub API client (tarball download with retries) |
| `base/` | Base layer — always deployed |
| `base/extensions.yaml` | Single source of truth for plugins + MCP servers |
| `base/models.yaml` | Canonical Claude model IDs |
| `profiles/` | Profile layer — exactly one selectable per repo |
| `flavors/` | Flavor layer — multiple combinable per repo |
| `e2e/` | End-to-end tests |

Contributions welcome — please open an issue and follow the development workflow from [CLAUDE.md](CLAUDE.md).

---

## License

[MIT](LICENSE) — Copyright (c) jmt-labs
