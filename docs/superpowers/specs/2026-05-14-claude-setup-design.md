# forgecrate — Design Spec

**Datum:** 2026-05-14  
**Status:** Genehmigt

## Ziel

Ein reproduzierbares, per Go-Binary verwaltetes forgecrate, das alle relevanten Claude-Dateien (CLAUDE.md, AGENTS.md, settings.json, Skills, Hooks) in ein Ziel-Repo deployt. Das Setup ist pro Repo konfigurierbar (Profile + Flavors + lokale Overrides) und erzwingt die Einhaltung definierter Workflows via Hooks und Pflicht-Skills.

## Prinzipien

- Kein manuelles Installationsskript — ein globales Binary reicht
- Keine globale Config für mehrere Repos — jedes Repo hat seine eigene `.forgecrate.yaml`
- Hooks sind schlank und gezielt, kein Payload-Overload
- Overrides werden nie überschrieben

---

## Abschnitt 1: Repo-Struktur

### Source-Repo (`forgecrate` auf GitHub)

```
forgecrate/
├── base/                        # Layer 1 — immer deployt
│   ├── CLAUDE.md
│   ├── AGENTS.md
│   ├── .claude/
│   │   ├── settings.json
│   │   └── commands/            # Base-Skills
│   └── hooks/
│       ├── prompt-submit.sh
│       └── pre-tool.sh
├── profiles/                    # Layer 2 — eines wählbar
│   ├── backend/
│   │   ├── CLAUDE.md
│   │   └── .claude/commands/
│   ├── frontend/
│   └── fullstack/
├── flavors/                     # Layer 2b — mehrere kombinierbar
│   ├── tdd/
│   ├── strict-review/
│   └── minimal/
└── cmd/forgecrate/            # Go-Binary Source
    └── main.go
```

### Ziel-Repo (nach `init`)

```
mein-projekt/
├── .forgecrate.yaml           # Ankerdatei
├── CLAUDE.md                    # Composited (base + profile + flavors)
├── AGENTS.md
└── .claude/
    ├── settings.json            # Deep-merged
    ├── commands/                # Alle Skills
    └── overrides/               # Layer 3 — lokal, nie überschrieben
        ├── CLAUDE.md.override
        └── settings.override.json
```

### Ankerdatei `.forgecrate.yaml`

```yaml
version: "1.0"
source: "github.com/jmt-labs/forgecrate"
ref: "main"
profile: backend
flavors:
  - tdd
  - strict-review
```

---

## Abschnitt 2: Go Binary

### Installation

```bash
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest
```

### Befehle

```bash
forgecrate init --profile backend --flavors tdd,strict-review
forgecrate update
forgecrate update --profile fullstack
```

### `init`-Ablauf

1. Liest `.forgecrate.yaml` falls vorhanden (idempotent)
2. Holt Source-Stand von GitHub (API oder `git clone --depth=1`)
3. Kompositioniert Layer: `base` → `profile` → `flavors`
4. Schreibt Dateien ins Ziel-Repo
5. Schreibt `.forgecrate.yaml`
6. Berührt `overrides/` nicht

### `update`-Ablauf

1. Liest `.forgecrate.yaml`
2. Holt neuen Stand von GitHub
3. Rekompositioniert alle generierten Dateien
4. `overrides/` bleibt unangetastet
5. Gibt Diff der Änderungen aus

### Layer-Kompositions-Strategie

| Datei | Strategie |
|---|---|
| `CLAUDE.md` / `AGENTS.md` | Sections-Merge via `<!-- GENERATED -->` / `<!-- CUSTOM -->` Marker |
| `settings.json` | Deep JSON Merge — Override-Keys gewinnen bei Konflikten |
| `.claude/commands/` | Additiv — alle Skills werden kopiert, Overrides überschreiben gleichnamige Dateien |

---

## Abschnitt 3: Enforcement

### Hooks in `settings.json`

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [{
          "type": "command",
          "command": "bash .claude/hooks/prompt-submit.sh"
        }]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash|Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "bash .claude/hooks/pre-tool.sh"
        }]
      }
    ]
  }
}
```

### Hook-Verhalten

**`prompt-submit.sh`** — bei jeder User-Nachricht:
- Gibt kompakte Erinnerung aus: aktives Profile, liste der Pflicht-Skills
- Kein großer Payload — wenige Zeilen, vollständig cached nach erster Ausführung

**`pre-tool.sh`** — vor jedem `Bash`/`Edit`/`Write`-Call:
- Prüft ob ein relevanter Pflicht-Skill für die aktuelle Aktion hätte aufgerufen werden müssen
- Gibt Warnung auf stderr aus wenn nicht — Claude sieht diese als Kontext

### Pflicht-Skills (in compositioniertem CLAUDE.md)

```markdown
## Pflicht-Skills

| Situation | Skill | Verhalten |
|---|---|---|
| Neues Feature / Bug-Fix | `superpowers:brainstorming` | MUSS vor Code aufgerufen werden |
| Implementierung | `superpowers:test-driven-development` | MUSS vor Code aufgerufen werden |
| Vor Commit/PR | `superpowers:verification-before-completion` | MUSS ausgeführt werden |
| Debug | `superpowers:systematic-debugging` | MUSS vor Fix aufgerufen werden |
```

### Enforcement-Flow

```
User schreibt Prompt
  → UserPromptSubmit-Hook: gibt Pflicht-Skill-Erinnerung aus
  → Claude erkennt Trigger aus Prompt + CLAUDE.md-Anweisung
  → Claude ruft Skill auf
  → PreToolUse-Hook: prüft vor Edit/Write ob Skill gelaufen
  → Tool wird ausgeführt
```

---

## Abschnitt 4: Testing

### Unit Tests (Go)

Alle internen Funktionen des Binary werden unit-getestet. Schwerpunkte:

| Paket | Was getestet wird |
|---|---|
| `compose` | Layer-Merge-Logik für CLAUDE.md (Marker-Erkennung, Section-Merge) |
| `compose` | Deep JSON Merge für settings.json (Konfliktauflösung, Override-Precedence) |
| `compose` | Skills-Komposition (Additiv, Override-Erkennung) |
| `github` | GitHub-API-Client (gemockt) — Tarball-Download, Ref-Auflösung |
| `config` | `.forgecrate.yaml` Lesen/Schreiben, Validierung |

```bash
go test ./...
```

### E2E Tests

Testen den vollständigen `init`- und `update`-Zyklus gegen ein echtes (oder gemocktes) GitHub-Repo in einem temporären Verzeichnis.

Szenarien:
- `init` auf leerem Verzeichnis → korrekte Dateistruktur
- `init` ist idempotent → zweiter Aufruf ändert nichts
- `update` mit neuer Source-Version → Basis-Dateien aktualisiert, Overrides erhalten
- `update --profile` wechselt Profile → neue Skills landen, alte Base-Skills bleiben
- Overrides werden bei keinem Befehl überschrieben

```bash
go test ./e2e/...
```

---

## Abschnitt 5: Technische Dokumentation

Liegt unter `docs/` im Source-Repo. Alle Diagramme als SVG (produziert aus Mermaid oder draw.io-Quellen, eingecheckt als SVG).

| Dokument | Inhalt |
|---|---|
| `docs/architecture.md` | Komponentendiagramm: Binary · Source-Repo · Ziel-Repo · GitHub-API |
| `docs/flows.md` | Ablaufdiagramme: `init`-Flow, `update`-Flow, Layer-Komposition, Enforcement-Flow |
| `docs/layer-system.md` | Detaillierte Erklärung des 3-Layer-Systems mit Beispielen |
| `docs/hooks.md` | Hook-Referenz: Zeitpunkt, Payload, Beispiel-Output |
| `docs/profiles-flavors.md` | Alle verfügbaren Profile und Flavors mit Beschreibung |

Diagramme werden als SVG unter `assets/` abgelegt und in die Docs eingebettet:

```markdown
![Init Flow](../assets/flow-init.svg)
```

---

## Abschnitt 6: Endbenutzerdokumentation (README.md)

### Stil: forgedeck-inspiriert

- SVG-Banner oben zentriert (`assets/banner.svg`) mit Tagline
- Deutsch durchgehend, kein Sprachmix
- Keine CI-Badges oder Shields
- Tabellen statt Aufzählungen für Navigation und Komponenten
- Horizontale Trennlinien (`---`) zwischen Hauptabschnitten
- Quick Start zuerst, Details in verlinkten Docs

### README-Struktur

```
<div align="center">
  <img src="assets/banner.svg" alt="forgecrate — Reproducible forgecrate" width="100%">
</div>

# forgecrate

Kurzbeschreibung (1 Satz).

Stack: Go · GitHub API · Layer-System · Hooks

---

## Quick Start

Voraussetzungen: Go 1.22+, GitHub-Zugriff.

```sh
go install github.com/jmt-labs/forgecrate/cmd/forgecrate@latest
forgecrate init --profile backend --flavors tdd
```

---

## Dokumentation

| Thema | Dokument |
|---|---|
| Architektur | docs/architecture.md |
| Abläufe | docs/flows.md |
| Profile & Flavors | docs/profiles-flavors.md |
| Hooks | docs/hooks.md |
| Entwicklung | docs/development.md |

---

## Komponenten

| Pfad | Zweck |
|---|---|
| base/ | Basis-Layer — immer deployt |
| profiles/ | Profil-Layer |
| flavors/ | Flavor-Layer |
| cmd/forgecrate/ | Go-Binary |
```

---

## Offene Entscheidungen

- GitHub-Zugriff: öffentliche API (kein Token nötig für public repos) oder `git clone --depth=1` — beides valide, API bevorzugt für kleine Payloads
- Merge-Marker-Format für CLAUDE.md: `<!-- GENERATED -->` / `<!-- CUSTOM -->` muss im Base-Template klar definiert sein
- Versionierung: `ref: main` für latest, Tags für stabile Releases (empfohlen für Prod-Repos)
