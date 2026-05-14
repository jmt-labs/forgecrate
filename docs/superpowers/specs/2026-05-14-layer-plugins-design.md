# Layer Plugins — Design

**Datum:** 2026-05-14
**Issue:** folgt

## Ziel

Zusätzliche Claude-Code-Plugins werden layer-spezifisch deklariert. Jeder Layer erhält nur die Plugins, die für seinen Kontext sinnvoll sind. Die Installation erfolgt automatisch über den bestehenden Extensions-Mechanismus (`base/extensions.yaml`, `profiles/*/extensions.yaml`, `flavors/*/extensions.yaml`).

## Plugin-Zuordnung

### Base Layer

Gilt für alle Repos. Plugins, die universell nützlich sind.

| Plugin | Source | Begründung |
|---|---|---|
| `superpowers` | `claude-plugins-official/superpowers` | Workflow-Skills (bereits vorhanden) |
| `commit-commands` | `claude-plugins-official/commit-commands` | Commit-Workflow-Unterstützung |
| `security-guidance` | `claude-plugins-official/security-guidance` | Security-Bewusstsein überall relevant |
| `claude-md-management` | `claude-plugins-official/claude-md-management` | CLAUDE.md aktuell halten |

### Profile: `frontend`

Für Frontend-Repos.

| Plugin | Source | Begründung |
|---|---|---|
| `frontend-design` | `claude-plugins-official/frontend-design` | Design-bewusste Frontend-Skills |
| `typescript-lsp` | `claude-plugins-official/typescript-lsp` | TypeScript LSP-Integration |
| `playwright` | `claude-plugins-official/playwright` (external) | Browser-Testing |

### Profile: `backend` und `fullstack`

Keine zusätzlichen Plugins in dieser Iteration. LSPs sind sprachspezifisch — sinnvoller über zukünftige sprachspezifische Flavors (z.B. `go`, `python`, `rust`).

### Flavor: `strict-review`

Für erhöhte Code-Review-Anforderungen.

| Plugin | Source | Begründung |
|---|---|---|
| `pr-review-toolkit` | `claude-plugins-official/pr-review-toolkit` | PR-Review-Workflows |
| `code-simplifier` | `claude-plugins-official/code-simplifier` | Code-Qualität vor dem Merge |

### Flavor: `tdd`

Keine zusätzlichen Plugins. `superpowers` deckt TDD-Skills vollständig ab.

### Flavor: `minimal`

Explizit keine Plugins — Minimalismus.

## Dateien

| Datei | Aktion |
|---|---|
| `base/extensions.yaml` | `commit-commands`, `security-guidance`, `claude-md-management` ergänzen |
| `profiles/frontend/extensions.yaml` | Neu anlegen mit `frontend-design`, `typescript-lsp`, `playwright` |
| `flavors/strict-review/extensions.yaml` | Neu anlegen mit `pr-review-toolkit`, `code-simplifier` |

## Nicht in Scope

- LSP-Plugins für Backend-Sprachen (Go, Python, Rust etc.) — folgen mit sprachspezifischen Flavors
- Externe Plugins wie `linear`, `context7` — separates Feature wenn Bedarf entsteht
- Flavors `tdd` und `minimal` — keine Änderung
