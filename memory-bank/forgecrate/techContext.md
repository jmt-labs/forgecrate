# Tech Context

## Stack

- **Sprache:** Go 1.24.2 | **Einstiegspunkt:** `cmd/forgecrate/main.go`
- **CLI-Framework:** [Cobra](https://github.com/spf13/cobra) v1.10.2
- **TUI/Forms:** [Charmbracelet Huh](https://github.com/charmbracelet/huh) v1.0.0 (interaktive Prompts)
- **YAML-Parsing:** `gopkg.in/yaml.v3`
- **Hashing:** `github.com/mitchellh/hashstructure/v2` (für Konflikt-Erkennung beim Update)
- **Konfigurationsformat:** `.forgecrate.yaml` im Ziel-Repo

## Tools & Infrastruktur

- **Build:** `make build` → `go build -o forgecrate ./cmd/forgecrate/`
- **Tests:** `make test` → `go test ./internal/... ./cmd/...`
- **E2E-Tests:** `make test-e2e` — nutzen `CLAUDE_BIN` oder ein Fake-Binary
- **Quality:** `make quality` → `go vet ./... && go build ./...`
- **Lint:** `go vet ./...` (kein golangci-lint konfiguriert)
- **Release:** `make release` → `goreleaser release --clean`
- **CI:** GitHub Actions (`.github/workflows/ci.yml`)
- **Zusätzliche Checks:** `make check-model-ids`, `make check-readme-coverage`

## Constraints

- Keine CGO — reines Go, statisch linkbar
- `claude` CLI muss im System vorhanden sein für Skill/Plugin-Installation (`forgecrate init`)
- GitHub-Releases sind die einzige Quelle für Profile/Flavors/Base-Layer (kein Registry-Server)
- Model-IDs dürfen nur in `base/models.yaml` stehen (enforced via `make check-model-ids`)
- Kein Raw-`exec.Command` mit unkontrollierten Inputs (außer `claude plugin install` mit festen Args)
