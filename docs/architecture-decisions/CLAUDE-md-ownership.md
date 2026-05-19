# CLAUDE.md Ownership

## Root `CLAUDE.md`

**Owner: manually maintained by the repository user.**

This file is the user-facing override layer. It is written once during `forgecrate init` and never overwritten by subsequent `forgecrate update` runs. Users place project-specific instructions, team conventions and custom workflow rules here (inside `<!-- CUSTOM:BEGIN / END -->` blocks or directly in the file).

The `<!-- GENERATED:BEGIN / END -->` markers in the root CLAUDE.md are a legacy artifact from an earlier convention that has since dissolved. The file is effectively fully manual now.

## `base/CLAUDE.md`

**Owner: forgecrate base layer — regenerated on every `forgecrate update`.**

This file ships as part of the forgecrate base layer and contains the default workflow instructions that apply to every repository using forgecrate. It is replaced on each update without prompting (unless the hash-tracking mechanism detects a local modification, in which case the standard conflict flow applies).

## Summary

| File | Ownership | Updated by |
|---|---|---|
| `CLAUDE.md` (root) | User / project team | Manual edits only |
| `base/CLAUDE.md` | forgecrate base layer | `forgecrate update` |
