# Lodestone Phase 1 (MVP) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Spec:** `docs/superpowers/specs/2026-05-19-lodestone-design.md`

**Goal:** Deterministischer Subcommand `forgecrate lodestone` mit Ingest + Fingerprint + Scoring + JSONL-Store + 2 Sources (GitHub Trending, HackerNews) + Fingerprint für Go+Node. Keine LLM-Aufrufe in Phase 1.

**Architecture:** Neue Go-Packages unter `internal/lodestone/` (`schema`, `store`, `ingest`, `fingerprint`, `scoring`) + Subcommand-Wrapper in `cmd/forgecrate/lodestone.go`. Datei-Handoff: JSONL/JSON unter `.forgecrate/lodestone/`. Default-`.gitignore` exkludiert alles außer `decisions.log`.

**Tech Stack:** Go 1.24, stdlib (`net/http`, `encoding/json`, `crypto/sha256`), cobra. Keine neuen externen Dependencies.

**Exit-Kriterium Phase 1:** `forgecrate lodestone score --json` produziert deterministisch sortierte Liste. `bash e2e/lodestone_test.sh` läuft grün. `forgecrate update` ohne lodestone-Flavor unverändert.

---

## Task 1 — Schemas anlegen

**Files:**
- Create: `internal/lodestone/schema/signal.go`
- Create: `internal/lodestone/schema/fingerprint.go`
- Create: `internal/lodestone/schema/recommendation.go`
- Create: `internal/lodestone/schema/workpackage.go`
- Create: `internal/lodestone/schema/schema_test.go`

- [ ] **Step 1: `signal.go` anlegen**

```go
package schema

import "time"

const SignalSchemaVersion = 1

type Signal struct {
    SchemaVersion    int       `json:"schema_version"`
    ID               string    `json:"id"`               // sha256(source + canonical_url)
    Source           string    `json:"source"`           // "github_trending", "hackernews", ...
    URL              string    `json:"url"`
    Title            string    `json:"title"`
    Summary          string    `json:"summary"`
    CapturedAt       time.Time `json:"captured_at"`
    Language         string    `json:"language,omitempty"`
    Stars            int       `json:"stars,omitempty"`
    TopicTags        []string  `json:"topic_tags,omitempty"`
    MaintenanceScore float64   `json:"maintenance_score,omitempty"`
    License          string    `json:"license,omitempty"`
    LastCommit       time.Time `json:"last_commit,omitempty"`
}
```

- [ ] **Step 2: `fingerprint.go` anlegen**

```go
package schema

const FingerprintSchemaVersion = 1

type Fingerprint struct {
    SchemaVersion    int               `json:"schema_version"`
    Languages        []string          `json:"languages"`
    Frameworks       []string          `json:"frameworks"`
    Deps             map[string]string `json:"deps"`             // name → version
    LOCPerLanguage   map[string]int    `json:"loc_per_language"`
    TestRatio        float64           `json:"test_ratio"`
    HasCI            bool              `json:"has_ci"`
    CIProvider       string            `json:"ci_provider,omitempty"`
    MCPServers       []string          `json:"mcp_servers,omitempty"`
    Goals            []string          `json:"goals,omitempty"`
    TechInterests    []string          `json:"tech_interests,omitempty"`
}
```

- [ ] **Step 3: `recommendation.go` anlegen**

```go
package schema

const RecommendationSchemaVersion = 1

type Recommendation struct {
    SchemaVersion   int      `json:"schema_version"`
    ID              string   `json:"id"`
    SignalID        string   `json:"signal_id"`
    Compatibility   float64  `json:"compatibility"`     // 0.0–1.0
    Effort          string   `json:"effort"`            // "XS"|"S"|"M"|"L"|"XL"
    ROI             string   `json:"roi"`               // "low"|"med"|"high"
    Risk            string   `json:"risk"`              // "low"|"med"|"high"
    Rationale       string   `json:"rationale,omitempty"`
    CounterEvidence string   `json:"counter_evidence,omitempty"`
    SuggestedNext   string   `json:"suggested_next_step,omitempty"`
}
```

- [ ] **Step 4: `workpackage.go` anlegen** (Phase 2 nutzt das, MVP definiert nur den Typ)

```go
package schema

const WorkPackageSchemaVersion = 1

type WorkPackage struct {
    SchemaVersion       int      `json:"schema_version"`
    ID                  string   `json:"id"`
    Type                string   `json:"type"`           // "epic"|"story"|"task"|"subtask"
    Title               string   `json:"title"`
    DependsOn           []string `json:"depends_on,omitempty"`
    FilesAffected       []string `json:"files_affected,omitempty"`
    ExpectedArtifacts   []string `json:"expected_artifacts,omitempty"`
    Executor            string   `json:"executor,omitempty"` // → base/models.yaml-Rolle
    EstimatedMinutes    int      `json:"estimated_minutes,omitempty"`
    AcceptanceCriteria  []string `json:"acceptance_criteria,omitempty"`
}
```

- [ ] **Step 5: `schema_test.go` mit JSON-Roundtrip pro Typ**

```go
package schema_test

import (
    "encoding/json"
    "testing"
    "time"
    "github.com/jmt-labs/forgecrate/internal/lodestone/schema"
)

func TestSignalRoundtrip(t *testing.T) {
    s := schema.Signal{
        SchemaVersion: schema.SignalSchemaVersion,
        ID:            "sha256:abc",
        Source:        "github_trending",
        URL:           "https://example.com",
        Title:         "Demo",
        CapturedAt:    time.Unix(0, 0).UTC(),
    }
    b, err := json.Marshal(s)
    if err != nil { t.Fatal(err) }
    var back schema.Signal
    if err := json.Unmarshal(b, &back); err != nil { t.Fatal(err) }
    if back.ID != s.ID { t.Fatalf("ID mismatch: %q vs %q", back.ID, s.ID) }
}
```
(analog für Fingerprint, Recommendation, WorkPackage)

- [ ] **Step 6: Tests ausführen — müssen grün sein**

```bash
go test ./internal/lodestone/schema/...
```

- [ ] **Step 7: Commit**

```bash
git add internal/lodestone/schema/
git commit -m "feat(lodestone): schemas for Signal, Fingerprint, Recommendation, WorkPackage"
```

---

## Task 2 — Store-Interface + FileStore

**Files:**
- Create: `internal/lodestone/store/store.go`
- Create: `internal/lodestone/store/filestore.go`
- Create: `internal/lodestone/store/filestore_test.go`

- [ ] **Step 1: Interface in `store.go`**

```go
package store

import "github.com/jmt-labs/forgecrate/internal/lodestone/schema"

type SignalStore interface {
    Append(s schema.Signal) (added bool, err error)
    ListSince(unixSec int64) ([]schema.Signal, error)
    Has(id string) (bool, error)
}

type FingerprintStore interface {
    Write(fp schema.Fingerprint) error
    Read() (schema.Fingerprint, error)
}

type RecommendationStore interface {
    Replace(rs []schema.Recommendation) error
    List() ([]schema.Recommendation, error)
}
```

- [ ] **Step 2: `filestore.go` implementiert beide Interfaces über Files unter `.forgecrate/lodestone/`**
  - Signals: JSONL append, In-Memory-Map als Index (rebuild on open)
  - Fingerprint: einzelne JSON-Datei
  - Recommendations: JSONL truncate+rewrite (`Replace`)
- [ ] **Step 3: `filestore_test.go`: tmpdir-basiert, Roundtrip Append→Has→ListSince, Replace→List, Write→Read**
- [ ] **Step 4: Tests grün**

```bash
go test ./internal/lodestone/store/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/lodestone/store/
git commit -m "feat(lodestone): file-based store with JSONL signals + JSON fingerprint"
```

---

## Task 3 — Ingest-Interface + GitHub-Trending-Adapter

**Files:**
- Create: `internal/lodestone/ingest/source.go`
- Create: `internal/lodestone/ingest/github_trending.go`
- Create: `internal/lodestone/ingest/github_trending_test.go`

- [ ] **Step 1: `source.go`**

```go
package ingest

import (
    "context"
    "github.com/jmt-labs/forgecrate/internal/lodestone/schema"
)

type Source interface {
    Name() string
    Fetch(ctx context.Context) ([]schema.Signal, error)
}
```

- [ ] **Step 2: `github_trending.go`** — nutzt offizielle Search-API `https://api.github.com/search/repositories?q=created:>{since}+stars:>{min}&sort=stars`.
  - Authentifiziert via `$GITHUB_TOKEN` falls gesetzt.
  - Normalisiert auf `schema.Signal` mit deterministischer ID `sha256("github_trending|" + html_url)`.
  - Cache-Datei `.forgecrate/lodestone/cache/github_trending-<YYYY-MM-DD>.json` mit roher API-Antwort.
  - Timeout 15 s, ein Retry mit exponentiellem Backoff.
- [ ] **Step 3: `github_trending_test.go`** mit `httptest.Server`:
  - Erfolgsfall mit Fixture-JSON.
  - Timeout-Fall.
  - Empty-Response.
- [ ] **Step 4: Tests grün**

```bash
go test ./internal/lodestone/ingest/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/lodestone/ingest/
git commit -m "feat(lodestone): ingest interface + github trending source"
```

---

## Task 4 — HackerNews-Adapter

**Files:**
- Create: `internal/lodestone/ingest/hackernews.go`
- Create: `internal/lodestone/ingest/hackernews_test.go`

- [ ] **Step 1: `hackernews.go`** — Firebase API `https://hacker-news.firebaseio.com/v0/topstories.json` + `/item/<id>.json`.
  - Filter auf Story-Type + Schlüsselwörter aus optionalem Config-Param `keywords` (default: `["ai", "llm", "mcp", "claude", "agent"]`).
  - Limit 50 Items pro Lauf.
  - Cache-Datei wie oben.
- [ ] **Step 2: `hackernews_test.go`** mit `httptest.Server`.
- [ ] **Step 3: Tests grün, Commit**

```bash
git add internal/lodestone/ingest/hackernews.go internal/lodestone/ingest/hackernews_test.go
git commit -m "feat(lodestone): hackernews source"
```

---

## Task 5 — Fingerprint (Go + Node)

**Files:**
- Create: `internal/lodestone/fingerprint/fingerprint.go`
- Create: `internal/lodestone/fingerprint/golang.go`
- Create: `internal/lodestone/fingerprint/node.go`
- Create: `internal/lodestone/fingerprint/fingerprint_test.go`
- Create: `internal/lodestone/fingerprint/testdata/go_minimal/go.mod`
- Create: `internal/lodestone/fingerprint/testdata/node_react/package.json`

- [ ] **Step 1: `fingerprint.go`** — Walker, der pro Sprache den passenden Detektor aufruft, aggregiert in `schema.Fingerprint`. LOC-Counting per File-Walk (skip `vendor/`, `node_modules/`, `.git/`). Test-Ratio = `_test.go` + `*.test.{js,ts}` + `*.spec.{js,ts}` LOC / Non-Test-LOC.
- [ ] **Step 2: `golang.go`** parst `go.mod` (stdlib `golang.org/x/mod/modfile` ist erlaubt — Teil der erweiterten stdlib, aber wenn extern unerwünscht: regex-basiert auf `require` blocks).
  - **Entscheidung:** regex-basiert, keine neue Dependency.
- [ ] **Step 3: `node.go`** parst `package.json` (stdlib `encoding/json`).
  - Framework-Heuristik: `react` → "react", `vue` → "vue", `next` → "nextjs", `@anthropic-ai/sdk` → "anthropic-sdk".
- [ ] **Step 4: Goldens unter `testdata/`**:
  - `go_minimal/go.mod` mit cobra-Dependency.
  - `node_react/package.json` mit react+next.
- [ ] **Step 5: `fingerprint_test.go`** — Run gegen Goldens, prüft `languages`, `frameworks`, `deps`.
- [ ] **Step 6: Tests grün, Commit**

```bash
git add internal/lodestone/fingerprint/
git commit -m "feat(lodestone): fingerprint for Go + Node (regex go.mod parse, package.json)"
```

---

## Task 6 — Scoring

**Files:**
- Create: `internal/lodestone/scoring/compatibility.go`
- Create: `internal/lodestone/scoring/effort.go`
- Create: `internal/lodestone/scoring/risk.go`
- Create: `internal/lodestone/scoring/scoring.go`
- Create: `internal/lodestone/scoring/scoring_test.go`

- [ ] **Step 1: `compatibility.go`** — gewichtete Jaccard zwischen `signal.topic_tags` ∪ `signal.language` und `fingerprint.frameworks` ∪ `fingerprint.languages`. Gewichtung: Sprach-Match 1.5×, Framework-Match 1.0×.
- [ ] **Step 2: `effort.go`** — Default `M`. Wenn `len(signal.topic_tags ∩ fingerprint.frameworks) > 0` und Signal-Stars < 100 → `S`. Wenn 0 Match → `XL` (separater Service). Konservative Heuristik, Doc dokumentiert die Tabelle.
- [ ] **Step 3: `risk.go`** — `low` wenn Stars ≥ 500 ∧ LastCommit < 90 Tage ∧ License non-empty; `high` wenn eine Bedingung fehlschlägt; sonst `med`.
- [ ] **Step 4: `scoring.go::Score(signals, fp) []Recommendation`** — orchestriert die drei Dimensionen, generiert Rec-ID als `sha256(signal_id + fp.canonical())`. Sortierung deterministisch: `compatibility DESC, signal.stars DESC, signal.id ASC`.
- [ ] **Step 5: `scoring_test.go`** — 5 Fixtures (perfect match, kein Match, mittlerer Match, low-star, no-license), prüft Bestimmtheit (2 Läufe → identische Reihenfolge).
- [ ] **Step 6: Tests grün, Commit**

```bash
git add internal/lodestone/scoring/
git commit -m "feat(lodestone): deterministic scoring (compatibility, effort, risk)"
```

---

## Task 7 — Config-Erweiterung (Goals + Lodestone-Block)

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: Schema-Erweiterung in `config.go`** (rückwärts-kompatibel, alle neuen Felder `omitempty`)

```go
type Config struct {
    // ... existing fields
    Goals          []string         `yaml:"goals,omitempty"`
    TechInterests  []string         `yaml:"tech_interests,omitempty"`
    Lodestone      *LodestoneConfig `yaml:"lodestone,omitempty"`
}

type LodestoneConfig struct {
    MinStars              int  `yaml:"min_stars,omitempty"`
    MinAgeDays            int  `yaml:"min_age_days,omitempty"`
    MaxLastCommitAgeDays  int  `yaml:"max_last_commit_age_days,omitempty"`
    RequireLicense        bool `yaml:"require_license,omitempty"`
}

func DefaultLodestone() LodestoneConfig {
    return LodestoneConfig{
        MinStars: 50, MinAgeDays: 30,
        MaxLastCommitAgeDays: 180, RequireLicense: true,
    }
}
```

- [ ] **Step 2: Test, dass alte `.forgecrate.yaml` ohne neue Felder weiter parst** (Goldens unter `internal/config/testdata/`).
- [ ] **Step 3: Tests grün, Commit**

```bash
git add internal/config/
git commit -m "feat(config): optional goals, tech_interests, lodestone block (backward-compatible)"
```

---

## Task 8 — Subcommand `forgecrate lodestone`

**Files:**
- Modify: `cmd/forgecrate/main.go`
- Create: `cmd/forgecrate/lodestone.go`

- [ ] **Step 1: `lodestone.go` mit Cobra-Subcommand-Baum**:
  - `lodestone ingest [--source ...]`
  - `lodestone fingerprint`
  - `lodestone score`
  - `lodestone signals [--since 7d] [--source ...] [--top N] [--json]`
- [ ] **Step 2: `main.go` registriert den neuen Subcommand**
- [ ] **Step 3: Manueller Smoke-Test im Repo**:

```bash
go run ./cmd/forgecrate lodestone fingerprint
cat .forgecrate/lodestone/fingerprint.json | head -20
```

- [ ] **Step 4: Commit**

```bash
git add cmd/forgecrate/lodestone.go cmd/forgecrate/main.go
git commit -m "feat(cmd): forgecrate lodestone subcommand (ingest, fingerprint, score, signals)"
```

---

## Task 9 — E2E-Smoke-Test

**Files:**
- Create: `e2e/lodestone_test.sh`
- Modify: `Makefile` (Target ergänzen)

- [ ] **Step 1: `lodestone_test.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail
TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT
cd "$TMP"
git init -q
# Minimal go.mod, damit Fingerprint etwas findet
go mod init example.com/test >/dev/null
"$FORGECRATE_BIN" lodestone fingerprint
test -f .forgecrate/lodestone/fingerprint.json
"$FORGECRATE_BIN" lodestone ingest --source github_trending --mock || true
"$FORGECRATE_BIN" lodestone score
test -f .forgecrate/lodestone/recommendations.jsonl
echo OK
```

  Anmerkung: `--mock` Flag respektiert eine Env-Var `LODESTONE_MOCK_FIXTURES=<dir>`, damit der Test offline läuft.
- [ ] **Step 2: Makefile-Target**

```makefile
test-e2e-lodestone: build
	FORGECRATE_BIN=$$PWD/bin/forgecrate bash e2e/lodestone_test.sh
```

- [ ] **Step 3: Lauf grün**

```bash
make test-e2e-lodestone
```

- [ ] **Step 4: Commit**

```bash
git add e2e/lodestone_test.sh Makefile
git commit -m "test(lodestone): e2e smoke test (fingerprint → ingest → score)"
```

---

## Task 10 — README + CHANGELOG + Doc-Coverage

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Create: `docs/lodestone.md`

- [ ] **Step 1: `README.md`** — `forgecrate lodestone` im Command-Table; Verweis auf `docs/lodestone.md`.
- [ ] **Step 2: `docs/lodestone.md`** — User-orientierter Überblick: was ist Lodestone, wie aktivieren (Phase 2 Flavor), Subcommand-Reference für Phase 1.
- [ ] **Step 3: `CHANGELOG.md`** — Eintrag `Added: lodestone subcommand (Phase 1 MVP)`.
- [ ] **Step 4: README-Coverage-CI-Check muss grün bleiben**

```bash
make readme-coverage   # falls Target existiert; sonst go test ./internal/readme/...
```

- [ ] **Step 5: Commit**

```bash
git add README.md CHANGELOG.md docs/lodestone.md
git commit -m "docs(lodestone): user-facing overview + README + CHANGELOG"
```

---

## Task 11 — Backward-Compat-Verifikation

- [ ] **Step 1:** Reproduktions-Workflow gegen Pre-Branch-Tag laufen lassen
  - Auf einem temporären Test-Repo `forgecrate init` + `update` mit demselben Profile (ohne lodestone-Flavor — der existiert in Phase 1 noch nicht).
  - SHA-Hash-Vergleich der erzeugten Dateien zwischen `main` und diesem Branch.
- [ ] **Step 2:** Alle bestehenden Tests grün:

```bash
go test ./...
make test-e2e
golangci-lint run
go vet ./...
govulncheck ./...
```

- [ ] **Step 3:** Falls Diff: dokumentieren, ob erwartet (z.B. neue optionale Config-Felder als no-op) oder ein Bug.

---

## Task 12 — Push & PR (nicht automatisch)

- [ ] **Step 1:** Branch pushen:

```bash
git push -u origin claude/ai-trend-intelligence-evolution-un74G
```

- [ ] **Step 2:** PR **nur auf explizite User-Aufforderung** erstellen (CLAUDE.md-Konvention + Tool-Brief verbietet ungefragten PR). Titel-Vorschlag:

```
feat: lodestone Phase 1 (MVP) — pipeline foundation
```

Body-Vorschlag (Was/Warum/Wie getestet):
- Was: Neuer Subcommand `forgecrate lodestone` mit `ingest|fingerprint|score|signals`; 2 Sources (GitHub Trending, HackerNews); Fingerprint für Go+Node; deterministisches Scoring; JSONL-Store unter `.forgecrate/lodestone/`.
- Warum: Fundament für AI-Trend-Intelligence (Spec: `docs/superpowers/specs/2026-05-19-lodestone-design.md`). Phase 2 (Planning + Flavor) und Phase 3 (MCP) bauen darauf auf.
- Wie getestet: Go unit-Tests pro Package; `e2e/lodestone_test.sh`; Backward-Compat-Check (pre-existing tests grün, kein Output-Diff für Repos ohne lodestone-Flavor).

---

## Abschluss-Checkliste

- [ ] Alle Tasks abgehakt
- [ ] `go test ./...` grün
- [ ] `make test-e2e` grün
- [ ] `golangci-lint run` grün
- [ ] `go vet ./...` grün
- [ ] `govulncheck ./...` grün
- [ ] README-Coverage-Check grün
- [ ] `forgecrate update` Backward-Compat verifiziert
- [ ] Spec referenziert: `docs/superpowers/specs/2026-05-19-lodestone-design.md`
- [ ] Branch gepusht: `claude/ai-trend-intelligence-evolution-un74G`
