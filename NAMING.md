# Naming: forgecrate

## Decision

The project was renamed from `claude-setup` to **`forgecrate`** on 2026-05-19.

**Rationale:** `forge` (smithy — shaping raw material into something useful) + `crate` (a container of components, also a Rust package unit with dev-tool connotations). The name shares the `forge` root with the sister project `forgedeck`, signaling the same product family, while `crate` distinguishes its purpose: a deployable bundle of Claude Code configuration. `forgecrate init` reads naturally as a command.

## Availability matrix (checked 2026-05-19)

| Candidate | Chars | GitHub | npm | PyPI | Result |
|---|---|---|---|---|---|
| `forgekit` | 8 | ✅ | ❌ | ❌ | Eliminated |
| `anvilkit` | 8 | ✅ | ✅ | ✅ | Top-3 |
| `hearthkit` | 9 | ✅ | ✅ | ✅ | Top-3 |
| `smithkit` | 8 | ✅ | ✅ | ✅ | — |
| `forgevault` | 10 | ✅ | ✅ | ❌ | Eliminated |
| **`forgecrate`** | 10 | ✅ | ✅ | ✅ | **Selected** |
| `anvilcrate` | 10 | ✅ | ✅ | ✅ | — |
| `millkit` | 7 | ✅ | ✅ | ✅ | — |
| `lathekit` | 8 | ✅ | ✅ | ✅ | — |
| `forgeseed` | 9 | ✅ | ✅ | ✅ | — |
| `anvilstack` | 10 | ✅ | ✅ | ✅ | — |
| `kilnkit` | 7 | ✅ | ✅ | ✅ | — |
| `smithcrate` | 10 | ✅ | ✅ | ✅ | — |
| `millcrate` | 9 | ✅ | ✅ | ✅ | — |
| `hearthstack` | 11 | ✅ | ✅ | ✅ | — |

*Note: This is a Go binary distributed via Homebrew/apt/Chocolatey. npm/PyPI conflicts are technically irrelevant but were checked for completeness.*
