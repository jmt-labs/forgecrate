# Layer-System

## Ebenen

| Layer | Quelle | Precedence |
|---|---|---|
| base | `base/` | Niedrigste — immer aktiv |
| profile | `profiles/<name>/` | Überschreibt base |
| flavor | `flavors/<name>/` | Überschreibt profile |
| override | `.claude/commands/overrides/` (Ziel-Repo) | Höchste — nie überschrieben |

## Kompositions-Regeln

### CLAUDE.md / AGENTS.md

Inhalte werden aneinandergehängt (base → profile → flavors).
Das Ergebnis landet im GENERATED-Block. Der CUSTOM-Block bleibt immer erhalten.

### settings.json

Deep JSON Merge. Objekte werden rekursiv gemergt. Arrays werden ersetzt.
`settings.override.json` im Ziel-Repo hat höchste Priorität.

### .claude/commands/

Alle Skill-Dateien werden additiv kopiert. Spätere Layer überschreiben frühere
gleichnamige Dateien. Dateien unter `overrides/` werden nie angefasst.

## Beispiel

```
base/CLAUDE.md          → "# Base..."
profiles/backend/CLAUDE.md → "## Backend..."
flavors/tdd/CLAUDE.md   → "## TDD..."

Ergebnis CLAUDE.md:
  <!-- GENERATED:BEGIN -->
  # Base...

  ## Backend...

  ## TDD...
  <!-- GENERATED:END -->

  <!-- CUSTOM:BEGIN -->
  [User-Inhalt]
  <!-- CUSTOM:END -->
```
