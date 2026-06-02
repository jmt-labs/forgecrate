# Progress

## Fertig

- Layering-System (base → profile → flavor) implementiert und stabil
- `forgecrate init`, `update`, `config`, `describe`, `list`, `hook` Commands vorhanden
- Hash-basierte Konfliktbehandlung beim Update
- memory-bank als MCP-Server in base layer integriert (PR #81, #82, #83, #84)
- Session-Start auf memory-bank umgestellt (liest via MCP statt direkter Datei-Tools)
- memory-bank/forgecrate/ Struktur angelegt (MCP-kompatibel, projectName=forgecrate)
- Flache memory-bank/*.md Dateien entfernt
- Validierung von Profil-/Flavor-Namen mit Levenshtein-Vorschlägen
  (internal/deploy/validate.go + validate_test.go); harter Abbruch vor Schreibvorgang
- codegraph-Flavor vollständig implementiert (#87, #88)
  - Flavor-Hook-Support in deploy.go, Flavor-Settings-Merge in compose.go
  - Gitignore-Appending in deploy.go, `flavors/codegraph/` mit allen Dateien
- PR #114 gemergt: Claude Plugins-Abschnitt in base/CLAUDE.md

## In Arbeit

- Workflow-Overhaul (Spec fertig: `docs/superpowers/specs/2026-06-02-workflow-overhaul-design.md`)

## Nächste Schritte

- Implementierungsplan für Workflow-Overhaul erstellen
- `forgecrate update` in abhängigen Repos anstoßen
