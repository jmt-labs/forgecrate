# Entwicklung

## Voraussetzungen

- **Go 1.24+** (siehe `go.mod`)
- **`make`** für die Standard-Targets
- **`claude` CLI** muss im `PATH` verfügbar sein, damit `forgecrate init` Plugins
  und MCP-Server installieren kann
- **`goreleaser`** nur für `make release` notwendig

## Make-Targets

| Target | Zweck |
|---|---|
| `make test` | Unit + Integration-Tests (`go test ./internal/... ./cmd/...`) |
| `make test-e2e` | E2E-Tests in `e2e/` — nutzt `CLAUDE_BIN` oder ein generiertes Fake-Binary |
| `make quality` | `go vet ./...` + `go build ./...` als Sanity-Check |
| `make build` | Lokales Binary `./forgecrate` bauen |
| `make check-model-ids` | Sicherstellen, dass Claude-Model-IDs nur in `base/models.yaml` stehen |
| `make check-readme-coverage` | Sicherstellen, dass jeder Flavor im `README.md` erwähnt ist |
| `make release` | GoReleaser-Release (benötigt `GITHUB_TOKEN`) |
| `make clean` | Build-Artefakte und Testcache löschen |

## Manuelle Befehle (Alternative zu `make`)

```bash
go build ./cmd/forgecrate              # Binary bauen
go test ./...                          # Alle Tests
go test ./cmd/forgecrate/... -v        # Nur CLI-Tests mit Output
go run ./cmd/forgecrate --help         # CLI lokal ausführen
```

## Neues Profil hinzufügen

1. Verzeichnis anlegen: `profiles/<name>/`
2. Pflicht: `profiles/<name>/CLAUDE.md` mit den Layer-spezifischen Regeln (siehe
   `profiles/backend/CLAUDE.md` als Referenz)
3. Optional: `profiles/<name>/.claude/settings.json` für Profil-spezifische
   Claude-Code-Settings (Modell, Hooks etc.)
4. Optional: `profiles/<name>/.claude/commands/<skill>.md` für Slash-Commands,
   die nur in diesem Profil aktiv sein sollen
5. Optional: `profiles/<name>/extensions.yaml` mit Plugins, MCP-Servern, Skills
   die zusätzlich zum base-layer installiert werden
6. Test: `internal/compose/compose_test.go` enthält Tabellentests, die das
   Layer-System validieren — beim Hinzufügen neuer Komposition-Regeln dort
   erweitern
7. README aktualisieren — `make check-readme-coverage` deckt nur Flavors ab,
   neue Profile gehören manuell in die README-Tabelle

## Neuen Flavor hinzufügen

Analog zu Profilen, nur unter `flavors/<name>/`. Zusätzlich:

- Eintrag in `docs/profiles-flavors.md` ergänzen
- README-Tabelle aktualisieren (CI-Check `make check-readme-coverage`
  schlägt sonst fehl)

## Release-Prozess

Vollständige Schritte siehe Skill `/forgecrate-release`. Kurz:

```bash
git tag v0.0.X
git push --tags
make release        # GoReleaser baut & publisht für alle Plattformen
```

GoReleaser-Konfiguration liegt in `.goreleaser.yaml`. CI veröffentlicht Homebrew-
Tap, apt-Repo und Chocolatey-Package automatisch beim Tag-Push.

## Continuous Integration

Workflows liegen in `.github/workflows/`:

- `ci.yml` — Tests + Lint + `make check-model-ids` bei jedem Push/PR
- Release-Workflow — getriggert durch Tag-Push, ruft GoReleaser auf

## Code-Konventionen

- **`internal/` für Business-Logik** — `cmd/forgecrate/` enthält nur dünne Cobra-Commands
- **`io.Writer`-Parameter** für Output-Funktionen statt direkt `os.Stdout` — testbar
- **Error wrapping** mit `fmt.Errorf("context: %w", err)`
- **Integrationstests** bevorzugt — schreiben in `t.TempDir()`, prüfen Datei-Output statt zu mocken
- **YAML** als Konfigurationsformat (`gopkg.in/yaml.v3`)
- **Kein CGO** — reines Go, statisch linkbar
