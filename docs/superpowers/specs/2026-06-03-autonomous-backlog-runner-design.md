# Spec: Autonomous Backlog Runner — autonome Langläufe mit Eskalation

> Status: Design / Backlog · Datum: 2026-06-03 · Epic: Autonomous Backlog Runner

## Kontext & Ziel

ForgeCrate kann heute ein einzelnes GitHub-Issue end-to-end abarbeiten
(`forgecrate-issue-resolver`: Research → Brainstorm → Plan → Worktree → Impl →
Test → PR) und bis zu 5 Issues parallel via Worktree/Background. Es gibt
WSJF-Backlog-Pflege (`forgecrate-roadmap-triage`), Persistenz (`memory`,
`memory-bank`) und passive Token-Sandbox (`context-mode`).

**Was fehlt** für das Ziel „Agenten arbeiten stundenlang autark das Backlog ab":

1. **Kein äußerer Loop** — nach einem Batch stoppt der Agent. Niemand schiebt das
   nächste Issue nach (triage → resolve → verify → next), bis Backlog leer /
   Budget erschöpft / ein Gate getroffen ist.
2. **Kein Checkpoint-Rhythmus** — `handoff` ist manuell; bei Context-Kompaktierung
   mitten im Lauf droht Drift.
3. **Planning nicht messbar tief** — issue-resolver Schritt 3 verlangt Test-Plan +
   Akzeptanzkriterien, prüft aber nicht, ob ein Plan „detailliert genug" ist. Kein
   Gate zwischen Plan und Code.
4. **Token-Budget nicht aktiv gesteuert** — context-mode läuft passiv, kein
   bewusstes Begrenzen von Fan-out/Iterationen.

**Ziel-Outcome:** Ein autonom-mit-Eskalation laufender Backlog-Runner (CLI + Skill
kombiniert), der nur an den bestehenden 5 Eskalations-Gates nachfragt, mit
erzwungenem Planning-Quality-Gate und aktiver Token-/Checkpoint-Cadence.

### Verbindliche Entscheidungen

- Fokus: alle drei — autonome Langläufe, tieferes Planning, Token-Effizienz
- Autonomie: autonom mit Eskalation — nur die bestehenden 5 Gates fragen nach
- Mechanismus: CLI-Command (Orchestrierung) + Skill (Per-Issue-Workflow) kombiniert

## Leitprinzipien (aus Code verifiziert)

1. **Go-Prozess = zustandsloses Hinweis-/State-Tool**, kein Daemon, spawnt keine
   Claude-Sessions. Vertrag wie `cmd/forgecrate/hook.go`: stdin-JSON/Flags lesen →
   Entscheidung → strukturierter stdout → fail-open. Der CLI-Command verwaltet
   Run-State und gibt *Anweisungen* aus, die der Skill liest.
2. **GitHub Issues bleiben System of Record** (roadmap-triage). Go kennt den
   Backlog nicht — es instruiert den Agenten, `roadmap.sh backlog-ranked` zu nutzen.
3. **Skills sind die Orchestrierungsebene**, Go nur State + Entscheidungsbaum.

## Design

### A) CLI-Command `forgecrate autopilot`

Registriert in `cmd/forgecrate/main.go` (`root.AddCommand(newAutopilotCmd())`),
Datei `cmd/forgecrate/autopilot.go` + `autopilot_test.go`. Subcommand-Struktur
analog `newHookCmd()`. State-Logik in neuem Paket `internal/autopilot/state.go`
(Read/Write/Validate, gespiegelt an `internal/config/config.go`).

State in `.forgecrate/autopilot-state.json` (NICHT `.forgecrate.yaml` — deklarative
Deploy-Config vs. ephemerer Laufzeit-State). Muss in `.gitignore`.

```go
type RunState struct {
    RunID        string    // timestamp/ULID
    StartedAt    time.Time
    Status       string    // running|paused|stopped|exhausted|blocked
    Iteration    int       // abgeschlossene Issue-Zyklen
    Budget       Budget    // MaxIterations, MaxConsecFails, ConsecFails, MaxParallel(<=5)
    Filter       Filter    // label, milestone, min-wsjf
    CurrentIssue int
    Gate         *Gate     // gesetzt wenn blocked: Reason(1 der 5 Gates), Issue, Question
    History      []Entry   // {issue, pr, outcome, iteration}
}
```

**Subcommands:** `init` (Flags `--label/--milestone/--min-wsjf/--max-iterations/--max-parallel`),
`next` (Kern: wertet Invarianten aus, gibt Anweisungsblock aus), `checkpoint`
(`--issue/--pr/--outcome shipped|failed|blocked|empty/--gate-reason/--gate-question`),
`status` (`--json`), `stop` (archiviert).

**`next`-Entscheidungsbaum** (reine, testbare Funktion wie `preToolOutput`):
1. `Gate != nil` → `ACTION: ESCALATE` + Frage, Exit
2. `Iteration >= MaxIterations` → `ACTION: STOP` (budget exhausted)
3. `ConsecFails >= MaxConsecFails` → `ACTION: ESCALATE` (Gate 5)
4. sonst → `ACTION: RESOLVE_NEXT` mit Filter-Parametern

**YAGNI:** Kein Token-Counting in Go (Harness/context-mode misst); Budget ist
iterations-/fail-basiert. Kein State-Mirror in GitHub, kein Pluggable-Backend.

### B) Skill `forgecrate-autonomous-runner` (github-Flavor)

Pfad `flavors/github/skills/forgecrate-autonomous-runner/SKILL.md` (+ `references/loop.md`).

**Loop-Invariante:** Nach jeder Iteration ist genau ein Issue terminal (shipped-PR
oder dokumentiert failed/blocked), RunState via `checkpoint` persistiert, ein
`handoff`-Checkpoint existiert, nie mehr als `MaxParallel` Worktrees.

**Ein Durchlauf:** `autopilot next` → bei RESOLVE_NEXT: `roadmap.sh backlog-ranked`
(Top-N) → **Plan-Gate (C)** → `issue-resolver count:N` (worktree+background) → pro
Issue `autopilot checkpoint` → `handoff` → `ctx_stats`-Check → Self-Continuation
via `/loop`+`send_later`.

**Abbruch (terminal):** Backlog leer, Budget erschöpft, Gate getroffen, 2× consec fail.

**Eskalations-Anbindung:** Die 5 Gates bleiben in issue-resolver. Trifft ein
Subagent ein Gate, bricht er NICHT den ganzen Run ab, sondern meldet `checkpoint
--outcome blocked --gate-reason <typ> --gate-question "<A/B/C>"`. Parallele Issues
laufen weiter; nur das geblockte eskaliert. `next` propagiert es als `ESCALATE` →
Skill stellt EINE A/B/C-Frage.

### C) Planning-Quality-Gate `forgecrate-plan-gate` (base)

Pfad `base/skills/forgecrate-plan-gate/SKILL.md`. Verpflichtender Zwischenschritt
**3b** im issue-resolver, zwischen writing-plans (Schritt 3) und Worktree (Schritt 4).

**6 Achsen, alle PASS (sonst Replan, max 1 Runde, dann Gate 1 „mehrdeutig"):**
Akzeptanzkriterien (≥1 prüfbar pro Anforderung), Test-Plan (≥1 Regressionstest
namentlich + Edge-Cases), betroffene Dateien (vollständig, je Änderungsart),
Risiko/Rollback (≥1 Risiko + konkreter Rollback), Definition-of-Done (abhakbar, je
Evidenzquelle), Scope-Sanity (≤1 PR / ≤3 Tage).

**Token-effizient:** Prüfung als Reviewer-Subagent (sonnet), nicht im Hauptkontext.
Plan-Kommentar erhält Block `📐 Plan-Gate: PASS (6/6)`.

### D) Token-Effizienz

1. **Aggressive Subagent-Delegation als Default** — Explore/Plan/Resolve nie im
   Hauptkontext. Resolve-Kontext wird nach PR verworfen → Langlauf-Kontext wächst
   nur ~1 Zeile History/Issue. Größter Hebel.
2. **Hartes Fan-out-Budget** via `MaxIterations`/`MaxParallel`.
3. **Checkpoint-Cadence** — Pflicht-`handoff` nach jeder Iteration; nach
   Kompaktierung Rehydration aus `memory_bank_read` + `autopilot status`.
4. **Was wohin:** memory-bank = RunState-Spiegel + Fokus; memory(graph) = zeitlose
   Entscheidungen; Hauptkontext = nur aktuelle Iteration; Issue-Kommentare = Audit-Trail.

### E) Packaging: Erweiterung base + github (KEIN neuer Flavor)

Der Runner ist ein Feature des github-Flavors (braucht roadmap-triage +
issue-resolver). Plan-Gate ist flavor-neutral → base. `check-readme-coverage`
iteriert nur über `ls flavors/` → kein neuer Flavor = automatisch erfüllt.

## Betroffene Dateien

`internal/autopilot/state.go`(+test), `cmd/forgecrate/autopilot.go`(+test),
`cmd/forgecrate/main.go`, `base/skills/forgecrate-plan-gate/SKILL.md`,
`flavors/github/skills/forgecrate-autonomous-runner/SKILL.md`(+`references/loop.md`),
`flavors/github/skills/forgecrate-issue-resolver/SKILL.md`, `base/CLAUDE.md`,
`flavors/github/CLAUDE.md`, `.gitignore`, `README.md`, `CHANGELOG.md`.

## Epic-Zerlegung (Sub-Issues, je ≤1 PR/≤3 Tage)

WSJF = (value+time+risk)/size, Fibonacci.

| # | Sub-Issue | WSJF | Prio | Dep |
|---|---|---|---|---|
| 1 | `internal/autopilot` State-Paket + Schema | (8+3+8)/3 = 6.3 | critical | — |
| 2 | `forgecrate autopilot` CLI (init/next/status/checkpoint/stop) | (8+3+5)/5 = 3.2 | high | #1 |
| 3 | `forgecrate-plan-gate` Skill (base) | (8+2+8)/2 = 9.0 | critical | — |
| 4 | issue-resolver um plan-gate erweitern | (5+3+3)/1 = 11.0 | critical | #3 |
| 5 | `forgecrate-autonomous-runner` Skill (github) | (13+3+5)/5 = 4.2 | critical | #2 |
| 6 | Token-Cadence + memory-bank-Integration | (5+2+8)/2 = 7.5 | critical | #5 |
| 7 | Self-Continuation via /loop + send_later | (5+3+3)/2 = 5.5 | critical | #5 |
| 8 | Docs + Packaging | (3+2+2)/2 = 3.5 | high | #2/#5 |

**Reihenfolge:** #1 → #2 → (#3 ∥ #4) → #5 → (#6 ∥ #7) → #8.

## Verifikation

- `make test` (`internal/autopilot` + `cmd/forgecrate/autopilot`), `make quality`.
  `next`-Baum als Tabellen-Tests (alle 4 ACTION-Zweige) wie `hook_test.go`.
- CLI manuell: `init --max-iterations 3` → `next` (RESOLVE_NEXT) → `checkpoint
  --outcome shipped` → `status` → bis `STOP`. Gate-Pfad: `checkpoint --outcome
  blocked --gate-reason ambiguous` → `next` = ESCALATE.
- `make check-readme-coverage` grün (kein neuer Flavor).
