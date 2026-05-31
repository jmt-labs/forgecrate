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

## In Arbeit

- Validierungs-Feature: implementiert + getestet (make test grün, golangci-lint 0 issues),
  Commit/Push auf Branch claude/project-improvement-planning-JVwWc

## Nächste Schritte

- Ggf. PR erstellen (auf Wunsch des Nutzers)
- `forgecrate update` in abhängigen Repos anstoßen, um neue Regeln zu verteilen
