# Layer-System

forgecrate komponiert die Ziel-Konfiguration aus vier Layern. Jeder Layer kann
Dateien hinzufügen oder vorherige überschreiben.

## Ebenen

| Layer | Quelle | Precedence |
|---|---|---|
| base | `base/` | Niedrigste — immer aktiv |
| profile | `profiles/<name>/` | Überschreibt base |
| flavor | `flavors/<name>/` | Überschreibt profile (kombinierbar) |
| override | lokal im Ziel-Repo (siehe unten) | Höchste — nie überschrieben |

## Kompositions-Regeln

### `CLAUDE.md` / `AGENTS.md`

Inhalte werden in Layer-Reihenfolge konkateniert (base → profile → flavors).
Das Ergebnis landet im GENERATED-Block. Der CUSTOM-Block der Root-`CLAUDE.md`
bleibt **immer** erhalten und wird nie überschrieben.

### `.claude/settings.json`

Deep JSON Merge. Objekte werden rekursiv gemergt; Arrays werden ersetzt (nicht
konkateniert). Override-Datei `.claude/overrides/settings.override.json` im
Ziel-Repo hat höchste Priorität.

### `.claude/commands/`

Skill-Dateien werden additiv kopiert. Spätere Layer überschreiben frühere
gleichnamige Dateien. Dateien unter `.claude/commands/overrides/` werden nie
angefasst.

### `.claude/hooks/`

Additive Kopie. Die forgecrate-Hooks (`prompt-submit.sh`, `pre-tool.sh`) werden
bei jedem Update überschrieben — eigene Hooks unter anderem Dateinamen ablegen
(siehe [docs/hooks.md](hooks.md)).

### `extensions.yaml` (Plugins + MCP)

Wird über alle Layer gemergt (base + profile + aktive flavors). Aus dem
zusammengeführten Ergebnis generiert forgecrate die Datei `.mcp.json` im
Ziel-Repo. Editiere `extensions.yaml`, nicht `.mcp.json` direkt.

## Override-Pfade im Ziel-Repo

| Pfad | Wirkt auf | Wird überschrieben? |
|---|---|---|
| `<!-- CUSTOM:BEGIN -->...<!-- CUSTOM:END -->` in `CLAUDE.md` | Markdown-Inhalt | Nein |
| `.claude/overrides/settings.override.json` | `.claude/settings.json` | Nein |
| `.claude/commands/overrides/*.md` | Slash-Commands | Nein |
| `.claude/hooks/<eigener-name>.sh` | Hooks | Nein (nur die forgecrate-eigenen Hooks werden überschrieben) |

## Beispiel

```
base/CLAUDE.md             → "# Claude-Konfiguration\n..."
profiles/backend/CLAUDE.md → "## Backend-Profil\n- API-Design: REST-First\n..."
flavors/tdd/CLAUDE.md      → "## TDD-Flavor\n- Test schreiben → ausführen\n..."

Ergebnis CLAUDE.md im Ziel-Repo:

  <!-- GENERATED:BEGIN -->
  # Claude-Konfiguration
  ...

  ## Backend-Profil
  - API-Design: REST-First
  ...

  ## TDD-Flavor
  - Test schreiben → ausführen
  ...
  <!-- GENERATED:END -->

  <!-- CUSTOM:BEGIN -->
  # Eigene Team-Konventionen
  - Wir nutzen Postgres 15
  - Naming: snake_case für DB-Spalten
  <!-- CUSTOM:END -->
```

> **Hinweis:** Die Root-`CLAUDE.md` ist laut [ADR](architecture-decisions/CLAUDE-md-ownership.md)
> manuell gepflegt — die GENERATED-Marker sind ein Legacy-Artefakt. Aktiv
> komponiert wird primär `base/CLAUDE.md` plus die Layer-Beiträge.
