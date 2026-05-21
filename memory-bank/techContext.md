# Tech Context

## Stack

- **Sprache:** Go 1.24 | **Einstiegspunkt:** `cmd/forgecrate/main.go`
- **CLI-Framework:** [Cobra](https://github.com/spf13/cobra) v1.10.2
- **TUI/Forms:** [Charmbracelet Huh](https://github.com/charmbracelet/huh) v1.0.0 (interaktive Prompts)
- **YAML-Parsing:** `gopkg.in/yaml.v3`
- **Hashing:** `github.com/mitchellh/hashstructure/v2` (für Konflikt-Erkennung)
- **Konfigurationsformat:** `.forgecrate.yaml` im Ziel-Repo

forgecrate installiert eine reproduzierbare Claude Code-Konfiguration in beliebige Git-Repositories. Ein einziges Binary lädt Profile, Flavors, Hooks, Skills und MCP-Server-Definitionen von GitHub und schreibt sie per Layering-System ins Ziel-Repo.

## Tools & Infrastruktur

- **Build:** `go build -o forgecrate ./cmd/forgecrate/`
- **Tests:** `go test ./internal/... ./cmd/...` (Unit + Integration)
- **E2E-Tests:** `make test-e2e` — benötigen `CLAUDE_BIN` oder Fake-Binary
- **Lint/Vet:** `go vet ./...`
- **Release:** `goreleaser release --clean`
- **CI:** GitHub Actions (`.github/workflows/ci.yml`)
- **Zusätzliche Make-Targets:** `check-model-ids`, `check-readme-coverage`

## Constraints

- Keine CGO — reines Go, statisch linkbar
- `claude` CLI muss im System vorhanden sein für Skill/Plugin-Installation (`forgecrate init`)
- GitHub-Releases sind die einzige Quelle für Profile/Flavors/Base-Layer (kein Registry-Server)
- Model-IDs dürfen nur in `base/models.yaml` stehen (enforced via `make check-model-ids`)
