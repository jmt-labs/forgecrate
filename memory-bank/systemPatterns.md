# System Patterns

## Architektur-Entscheidungen

- **Layering-System:** `base/` → `profiles/<name>/` → `flavors/<name>/` → lokale Overrides. Spätere Layer überschreiben frühere. So bleibt base immer gültig, Profiles und Flavors können überschreiben.
- **Hash-basierte Konfliktbehandlung:** Beim Update wird ein Hash der deploytes Datei gespeichert. Ein Konflikt entsteht nur wenn lokale Datei UND Upstream sich beide geändert haben — verhindert false positives bei unveränderten Dateien.
- **Single Binary:** Kein Server, kein Daemon. Das CLI-Binary ist die einzige Laufzeitkomponente.
- **`internal/` für Business-Logik:** Cobra-Commands in `cmd/forgecrate/` sind dünn — Business-Logik liegt in `internal/` (deploy, config, extensions, github).
- **Model-IDs zentralisiert:** Alle Claude-Model-IDs ausschließlich in `base/models.yaml` — enforced via CI.

## Wiederkehrende Muster

- **Error wrapping:** `fmt.Errorf("context: %w", err)` — konsistent durch das gesamte Projekt
- **`io.Writer`-Parameter für Output:** Funktionen nehmen `out io.Writer` statt direkt `os.Stdout` — ermöglicht einfaches Testen
- **Testfiles neben Production Code:** `*_test.go` liegen direkt neben den getesteten Dateien
- **Integrationstests bevorzugt:** Tests schreiben in temporäre Verzeichnisse und prüfen Datei-Output, statt Mocks zu verwenden
- **YAML als Konfigurationsformat:** Sowohl für `.forgecrate.yaml` als auch für Profil/Flavor-Definitionen

## Anti-Patterns

- Keine Raw-Queries oder unsichere Shellkonstrukte (kein `exec.Command` mit unkontrollierten Inputs)
- Keine Model-IDs hardcoded außerhalb von `base/models.yaml`
