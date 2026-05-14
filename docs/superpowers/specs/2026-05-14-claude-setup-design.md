# Claude Setup — Design Spec

**Datum:** 2026-05-14  
**Status:** Genehmigt

## Ziel

Ein reproduzierbares, per Go-Binary verwaltetes Claude-Setup, das alle relevanten Claude-Dateien (CLAUDE.md, AGENTS.md, settings.json, Skills, Hooks) in ein Ziel-Repo deployt. Das Setup ist pro Repo konfigurierbar (Profile + Flavors + lokale Overrides) und erzwingt die Einhaltung definierter Workflows via Hooks und Pflicht-Skills.

## Prinzipien

- Kein manuelles Installationsskript — ein globales Binary reicht
- Keine globale Config für mehrere Repos — jedes Repo hat seine eigene `.claude-setup.yaml`
- Hooks sind schlank und gezielt, kein Payload-Overload
- Overrides werden nie überschrieben

---

## Abschnitt 1: Repo-Struktur

### Source-Repo (`claude-setup` auf GitHub)

```
claude-setup/
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
└── cmd/claude-setup/            # Go-Binary Source
    └── main.go
```

### Ziel-Repo (nach `init`)

```
mein-projekt/
├── .claude-setup.yaml           # Ankerdatei
├── CLAUDE.md                    # Composited (base + profile + flavors)
├── AGENTS.md
└── .claude/
    ├── settings.json            # Deep-merged
    ├── commands/                # Alle Skills
    └── overrides/               # Layer 3 — lokal, nie überschrieben
        ├── CLAUDE.md.override
        └── settings.override.json
```

### Ankerdatei `.claude-setup.yaml`

```yaml
version: "1.0"
source: "github.com/markus/claude-setup"
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
go install github.com/markus/claude-setup/cmd/claude-setup@latest
```

### Befehle

```bash
claude-setup init --profile backend --flavors tdd,strict-review
claude-setup update
claude-setup update --profile fullstack
```

### `init`-Ablauf

1. Liest `.claude-setup.yaml` falls vorhanden (idempotent)
2. Holt Source-Stand von GitHub (API oder `git clone --depth=1`)
3. Kompositioniert Layer: `base` → `profile` → `flavors`
4. Schreibt Dateien ins Ziel-Repo
5. Schreibt `.claude-setup.yaml`
6. Berührt `overrides/` nicht

### `update`-Ablauf

1. Liest `.claude-setup.yaml`
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

## Offene Entscheidungen

- GitHub-Zugriff: öffentliche API (kein Token nötig für public repos) oder `git clone --depth=1` — beides valide, API bevorzugt für kleine Payloads
- Merge-Marker-Format für CLAUDE.md: `<!-- GENERATED -->` / `<!-- CUSTOM -->` muss im Base-Template klar definiert sein
- Versionierung: `ref: main` für latest, Tags für stabile Releases (empfohlen für Prod-Repos)
