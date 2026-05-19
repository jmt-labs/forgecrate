# Entwicklung

## Voraussetzungen

Go 1.22+

## Tests ausführen

```bash
go test ./...        # Unit-Tests
go test ./e2e/...    # E2E-Tests (gegen lokales Repo)
```

## Binary bauen

```bash
go build ./cmd/forgecrate/
```

## Neues Profil hinzufügen

1. `profiles/<name>/CLAUDE.md` anlegen
2. Optional: `profiles/<name>/.claude/settings.json` für Profil-spezifische Settings
3. Optional: `profiles/<name>/.claude/commands/` für Profil-spezifische Skills
