# Design: Lodestone — AI-Trend-Intelligence für forgecrate

> Status: RFC / Spec • Autor: Architektur-Evolution • Branch: `claude/ai-trend-intelligence-evolution-un74G`
> Folge-Plan: `docs/superpowers/plans/2026-05-19-lodestone-mvp.md` (Phase 1)

## Ziel

Externe AI-Ökosystem-Signale (AI-News, GitHub-Trending, MCP-Registry, Framework-Releases, ArXiv) in **kontextualisierte Implementierungspläne** für das konkrete Repo überführen — opt-in, ohne `forgecrate update`-Verhalten für bestehende Nutzer zu verändern.

## Hintergrund

forgecrate ist heute ein deklarativer Config-Generator: ein Go-CLI (Cobra), das via 4-Layer-Composition (`base` → `profile` → `flavors` → `overrides`) `.claude/`, `CLAUDE.md`, `AGENTS.md` und `.mcp.json` in beliebigen Repos materialisiert. Bestehende AI-Hilfen (`forgecrate-advisor`, `forgecrate-repo-health`, Flavor `getbetter`) sind punktuelle Repo-Checks.

Lücken:
- Keine **externe Signal-Ingestion** (AI-News, Trending, MCP, Frameworks).
- Keine **Recommendation-Engine**, die Trends gegen den konkreten Stack scort.
- Kein **Planning-Output** im forgecrate-eigenen Spec/Plan-Format.
- Keine **stateful, zeitabhängige Daten** im Repo.

## Naming

**Produktname: `lodestone`**. Ein Lodestone ist ein natürlich magnetischer Stein, historisch Vorgänger des Kompasses. Das Bild passt: das System richtet das Repo nach externen Signalen aus, ohne den Kurs vorzuschreiben. Neutral, nicht-hype, harmoniert phonetisch mit `forgecrate` (handwerklich, Stein/Eisen).

Alternativen erwogen: `cairn` (Wegmarke), `compass` (zu generisch), `signalforge` (zu korporativ), `tide` (zu metaphorisch).

CLI-Subkommando: `forgecrate lodestone <verb>` mit `ingest`, `score`, `plan`, `signals`, `recommend`, `apply`.

## Integrations-Modell — Hybrid

Drei Schichten, jede an dem Ort, an dem sie semantisch hingehört:

1. **CLI-Subcommand `forgecrate lodestone`** — deterministische Pipeline (Ingest, Caching, Scoring, Planning-Skelett) als Go-Code. Spricht keine LLMs direkt — produziert strukturierte Daten und ruft `claude --print --skill ...` nur auf, wenn explizit gewünscht.
2. **Neuer Flavor `flavors/lodestone/`** — bringt Skills, MCP-Verweise und `.gitignore`-Snippet mit. Opt-in: Repos ohne den Flavor merken nichts.
3. **Optionaler MCP-Server `lodestone-mcp`** (Phase 3) — separates Go-Binary, das Pipeline-Ergebnisse Claude-Sessions als Tools anbietet.

**Warum kein reiner Subcommand:** Skills brauchen einen Einstiegspunkt für Claude.
**Warum kein reiner MCP-Server:** Ingestion und Scoring sind Batch-Processing, gehören in Go-Code, der MCP-Server ist dünne Lese-Schicht.
**Warum kein reiner Flavor:** Statische Flavor-Konfiguration kann keine HTTP-Calls, kein Caching, keine Pläne generieren.

## Architektur

### Neue Komponenten (additiv, nicht ersetzend)

| Komponente | Zweck | Ort | Phase |
|---|---|---|---|
| Subcommand `lodestone` | Pipeline-Orchestrierung | `cmd/forgecrate/lodestone.go` | 1 |
| Pipeline-Packages | Ingest/Fingerprint/Scoring/Planning | `internal/lodestone/{ingest,fingerprint,scoring,planning,schema,store}/` | 1–2 |
| Flavor `lodestone` | Skills + MCP-Hooks für Zielrepos | `flavors/lodestone/{CLAUDE.md,extensions.yaml,skills/}` | 2 |
| MCP-Server | Live-Query-Layer für Claude | `cmd/lodestone-mcp/main.go` | 3 |
| Goals-Erweiterung | `.forgecrate.yaml` neue optionale Felder | `internal/config/config.go` | 1 |

### Datei-Handoff statt Event-Bus

```
ingest      → .forgecrate/lodestone/signals.jsonl
fingerprint → .forgecrate/lodestone/fingerprint.json
score       → .forgecrate/lodestone/recommendations.jsonl
plan <id>   → docs/superpowers/specs/YYYY-MM-DD-<slug>-design.md
            → docs/superpowers/plans/YYYY-MM-DD-<slug>.md
            → docs/superpowers/plans/YYYY-MM-DD-<slug>-tasks.yaml
```

Jede Stufe ist einzeln re-runnbar und debuggbar. Folgt dem bestehenden forgecrate-Pattern (Files + Hash-Tracking in `internal/config`).

### Lokale Laufzeit-Artefakte (per Flavor in `.gitignore`)

```
.forgecrate/lodestone/
├── cache/<source>-<YYYY-MM-DD>.json   # rohe Antworten, TTL via Datum
├── signals.jsonl                       # normalisiert, dedupliziert
├── signals.idx                         # signal_id → file_offset
├── fingerprint.json                    # letzter Repo-Scan
├── recommendations.jsonl               # gescorte Vorschläge
└── decisions.log                       # COMMIT-BAR: Audit-Trail
```

Default-`.gitignore`-Snippet ignoriert alles außer `decisions.log` (defensiver Datenhaltungs-Default).

## Agent-System

Jeder „Agent" ist entweder **Go-Pipeline-Schritt** (deterministisch, keine LLMs) oder **Claude-Skill** (LLM-getrieben). Modelle gemäß `base/models.yaml`.

| Agent | Typ | Modell | Trigger | Output |
|---|---|---|---|---|
| Trend Scout | Go | — | `lodestone ingest` | `signals.jsonl` |
| GitHub Intelligence | Go | — | innerhalb Scout | reichere Signal-Felder |
| Repo Analyzer | Go | — | `lodestone fingerprint` | `fingerprint.json` |
| Compatibility Scorer | Go + LLM | `mechanical` (haiku-4-5) für Rationale | `lodestone score` | `recommendations.jsonl` |
| Recommendation Skill | Skill | `default` (sonnet-4-6) | `/lodestone-recommend` | Aufbereitete Liste |
| Planning Engine | Go + LLM | `planning` (opus-4-7) | `lodestone plan <id>` | Spec/Plan/Tasks |
| Architecture Reviewer | Skill | `review` (sonnet-4-6) | optional vor Commit | Review-Kommentare |

**Memory-Strategie:** Nur Entscheidungen und Architektur-Begründungen gehen nach `.claude/memory.json` (Entities `lodestone:decision`, `architecture-choice`). Rohe Daten bleiben in `.forgecrate/lodestone/`.

## Trend Discovery — Quellen

| Quelle | Methode | Recency |
|---|---|---|
| GitHub Trending | Search-API mit `created:>...` + `stars:>...` (offiziell, kein Mirror) | täglich |
| HackerNews | Firebase API, Filter auf Story-Type + Schlüsselwörter | täglich |
| ArXiv | Atom-Feed `cs.AI`, `cs.SE` | täglich |
| Anthropic Changelog | HTTPS-Fetch + Diff zur Cache-Version | wöchentlich |
| OpenAI Release Notes | dito | wöchentlich |
| MCP Registry (lobehub, mcp.so) | HTTP, Robots-Policy respektieren | täglich |
| npm/PyPI Trends | offizielle Download-Stats-APIs | wöchentlich |
| `awesome-*` Listen | GitHub-Repos lesen, Markdown-Links | wöchentlich |

**Adapter-Interface** `ingest.Source`:
```go
type Source interface {
    Name() string
    Fetch(ctx context.Context) ([]RawSignal, error)
}
```

**Pipeline:** Fetch → Cache (`<source>-<YYYY-MM-DD>.json`) → Normalize → Dedup (`sha256(source + canonical_url)`) → Filter (Anti-Hype).

**Anti-Hype-Defaults** (konfigurierbar in `.forgecrate.yaml` unter `lodestone:`):
- `min_stars: 50`
- `min_age_days: 30`
- `max_last_commit_age_days: 180`
- `require_license: true`

**Scheduling:** Default manuell (`forgecrate lodestone ingest`). Optional GitHub-Action-Template (wöchentlicher Cron + Branch-PR). Kein Daemon, kein Background-Polling.

## Recommendation/Scoring

### Repo-Fingerprint

Deterministisch in Go. Felder:
- `languages[]` (Manifest-Existenz)
- `frameworks[]` (Top-Level-Deps + Heuristik)
- `deps[]` mit Version
- `loc_per_language`, `test_ratio`, `has_ci`, `ci_provider`
- `mcp_servers[]` (parse `base/extensions.yaml` + `.mcp.json`)
- `goals[]`, `tech_interests[]` (neuer Block in `.forgecrate.yaml`)

Falls `goals:` fehlt: Skill fragt interaktiv beim ersten Lauf, schreibt zurück.

### Scoring-Dimensionen

| Dimension | Skala | Berechnung |
|---|---|---|
| `compatibility` | 0.0–1.0 | Gewichtete Jaccard-Ähnlichkeit `signal.topic_tags` ∩ `fingerprint.{frameworks,languages}` |
| `effort` | XS / S / M / L / XL | Heuristik: 0 neue Deps + 0 neue Files = XS; neuer Service = L |
| `roi` | low / med / high | Mapping aus `goals[]`-Treffern + `compatibility` |
| `risk` | low / med / high | Sterne + Maintenance + Alter + License + (Phase 2) CVE |

**Explanation-Layer:** Jede `Recommendation` enthält `rationale` (3 Sätze) und `counter_evidence` (1 Satz). Generiert vom `mechanical`-Modell — kurz, schematisch.

**Compatibility-Threshold:** ≥ 0.4 angezeigt, ≥ 0.7 als „Empfehlung" markiert.

## Implementation-Planning-Engine

Pipeline `Recommendation → Epic → Story → Task → Subtask → Agent-Job`:

1. `forgecrate lodestone plan <rec-id>`.
2. Go-Code lädt Fingerprint + Recommendation.
3. Ruft `claude --print --skill lodestone-plan --input <json>` auf.
4. Skill nutzt `superpowers:writing-plans` und produziert Spec + Plan + WorkPackages-YAML.
5. Go-Code validiert gegen Schema in `internal/lodestone/schema/`, schreibt nur bei Validität.

**WorkPackage-Schema** (`docs/superpowers/plans/...-tasks.yaml`):
```yaml
- id: WP-001
  type: task
  title: "Add fetch adapter for HackerNews"
  depends_on: []
  files_affected:
    - internal/lodestone/ingest/hackernews.go
    - internal/lodestone/ingest/hackernews_test.go
  expected_artifacts:
    - "HackerNewsSource type implementing ingest.Source"
  executor: developer        # → models.yaml-Mapping
  estimated_minutes: 45
  acceptance_criteria:
    - "Unit-Test deckt erfolgreichen Fetch und Timeout ab"
    - "Rückgabe als []Signal mit gültiger ID"
```

DAG: Topologie-Sort + parallelisierbare Gruppen als Mermaid-Diagramm im Plan-Markdown.

## MCP + Claude Integration

**Phase 1–2:** Keine eigenen MCP-Tools. Skills rufen `forgecrate lodestone …` via Bash, lesen JSON, präsentieren aufbereitet.

**Phase 3:** `lodestone-mcp` Server (Go-Binary) mit Tools:

| Tool | Input | Output |
|---|---|---|
| `lodestone.list_signals` | `{since?, source?, language?, min_score?}` | `Signal[]` |
| `lodestone.query_trends` | `{query, top_k?}` | `Signal[]`, BM25-gerankt |
| `lodestone.score_repo` | `{signal_id?}` | `Recommendation[]` |
| `lodestone.generate_plan` | `{recommendation_id}` | Pfade zu Spec/Plan/Tasks |
| `lodestone.record_decision` | `{recommendation_id, decision, note?}` | OK + Memory-Eintrag |

**Skills (Phase 2):**
- `/lodestone-scout` — ruft `lodestone ingest`, zeigt Top-10.
- `/lodestone-recommend` — `fingerprint` + `score`, fragt Goals wenn unbekannt.
- `/lodestone-plan <rec-id>` — ruft `lodestone plan`, fragt nach Branch-Anlage.
- `/lodestone-review-trends` — gruppiert Recommendations, persistiert Decisions.

## Phasen-Roadmap

| Phase | Goal | Deliverables | Dauer |
|---|---|---|---|
| 1 (MVP) | Pipeline-Foundation | 2 Sources, Fingerprint Go+Node, Scoring, JSONL-Store | ~2 Wo |
| 2 (Planning) | Plan-Generation + Flavor | 4 weitere Sources, 4 Skills, Spec/Plan-Writer | ~4 Wo |
| 3 (MCP) | Continuous + Live-Queries | `lodestone-mcp` Binary, GitHub-Action-Template | ~6 Wo |
| 4 (Autonomous) | Auto-PR + Self-Improvement | Auto-PR-Engine, Success-Tracker, Re-Scoring | TBD |

### Phase 4 — Autonomous (Sicherheitsschranken)

Auto-PR-Engine (`internal/lodestone/autopr/`) öffnet **nur** PRs mit:
- `risk: low`
- `effort: XS` (≤ 30 min)
- `compatibility: ≥ 0.85`
- Max 1 Auto-PR pro Tag pro Repo
- Nie direkt auf `main` (forgecrate-Konvention, `pre-tool.sh` bleibt aktiv)

Rollback: `forgecrate lodestone undo <pr-number>`.

Self-Improvement (`internal/lodestone/feedback/`): Quartals-Re-Scoring auf Basis von `decisions.log`-Outcomes (`merged`/`reverted`/`stale`) — als Vorschlag-PR auf `.forgecrate.yaml`, nie automatisch.

Cross-Repo-Learning (Epic E5): Erfordert eigenen Brainstorming-Spec (Privacy, Trust).

## Tradeoffs

- **Build vs. Buy:** Renovate/Dependabot decken Deps, nicht das AI/MCP-Ökosystem. Off-the-shelf-Ersatz fehlt → Build.
- **Local vs. External:** Lokal. Keine Telemetrie. LLM-Aufrufe nur opt-in (`recommend`, `plan`); Ingestion+Scoring LLM-frei.
- **Pull vs. Push:** Pull-Default. Push via GitHub-Action-Template opt-in.
- **Storage:** Files+JSONL für MVP. SQLite/bbolt erst bei nachgewiesenem Skalierungs-Bedarf — passt zum minimal-Dependency-Charakter von forgecrate.

## Nicht in Scope (für Phase 1)

- LLM-getriebenes Planning (Phase 2)
- MCP-Server (Phase 3)
- Auto-PR-Engine (Phase 4)
- Cross-Repo-Sharing (Epic E5, eigener Spec)
- Marketplace/Registry (Phase 4+)

## Testbarkeit

- Go unit-Tests pro Package, Coverage ≥ 70%.
- E2E: `e2e/lodestone_test.sh` — `forgecrate lodestone ingest && score && signals --top 5`.
- Fingerprint-Goldens unter `internal/lodestone/fingerprint/testdata/`.
- Schema-Validation: alle JSON-Outputs gegen Schema validiert.
- Backward-Compat: `forgecrate update` ohne `lodestone`-Flavor erzeugt identischen Output vs. Pre-Branch.

## Offene Punkte (nicht-blockierend, vor jeweiliger Phase klären)

1. **Phase 3 — `lodestone-mcp` Distribution:** gebündelt oder separates Paket. Vorschlag: separat.
2. **Phase 4 — Cross-Repo-Sharing:** Privacy-Model + Trust-Setup vor Implementation.
3. **License-Verifikation:** Lodestone-Lizenz = forgecrate-Lizenz (vor Phase 2 bestätigen).

## Affected Files (Phase 1)

- `cmd/forgecrate/main.go` — Subcommand registriert
- `cmd/forgecrate/lodestone.go` — **NEU**
- `internal/config/config.go` — `goals`/`lodestone:`-Felder optional
- `internal/lodestone/**` — **NEU**, ~20 Dateien
- `e2e/lodestone_test.sh` — **NEU**
- `README.md` — Subcommand-Eintrag
- `CHANGELOG.md` — `Added: lodestone subcommand (MVP)`
