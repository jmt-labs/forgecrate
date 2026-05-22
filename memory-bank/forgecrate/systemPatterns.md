# System Patterns

## Architektur-Entscheidungen

- **Layering-System:** `base/` → `profiles/<name>/` → `flavors/<name>/` → lokale Overrides. Spätere Layer überschreiben frühere. So bleibt base immer gültig, Profiles und Flavors können überschreiben.
- **Hash-basierte Konfliktbehandlung:** Beim Update wird ein Hash der deploytes Datei gespeichert. Ein Konflikt entsteht nur wenn lokale Datei UND Upstream sich beide geändert haben — verhindert false positives bei unveränderten Dateien.
- **Single Binary:** Kein Server, kein Daemon. Das CLI-Binary ist die einzige Laufzeitkomponente.
- **`internal/` für Business-Logik:** Cobra-Commands in `cmd/forgecrate/` sind dünn — Business-Logik liegt in `internal/` (deploy, config, extensions, github).
- **Model-IDs zentralisiert:** Alle Claude-Model-IDs ausschließlich in `base/models.yaml` — enforced via `make check-model-ids`.

## Verzeichnisstruktur

| Pfad | Zweck |
|---|---|
| `cmd/forgecrate/` | CLI-Entry-Points (Cobra-Commands: init, update, config, describe, list, hook) |
| `internal/config/` | `.forgecrate.yaml` lesen/schreiben |
| `internal/deploy/` | Profil+Flavor nach Ziel-Repo deployen |
| `internal/compose/` | Layer-System: Markdown, JSON, Skills zusammenführen |
| `internal/extensions/` | Plugin/Skill-Installation via `claude plugin install` |
| `internal/github/` | GitHub-Release-Download via `net/http` + `api.github.com` |
| `base/` | Base-Layer: CLAUDE.md-Template, Hooks, Skills, MCP-Config |
| `profiles/` | Profil-Definitionen: backend, frontend, fullstack, getbetter, github, gitops, minimal |
| `flavors/` | Flavor-Definitionen: tdd, strict-review, no-research |
| `e2e/` | End-to-End-Tests (benötigen `CLAUDE_BIN` oder Fake-Binary) |

## Compose-Pipeline

`internal/compose/` ist der Kern des Layering-Systems:
- `compose.go` — orchestriert die gesamte Pipeline (Markdown, JSON, Skills)
- `markdown.go` — merged Markdown-Dateien aus mehreren Layers mit CUSTOM-Abschnitts-Schutz
- `jsonmerge.go` — Deep-Merge für `settings.json` (base → profile → overrides)
- `skills.go` — kopiert/merged Slash-Commands aus allen Layer-Verzeichnissen

## Wiederkehrende Muster

- **Error wrapping:** `fmt.Errorf("context: %w", err)` — konsistent durch das gesamte Projekt
- **`io.Writer`-Parameter für Output:** Funktionen nehmen `out io.Writer` statt direkt `os.Stdout` — ermöglicht einfaches Testen
- **Testfiles neben Production Code:** `*_test.go` liegen direkt neben den getesteten Dateien
- **Integrationstests bevorzugt:** Tests schreiben in temporäre Verzeichnisse und prüfen Datei-Output, statt Mocks zu verwenden
- **YAML als Konfigurationsformat:** Sowohl für `.forgecrate.yaml` als auch für Profil/Flavor-Definitionen

## Externe Abhängigkeiten

- GitHub API (`api.github.com`) für Release-Download via `internal/github/`
- `claude` CLI-Binary (Laufzeit-Dependency für Extensions-Install via `exec.Command`)

## Anti-Patterns

- Keine Raw-Queries oder unsichere Shellkonstrukte (kein `exec.Command` mit unkontrollierten Inputs)
- Keine Model-IDs hardcoded außerhalb von `base/models.yaml`
