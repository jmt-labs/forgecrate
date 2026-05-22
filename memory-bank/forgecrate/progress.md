# Progress

## Fertig

- Layering-System (base → profile → flavor) implementiert und stabil
- `forgecrate init`, `update`, `config`, `describe`, `list`, `hook` Commands vorhanden
- Hash-basierte Konfliktbehandlung beim Update
- memory-bank als MCP-Server in base layer integriert (PR #81, #82, #83, #84)
- Session-Start auf memory-bank umgestellt (liest via MCP statt direkter Datei-Tools)
- memory-bank/forgecrate/ Struktur angelegt (MCP-kompatibel, projectName=forgecrate)
- Flache memory-bank/*.md Dateien entfernt

## In Arbeit

- PR: memory-bank-Migration (flat → forgecrate/) + PR-Workflow-Regel in base/CLAUDE.md

## Nächste Schritte

- PR erstellen für aktuelle Änderungen
- `forgecrate update` in abhängigen Repos anstoßen, um neue Workflow-Regel zu verteilen
