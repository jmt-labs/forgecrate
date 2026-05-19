# Design: Slash-Commands für forgecrate-Skills

## Ziel

Alle Skills, die `forgecrate run` installiert, sollen als Slash-Commands im Format
`/forgecrate-<name>` aufrufbar sein. Das ermöglicht schnellen Zugriff ohne den
genauen Skill-Namen kennen zu müssen.

## Hintergrund

Claude Code lädt `.md`-Dateien aus `~/.claude/commands/` als Slash-Commands.
`forgecrate` hat in `compose.go` bereits `composeSkills`, das Dateien aus
`base/.claude/commands/`, `profiles/<p>/.claude/commands/` und
`flavors/<f>/.claude/commands/` nach `<destDir>/.claude/commands/` merged.

Die Verzeichnisse `base/.claude/commands/`, `profiles/*/.claude/commands/` und `flavors/*/.claude/commands/` existieren noch nicht.

## Design

### Kein Umbau — nur neue Dateien

Skills bleiben in ihren bestehenden Verzeichnissen (`base/skills/release/`,
`flavors/tdd/skills/test-coverage/`, etc.). Es werden nur Command-Wrapper-Dateien
hinzugefügt.

### Command-Dateien nach Layer

Commands liegen im selben Layer wie ihr Skill — base-Skills in `base/.claude/commands/`,
Profil-Skills in `profiles/<p>/.claude/commands/`, Flavor-Skills in
`flavors/<f>/.claude/commands/`. So werden Commands nur deployt, wenn der zugehörige
Skill auch tatsächlich installiert wird.

| Command-Datei | Slash-Command | Ruft auf |
|---|---|---|
| `forgecrate-advisor.md` | `/forgecrate-advisor` | `Skill("forgecrate-advisor")` |
| `forgecrate-release.md` | `/forgecrate-release` | `Skill("release")` |
| `forgecrate-repo-health.md` | `/forgecrate-repo-health` | `Skill("repo-health")` |
| `forgecrate-repo-onboarding.md` | `/forgecrate-repo-onboarding` | `Skill("repo-onboarding")` |
| `forgecrate-db-migration.md` | `/forgecrate-db-migration` | `Skill("db-migration")` |
| `forgecrate-test-coverage.md` | `/forgecrate-test-coverage` | `Skill("test-coverage")` |
| `forgecrate-pr-checklist.md` | `/forgecrate-pr-checklist` | `Skill("pr-checklist")` |
| `forgecrate-accessibility-audit.md` | `/forgecrate-accessibility-audit` | `Skill("accessibility-audit")` |
| `forgecrate-github-release.md` | `/forgecrate-github-release` | `Skill("github-release")` |

### Dateiformat

Jede Command-Datei ist minimal:

```markdown
---
description: <Kurzbeschreibung aus dem SKILL.md>
---

Use the Skill tool to invoke the "<skill-name>" skill.
```

### Deploy-Pfad

`forgecrate run` → `compose.Run()` → `composeSkills()` → merged alle
`*/.claude/commands/*.md`-Layer nach `~/.claude/commands/` → Slash-Commands aktiv.

Kein Umbau an `deploy.go`, `compose.go` oder `extensions.yaml` nötig.

### Zukünftiger Plugin-Schritt (nicht Teil dieses Features)

Später wird das Plugin-System genutzt, um den Namespace `forgecrate:<name>`
(mit Doppelpunkt) zu ermöglichen. Die Command-Dateien können dann 1:1 übernommen werden.

## Nicht in scope

- Umbenennung bestehender Skill-Verzeichnisse
- Änderungen an `deploy.go` oder `compose.go`
- Plugin-Registrierung
