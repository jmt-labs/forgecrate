# Model & Subagent Wiring — Current State and Deferred Work

## Status

`base/models.yaml` and the "Team-Rollen & Subagent-Konfiguration" table in
`base/CLAUDE.md` are **reference documentation**, not active configuration. This
note records what is wired today and what is deliberately deferred because it
requires Go changes.

## What is real today

- The session model is the literal `"model"` field in `base/.claude/settings.json`.
  It is global (one model per session) and is the only model setting forgecrate
  actually deploys.
- `base/models.yaml` is a human-readable reference of which concrete model version
  each role maps to. No Go code reads it, and it is not deployed into target repos.
- The role table prescribes `opus`/`sonnet`/`haiku` — the family aliases that the
  Agent/Task tool actually accepts as a `model` parameter. These are advisory: the
  main agent chooses them when dispatching subagents; nothing applies them
  automatically.

## Deferred (requires Go — out of scope for content-only changes)

1. **Propagate `models.yaml` → `settings.json`.** Add a YAML reader + template step
   in `internal/compose` so the deployed `settings.json` `model` derives from
   `models.yaml` instead of being a hardcoded literal. This would make the
   "change only models.yaml on upgrade" promise true.
2. **Generate & deploy `.claude/agents/*.md` subagent definitions.** The Agent tool
   reads per-subagent definition files with `model`/`description` frontmatter. To
   turn the role table into enforced config, forgecrate needs a new source dir and a
   copy step in `internal/deploy/deploy.go`. No such files are produced or copied
   today, so the role table cannot be machine-enforced.
3. **Deploy `models.yaml` itself.** Only needed if deployed repos should reference
   it. Today the deployed `base/CLAUDE.md` deliberately does NOT reference
   `models.yaml`, because the file is absent in consumer repos.

## Why this is documented here

`docs/` is excluded from the `check-model-ids` CI guard and is never deployed into
target repos, so it is the correct home for forgecrate-internal backlog items.
These notes must not be added to `base/CLAUDE.md`, `base/models.yaml` headers, or
any other deployed file — that would pollute every consumer session with
forgecrate-internal TODOs.
