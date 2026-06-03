# Profile und Flavors

forgecrate kombiniert genau **ein** Profil mit **null oder mehreren** Flavors.
Profile setzen den technischen Kontext, Flavors fügen Querschnitts-Praktiken
hinzu.

## Profile

| Profil | Fokus | Zusätzliche Plugins | Zusätzliche MCP | Profil-Skills |
|---|---|---|---|---|
| `backend` (Standard) | API, Datenbank, Integrationstests | — | — | `/forgecrate-db-migration` |
| `frontend` | Komponenten, State, Barrierefreiheit | `frontend-design`, `typescript-lsp`, `playwright` | `playwright` | `/accessibility-audit`, `/ui-ux-audit` |
| `fullstack` | Backend + Frontend kombiniert, E2E, Shared Types | — | `playwright` | _(keine eigenen Skills — Mix aus base und Konventionen)_ |

Plugin- und MCP-Listen kommen aus `profiles/<name>/extensions.yaml`. Der base
layer steuert zusätzlich vier Plugins (`superpowers`, `commit-commands`,
`security-guidance`, `claude-md-management`) und sechs MCP-Server (siehe
[CLAUDE.md → MCP-Server](../base/CLAUDE.md)) für **alle** Profile.

## Flavors

| Flavor | Fokus | Skills | Plugins |
|---|---|---|---|
| `tdd` | Test-First, kein Produktionscode ohne Test, Regressionstest vor jedem Bugfix | `/forgecrate-test-coverage` | — |
| `strict-review` | Pflicht-Review vor jedem Commit, keine Direct-Pushes auf `main` | `/forgecrate-pr-checklist` | `pr-review-toolkit`, `code-simplifier` |
| `minimal` | Explizite Entscheidung: keine zusätzliche Konfiguration über base hinaus | — | — |
| `gitops` | ArgoCD-only Deployments, Kyverno/Gatekeeper/`RULES.md`-Enforcement, keine `latest`-Tags | `/forgecrate-gitops-status` | — |
| `getbetter` | Erkenntnisse aus jeder Session in `.claude/GETBETTER.md` festhalten | `/forgecrate-getbetter` | — |
| `github` | Releases via `gh release create`, CI-Checks vor Tag, proaktive Parallelisierung | `/forgecrate-issue-resolver`, `/forgecrate-github-release` | — |
| `force-research` | Erzwungene Web-Recherche vor jeder Code-Änderung; harter Block zusätzlich für schreibende Bash-Befehle | — | — |
| `no-research` | Recherche-Pflicht aus base layer **deaktivieren**, inkl. hartem PreToolUse-Block (für air-gapped / Compliance-Repos) | — | — |
| `codegraph` | Semantischer Code-Wissensgraph als lokaler MCP-Server (colbymchenry/codegraph) | — | — |

Skill- und Plugin-Listen kommen aus den jeweiligen
`flavors/<name>/skills/` bzw. `flavors/<name>/extensions.yaml`.

## Kombinationsregeln

- **Profile schließen sich aus** — exakt eines pro Repo. Wechsel via
  `forgecrate update --profile <neu>` oder `forgecrate config`
- **Flavors sind additiv** — beliebig viele lassen sich kombinieren (z. B.
  `--flavors tdd,strict-review,github`)
- **`no-research` invertiert** — es entfernt die Recherche-Empfehlung aus base und
  deaktiviert die PreToolUse-Warnung vollständig. Sinnvoll nur isoliert oder in stark
  eingeschränkten Umgebungen. Hat Vorrang vor `force-research`, falls beide aktiv sind
- **`force-research` verschärft** — erweitert die Recherche-Warnung zusätzlich auf
  schreibende Bash-Befehle (`sed -i`, `tee`, `dd of=`, Redirects außerhalb `/tmp`)
- **`minimal` fügt nichts hinzu** — es ist ein Signal "keine extras". Kombination
  mit anderen Flavors funktioniert problemlos

## Eigene Profile oder Flavors anlegen

Siehe [docs/development.md](development.md) für Schritt-für-Schritt-Anweisungen.
Kurz:

1. Verzeichnis `profiles/<name>/` oder `flavors/<name>/` anlegen
2. Mindestens `CLAUDE.md` mit den Layer-spezifischen Regeln
3. Optional `extensions.yaml` für eigene Plugins/MCP-Server
4. Optional `skills/<skill-name>/SKILL.md` für eigene Slash-Commands
5. Eintrag in README ergänzen, damit `make check-readme-coverage` grün bleibt
