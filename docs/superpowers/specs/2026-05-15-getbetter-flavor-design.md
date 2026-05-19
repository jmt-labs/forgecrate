# Design: GETBETTER-Flavor

## Ziel

Ein opt-in Flavor `getbetter`, der Claude nach jeder Session dazu bringt,
Erkenntnisse zu reflektieren und persistent zu speichern — damit die nächste
Session nicht bei Null beginnt.

## Hintergrund

Claude Code hat keinen Session-End-Hook. Die verfügbaren Hook-Events
(`UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Stop`) erlauben kein
verlässliches Session-Ende zu detektieren. Der GETBETTER-Flavor löst das
durch einen expliziten Slash-Command statt eines automatischen Hooks.

## Design

### Flavor-Struktur

```
flavors/getbetter/
├── CLAUDE.md
├── skills/getbetter/SKILL.md
└── .claude/commands/forgecrate-getbetter.md
```

Kein Go-Code. Kein `extensions.yaml`. Nur Markdown-Dateien.

### CLAUDE.md-Abschnitt

`flavors/getbetter/CLAUDE.md` wird in den composed CLAUDE.md eingebettet:

```markdown
## GETBETTER-Flavor

Falls `.claude/GETBETTER.md` existiert, MUSS sie vor allem anderen gelesen werden.
```

Die Instruktion ist bewusst imperativ (`MUSS`): Claude soll die Datei nicht
optional lesen, sondern zwingend — sofern sie vorhanden ist.

### Skill: `getbetter`

`flavors/getbetter/skills/getbetter/SKILL.md` führt Claude durch drei Schritte:

1. **Laden** — liest `.claude/GETBETTER.md` (falls vorhanden), sonst leere Basis
2. **Reflektieren** — analysiert die aktuelle Session und extrahiert:
   - Architekturentscheidungen (Was wurde entschieden und warum?)
   - Anti-Patterns (Was lief schief? Was hätte früher erkannt werden sollen?)
   - Was gut funktioniert hat (Ansätze, die sich bewährt haben)
   Claude formuliert frei innerhalb dieser Kategorien — kein starres Template.
3. **Synthetisieren** — führt bestehende und neue Erkenntnisse zusammen:
   - Bestehende Erkenntnisse bleiben erhalten
   - Neue ergänzen oder ersetzen überschneidende Punkte
   - Redundantes wird verdichtet
   - Schreibt das Ergebnis nach `.claude/GETBETTER.md`

### Slash-Command

`flavors/getbetter/.claude/commands/forgecrate-getbetter.md`:

```markdown
---
description: Aktuelle Session reflektieren und GETBETTER.md aktualisieren
---

Use the Skill tool to invoke the "getbetter" skill.
```

### GETBETTER.md Format

Kein starres Schema — Claude entscheidet Inhalt und Formulierung. Grobe Struktur:

```markdown
# GETBETTER

_Letzte Aktualisierung: 2026-05-15_

## Entscheidungen
...

## Anti-Patterns
...

## Was funktioniert
...
```

Die Datei liegt unter `.claude/GETBETTER.md` und ist versioniert (Teil des Repos).

## Nicht in Scope

- Automatischer `Stop`-Hook
- `/getbetter-load` Skill (Start-Seite ist CLAUDE.md-Instruktion)
- Session-Timestamp-Tracking pro Eintrag
- Lösch- oder Archivierungsmechanismus für alte Einträge

## Testbarkeit

E2E-Test analog zu `TestDeployIncludesFlavorSkill`:
- Deploy mit `Flavors: []string{"getbetter"}`
- Prüft: `flavors/getbetter/skills/getbetter/SKILL.md` wird nach
  `.claude/skills/getbetter/SKILL.md` gedeployt
- Prüft: `.claude/commands/forgecrate-getbetter.md` existiert
